/* clparse package contains utils for command line protocols creation */

package clparse

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ParseResult []string

type ArgNumError int

func (e ArgNumError) Error() string {
	return fmt.Sprintf("Wrong number of arguments, must be %v", int(e))
}

func (res ParseResult) RawArg(buffer *bytes.Buffer) ParseResult {
	if buffer.Len() > 0 {
		res = append(res, buffer.String())
		buffer.Reset()
	}
	return res
}

func (res ParseResult) QuotedArg(buffer *bytes.Buffer) (ParseResult, error) {
	unq, err := strconv.Unquote(buffer.String())
	if err != nil {
		return res, err
	}
	res = append(res, unq)
	buffer.Reset()
	return res, nil
}

func SplitCommand(input string) (command, argpart string) {
	splited := strings.SplitN(input, " ", 2)
	if len(splited) == 1 {
		return splited[0], ""
	}
	return splited[0], splited[1]
}

func ParseArgs(argString string, argNum int) ([]string, error) {

	var buffer = new(bytes.Buffer)
	var isq bool // is we in quote
	var err error = nil

	res := make(ParseResult, 0)

	if argNum == 0 && argString != "" {
		return res, ArgNumError(argNum)
	}

	for i := 0; i < len(argString); i++ {
		if len(res) >= argNum {
			return res, ArgNumError(argNum)
		}
		ch := rune(argString[i])
		if ch == '"' {
			buffer.WriteRune('"')
			if isq {
				if len(argString) > i+1 && !unicode.IsSpace(rune(argString[i+1])) {
					return res, fmt.Errorf("Closing quote must follow by space character or nothing at all")
				}
				res, err = res.QuotedArg(buffer)
				if err != nil {
					return res, err
				}
			}
			isq = !isq
			continue
		}
		if isq {
			buffer.WriteRune(ch)
			if ch == '\\' {
				i++
				ch := rune(argString[i])
				buffer.WriteRune(ch)
			}
			continue
		}
		if ch == ' ' {
			res = res.RawArg(buffer)
			continue
		}
		buffer.WriteRune(ch)
	}
	if isq {
		return res, fmt.Errorf("Unbalanced quotes")
	}
	res = res.RawArg(buffer)

	if len(res) != argNum {
		return res, ArgNumError(argNum)
	}

	return res, nil
}
