/* clparse package contains utils for command line protocols creation */

package clparse

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const defaultCap = 1024

type ArgNumError int

func (e ArgNumError) Error() string {
	return fmt.Sprintf("Wrong number of arguments, must be %v", int(e))
}

func SplitCommand(input string) (command, argpart string) {
	index := strings.IndexRune(input, ' ')
	if index == -1 {
		return input, ""
	}
	lindex := index + 1
	for ; input[lindex] == ' ' && lindex < len(input); lindex++ {
	}
	return input[0:index], input[lindex:]
}

func ParseArgs(argString string, argNum int) ([]string, error) {

	var isq bool // is we in quote

	res := make([]string, 0, argNum)
	buf := make([]byte, 0, defaultCap)

	if argNum == 0 && argString != "" {
		return res, ArgNumError(argNum)
	}

	for i := 0; i < len(argString); i++ {
		if len(res) == argNum {
			return res, ArgNumError(argNum)
		}
		ch := argString[i]
		if ch == '"' {
			buf = append(buf, '"')
			if isq {
				if len(argString) > i+1 && !unicode.IsSpace(rune(argString[i+1])) {
					return res, fmt.Errorf("Closing quote must follow by space character or nothing at all")
				}
				unq, err := strconv.Unquote(string(buf))
				if err != nil {
					return res, err
				}
				res = append(res, unq)
				buf = buf[:0]
			}
			isq = !isq
			continue
		}
		if isq {
			buf = append(buf, ch)
			if ch == '\\' {
				i++
				buf = append(buf, argString[i])
			}
			continue
		}
		if ch == ' ' {
			if len(buf) > 0 {
				res = append(res, string(buf))
				buf = buf[:0]
			}
			continue
		}
		buf = append(buf, ch)
	}
	if isq {
		return res, fmt.Errorf("Unbalanced quotes")
	}
	if len(buf) > 0 {
		res = append(res, string(buf))
	}

	if len(res) != argNum {
		return res, ArgNumError(argNum)
	}

	return res, nil
}
