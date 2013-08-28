/* Commands parsing helpers */
package commands

import (
	"clparse"
	"fmt"
	dict "godict"
)

var Storage = dict.New()

const okFormat = "OK %v"
const errFormat = "ERR %v"

func Set(args ...string) string {
	err := Storage.Set(args[0], args[1])
	if err != nil {
		return fmt.Sprintf(errFormat, err)
	}
	return "OK"
}

func Get(args ...string) string {
	slot, err := Storage.Get(args[0])
	if err != nil {
		return fmt.Sprintf(errFormat, err)
	}
	return fmt.Sprintf(okFormat, slot.Value())
}

func Delete(args ...string) string {
	err := Storage.Delete(args[0])
	if err != nil {
		return fmt.Sprintf(errFormat, err)
	}
	return "OK"
}

type CommandErr struct {
	err string
}

func (e CommandErr) Error() string {
	return fmt.Sprintf(errFormat, e.err)
}

type CommandFunc func(...string) string

type Command struct {
	ArgNumber int
	f         CommandFunc
}

var CommandsMap = map[string]Command{
	"set":    {2, Set},
	"get":    {1, Get},
	"delete": {1, Delete},
}

func ProcessTcpInput(input string) (string, error) {
	command, argString := clparse.SplitCommand(input)
	opts, ok := CommandsMap[command]
	if !ok {
		return "", CommandErr{fmt.Sprintf("Wrong command %s", command)}
	}
	args, err := clparse.ParseArgs(argString, opts.ArgNumber)
	if err != nil {
		return "", CommandErr{err.Error()}
	}
	return opts.f(args...), nil
}
