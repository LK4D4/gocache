package main

import (
	"fmt"
	dict "godict"
)

var storage = dict.New()

const okFormat = "OK %v"
const errFormat = "ERR %v"

type commandErr struct {
	err string
}

func (e commandErr) Error() string {
	return fmt.Sprintf(errFormat, e.err)
}

type commandFunc func(...string) string

type commandOpt struct {
	argNumber int
	f         commandFunc
}

var commandsMap = map[string]commandOpt{
	"set":    {2, set},
	"get":    {1, get},
	"delete": {1, delete},
}

func set(args ...string) string {
	err := storage.Set(args[0], args[1])
	if err != nil {
		return fmt.Sprintf(errFormat, err)
	}
	return "OK"
}

func get(args ...string) string {
	slot, err := storage.Get(args[0])
	if err != nil {
		return fmt.Sprintf(errFormat, err)
	}
	return fmt.Sprintf(okFormat, slot.Value())
}

func delete(args ...string) string {
	err := storage.Delete(args[0])
	if err != nil {
		return fmt.Sprintf(errFormat, err)
	}
	return "OK"
}
