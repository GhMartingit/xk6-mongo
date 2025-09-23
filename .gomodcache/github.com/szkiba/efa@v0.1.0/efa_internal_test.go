package efa

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestGetEfa(t *testing.T) {
	t.Parallel()

	efa := GetEfa()

	if efa.prefix != exeName() {
		t.Errorf("prefix = %v, want %v", efa.prefix, exeName())
	}

	if efa.lookup == nil {
		t.Error("lookup is nil")
	}

	if efa.set == nil {
		t.Error("set is nil")
	}
}

func Test_exeName(t *testing.T) {
	t.Parallel()

	name := exeName()

	if len(name) == 0 {
		t.Error("exeName() returns empty string")
	}

	want, err := os.Executable()
	if err != nil {
		want = os.Args[0]
	}

	want = filepath.Base(want)

	if name != want {
		t.Errorf("exeName() = %v, want %v", name, want)
	}
}

func Test_toUpperSnake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want string
	}{
		{name: "simple", arg: "foo", want: "FOO"},
		{name: "spaces", arg: "foo bar dummy", want: "FOO_BAR_DUMMY"},
		{name: "dash", arg: "foo-bar-dummy", want: "FOO_BAR_DUMMY"},
		{name: "mixed", arg: "foo-bar dummy", want: "FOO_BAR_DUMMY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := toUpperSnake(tt.arg); got != tt.want {
				t.Errorf("toUpperSnake() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	flagset := flag.NewFlagSet("foo", flag.ContinueOnError)

	lookup := func(string) (string, bool) { return "", false }

	tests := []struct {
		name   string
		flags  FlagSet
		prefix string
		lookup LookupFunc
	}{
		{name: "empty", flags: nil, prefix: "", lookup: nil},
		{name: "prefix", flags: nil, prefix: "foo", lookup: nil},
		{name: "flagset", flags: flagset, prefix: "", lookup: nil},
		{name: "lookup", flags: nil, prefix: "", lookup: lookup},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(tt.flags, tt.prefix, tt.lookup)

			if got.prefix != tt.prefix {
				t.Errorf("prefix = %v, want %v", got.prefix, tt.prefix)
			}

			if got.lookup == nil {
				t.Error("lookup is nil")
			}

			if got.set == nil {
				t.Error("set is nil")
			}
		})
	}
}
