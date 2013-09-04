package clparse

import (
	"testing"
)

var argTable = []struct {
	input    string
	num      int
	expected []string
}{
	{`"a b" 1`, 2, []string{"a b", "1"}},
	{`"test \"super\" quote" "a b"`, 2,
		[]string{"test \"super\" quote", "a b"}},
	{`set "a b" 1`, 3, []string{"set", "a b", "1"}},
	{`set "a\nb" 1`, 3, []string{"set", "a\nb", "1"}},
	{`set "a\tb" 1`, 3, []string{"set", "a\tb", "1"}},
	{`set   testing::key "{\"key\": \"value\"}"`, 3,
		[]string{"set", "testing::key", "{\"key\": \"value\"}"}},
	{`set "a\\" 1`, 3, []string{"set", "a\\", "1"}},
	{`set "a" "x y "`, 3, []string{"set", "a", "x y "}},
	{`юникод "арг1" "x y "`, 3, []string{"юникод", "арг1", "x y "}},
}

func isEqual(input, expected []string) bool {
	if len(input) != len(expected) {
		return false
	}

	for i, v := range input {
		if v != expected[i] {
			return false
		}
	}
	return true
}

func Test_ParseArgs(t *testing.T) {
	for _, args := range argTable {
		result, err := ParseArgs(args.input, args.num)
		if err != nil {
			t.Errorf("Parse error: %v", err)
		}
		if !isEqual(result, args.expected) {
			t.Errorf("Result: %q, expected: %q, input was: %q",
				result, args.expected, args.input)
		}
	}
}

var argTableErr = []struct {
	input    string
	num      int
	expected string
}{
	{`"a b" 1`, 1, "Wrong number of arguments, must be 1"},
	{`"a b"`, 2, "Wrong number of arguments, must be 2"},
	{`a`, 2, "Wrong number of arguments, must be 2"},
	{`"a b 1`, 1, "Unbalanced quotes"},
	{`"a b"1`, 1, "Closing quote must follow by space character or nothing at all"},
}

func Test_ParseArgsErrors(t *testing.T) {
	for _, args := range argTableErr {
		result, err := ParseArgs(args.input, args.num)
		if err == nil {
			t.Errorf("Must be error, but got result: %v", result)
		}
		if err.Error() != args.expected {
			t.Errorf("Wrong error %s, must be %s", err, args.expected)
		}
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseArgs(`"strkeysad" "\"\,asdasd\"\""`, 2)
	}
}

var argForSplit = []struct {
	input           string
	expectedCommand string
	expectedArgs    string
}{
	{`set "a b" 1`, "set", `"a b" 1`},
	{`delete`, "delete", ``},
	{`delete   a b c`, "delete", `a b c`},
}

func TestSplit(t *testing.T) {
	for _, argSplit := range argForSplit {
		command, args := SplitCommand(argSplit.input)
		if command != argSplit.expectedCommand {
			t.Errorf("Command: %q, expected: %q, input was: %q",
				command, argSplit.expectedCommand, argSplit.input)
		}
		if args != argSplit.expectedArgs {
			t.Errorf("Arguments: %q, expected: %q, input was: %q",
				args, argSplit.expectedArgs, argSplit.input)
		}
	}
}
