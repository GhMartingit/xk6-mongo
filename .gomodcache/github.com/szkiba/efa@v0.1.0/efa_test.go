package efa_test

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/szkiba/efa"
)

func TestGetEfa(t *testing.T) {
	t.Parallel()

	if efa.GetEfa() == nil {
		t.Error("GetEfa() returns nil")
	}
}

func testLookup(t *testing.T, envstr string) efa.LookupFunc {
	t.Helper()

	envvars := strings.Split(envstr, ";")
	env := make(map[string]string, len(envvars))

	const nvlen = 2

	for _, envvar := range envvars {
		parts := strings.SplitN(envvar, "=", nvlen)

		env[parts[0]] = parts[1]
	}

	return func(name string) (string, bool) {
		value, found := env[name]

		return value, found
	}
}

func testEnv(t *testing.T, envstr string) {
	t.Helper()

	envvars := strings.Split(envstr, ";")

	const nvlen = 2

	for _, envvar := range envvars {
		parts := strings.SplitN(envvar, "=", nvlen)

		t.Setenv(parts[0], parts[1])
	}
}

type testDataType struct {
	name    string
	prefix  string
	env     string
	flag    string
	value   int
	want    string
	wantErr bool
}

func testData(t *testing.T) []testDataType {
	t.Helper()

	return []testDataType{
		{name: "simple", prefix: "cli", env: "CLI_FOO=3", flag: "foo", want: "3"},
		{name: "missing", prefix: "cli", env: "CLI_FOO=3", flag: "bar", want: "0"},
		{name: "error", prefix: "cli", env: "CLI_FOO=bad", flag: "foo", want: "0", wantErr: true},
		{name: "no_prefix", prefix: "", env: "FOO=5", flag: "foo", want: "5"},
	}
}

func TestEfa_Bind(t *testing.T) {
	t.Parallel()

	for _, tt := range testData(t) {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lookup := testLookup(t, tt.env)
			flags := flag.NewFlagSet("test", flag.ContinueOnError)

			flags.Int(tt.flag, tt.value, "usage")

			e := efa.New(flags, tt.prefix, lookup)

			if err := e.Bind(tt.flag); (err != nil) != tt.wantErr {
				t.Errorf("Efa.Bind() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := flags.Lookup(tt.flag).Value.String(); got != tt.want {
				t.Errorf(`want "%v", got "%v"`, tt.want, got)
			}
		})
	}
}

//nolint:paralleltest
func TestEfa_Bind_OS_Lookup(t *testing.T) {
	for _, tt := range testData(t) {
		t.Run(tt.name, func(t *testing.T) {
			testEnv(t, tt.env)

			flags := flag.NewFlagSet("test", flag.ContinueOnError)

			flags.Int(tt.flag, tt.value, "usage")

			e := efa.New(flags, tt.prefix, nil)

			if err := e.Bind(tt.flag); (err != nil) != tt.wantErr {
				t.Errorf("Efa.Bind() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := flags.Lookup(tt.flag).Value.String(); got != tt.want {
				t.Errorf(`want "%v", got "%v"`, tt.want, got)
			}
		})
	}
}

func TestEfa_Bind_Multiple(t *testing.T) {
	t.Parallel()

	flags := flag.NewFlagSet("test", flag.ContinueOnError)

	foo := flags.Int("foo", -1, "foo usage")
	bar := flags.Int("bar", -1, "bar usage")

	e := efa.New(flags, "cli", testLookup(t, "CLI_FOO=3;CLI_BAR=4"))

	if err := e.Bind("bar", "foo"); err != nil {
		t.Errorf("Bind returns with error %v", err)
	}

	if *bar != 4 {
		t.Errorf("bar got=%d, want=4", *bar)
	}

	if *foo != 3 {
		t.Errorf("foo got=%d, want=3", *foo)
	}
}

func TestEfa_BindTo_Multiple(t *testing.T) {
	t.Parallel()

	flags := flag.NewFlagSet("test", flag.ContinueOnError)

	foo := flags.Int("foo", -1, "foo usage")
	bar := flags.Int("bar", -1, "bar usage")

	e := efa.New(flags, "cli", testLookup(t, "DUMMY_FOO=3;DUMMY_BAR=4"))

	if err := e.BindTo("bar", "DUMMY_BAR", "foo", "DUMMY_FOO"); err != nil {
		t.Errorf("Bind returns with error %v", err)
	}

	if *bar != 4 {
		t.Errorf("bar got=%d, want=4", *bar)
	}

	if *foo != 3 {
		t.Errorf("foo got=%d, want=3", *foo)
	}
}

func TestEfa_BindTo_With_Error(t *testing.T) {
	t.Parallel()

	flags := flag.NewFlagSet("test", flag.ContinueOnError)

	e := efa.New(flags, "", testLookup(t, "FOO=3;BAR=4"))

	if err := e.BindTo("bar", "DUMMY_BAR", "foo"); err == nil {
		t.Errorf("want=%v, got=nil", efa.ErrInvalidArgument)
	}
}

func TestEfa_Bind_Annotatable(t *testing.T) {
	t.Parallel()

	flags := &aFlagSet{t: t}

	e := efa.New(flags, "cli", testLookup(t, "CLI_FOO=3"))

	if err := e.Bind("foo"); err != nil {
		t.Errorf("Bind() error = %v", err)
	}

	if flags.annotations == nil {
		t.Error("annotations is nil")
	}

	if a, has := flags.annotations[efa.Annotation]; !has || a[0] != "CLI_FOO" {
		t.Error("invalid annotation")
	}

	if err := e.Bind("error"); err == nil {
		t.Error("want error")
	}
}

type aFlagSet struct {
	t           *testing.T
	annotations map[string][]string
}

func (a *aFlagSet) SetAnnotation(name, key string, values []string) error {
	a.t.Helper()

	if name == "error" {
		return os.ErrNotExist
	}

	a.annotations = map[string][]string{key: values}

	return nil
}

func (a *aFlagSet) Set(_, _ string) error {
	a.t.Helper()

	return nil
}
