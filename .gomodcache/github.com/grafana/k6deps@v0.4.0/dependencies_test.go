package k6deps_test

import (
	"testing"

	"github.com/grafana/k6deps"
	"github.com/stretchr/testify/require"
)

func testDeps(t *testing.T, k6 string) k6deps.Dependencies {
	t.Helper()

	deps := make(k6deps.Dependencies)

	dep, err := k6deps.NewDependency("foo", ">v0.1.0")
	require.NoError(t, err)
	deps[dep.Name] = dep

	dep, err = k6deps.NewDependency("bar", ">v0.2.0")
	require.NoError(t, err)
	deps[dep.Name] = dep

	if len(k6) == 0 {
		return deps
	}

	dep, err = k6deps.NewDependency("k6", k6)
	require.NoError(t, err)
	deps[dep.Name] = dep

	return deps
}

func Test_Dependencies_Merge(t *testing.T) {
	t.Parallel()

	to := testDeps(t, "")
	from := make(k6deps.Dependencies)

	dep, err := k6deps.NewDependency("dumb", ">v0.5.0")
	require.NoError(t, err)
	from[dep.Name] = dep

	dep, err = k6deps.NewDependency("dumber", ">v0.5.0")
	require.NoError(t, err)
	from[dep.Name] = dep

	err = to.Merge(from)
	require.NoError(t, err)
	require.Contains(t, to, "dumb")
	require.Contains(t, to, "dumber")

	err = to.Merge(from)
	require.NoError(t, err)

	dep, err = k6deps.NewDependency("dumb", ">v0.6.0")
	require.NoError(t, err)
	from[dep.Name] = dep

	err = to.Merge(from)
	require.Error(t, err)
}

func Test_Dependencies_MarshalText(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, "")

	text, err := deps.MarshalText()
	require.NoError(t, err)
	require.Equal(t, "bar>v0.2.0;foo>v0.1.0", string(text))

	deps = testDeps(t, ">0.49")

	text, err = deps.MarshalText()
	require.NoError(t, err)
	require.Equal(t, "k6>0.49;bar>v0.2.0;foo>v0.1.0", string(text))
}

func Test_Dependencies_String(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, ">0.49")

	text, err := deps.MarshalText()

	require.NoError(t, err)

	require.Equal(t, string(text), deps.String())
}

func Test_Dependencies_UnmarshalText(t *testing.T) {
	t.Parallel()

	deps := make(k6deps.Dependencies)

	err := deps.UnmarshalText([]byte("k6>0.49.0;bar>1.0;foo>2.0"))
	require.NoError(t, err)

	require.Equal(t, "k6>0.49.0", deps["k6"].String())
	require.Equal(t, "bar>1.0", deps["bar"].String())
	require.Equal(t, "foo>2.0", deps["foo"].String())

	err = deps.UnmarshalText([]byte("k6>0.49.0;bar>>1.0"))
	require.Error(t, err)

	err = deps.UnmarshalText([]byte("k6>0.49.0;bar>1.0;bar>2.0"))
	require.Error(t, err)
}

func Test_Dependencies_MarshalJSON(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, ">0.50")

	bin, err := deps.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `{"k6":">0.50","bar":">v0.2.0","foo":">v0.1.0"}`, string(bin))
}

func Test_Dependencies_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	deps := make(k6deps.Dependencies)

	err := deps.UnmarshalJSON([]byte(`{"k6":">0.50","bar":">v0.2.0","foo":">v0.1.0"}`))
	require.NoError(t, err)

	require.Equal(t, "k6>0.50", deps["k6"].String())
	require.Equal(t, "bar>v0.2.0", deps["bar"].String())
	require.Equal(t, "foo>v0.1.0", deps["foo"].String())

	err = deps.UnmarshalJSON([]byte(`{"k6":">0.50","bar":">>v0.2.0","foo":">>v0.1.0"}`))
	require.Error(t, err)
}

func Test_Dependencies_MarshalJS(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, ">0.50")

	bin, err := deps.MarshalJS()
	require.NoError(t, err)
	require.Equal(t, `"use k6>0.50";
"use k6 with bar>v0.2.0";
"use k6 with foo>v0.1.0";
`, string(bin))
}

func Test_Dependencies_UnmarshalJS(t *testing.T) {
	t.Parallel()

	deps := make(k6deps.Dependencies)

	err := deps.UnmarshalJS([]byte(`"use k6>0.50";
"use k6 with k6/x/bar>v0.2.0";
"use k6 with k6/x/foo>v0.1.0";
import "k6/x/dumber";
import hello from "k6/x/hello"
import bar from "k6/x/foo/bar";
//import baz from "k6/x/baz";
/* import baz from "k6/x/baz"; */
let dumb = require("k6/x/dumb");
`))
	require.NoError(t, err)

	require.Equal(t, "k6>0.50", deps["k6"].String())
	require.Equal(t, "k6/x/bar>v0.2.0", deps["k6/x/bar"].String())
	require.Equal(t, "k6/x/foo>v0.1.0", deps["k6/x/foo"].String())
	require.Equal(t, "k6/x/dumb*", deps["k6/x/dumb"].String())
	require.Equal(t, "k6/x/dumber*", deps["k6/x/dumber"].String())
	require.Equal(t, "k6/x/hello*", deps["k6/x/hello"].String())
	require.Equal(t, "k6/x/foo/bar*", deps["k6/x/foo/bar"].String())
	require.Nil(t, deps["k6/x/baz"]) // baz should had been ignored (is commented)

	err = deps.UnmarshalJS([]byte(`"use k6 with k6/x/foo>v0.1.0";
"use k6 with k6/x/dumb>v0.4.0";
`))
	require.NoError(t, err)
	require.Equal(t, "k6/x/dumb>v0.4.0", deps["k6/x/dumb"].String())

	err = deps.UnmarshalJS([]byte(`"use k6 >>1";
	`))
	require.Error(t, err)

	err = deps.UnmarshalJS([]byte(`"use k6 with k6/x/foo>>1";
	`))
	require.Error(t, err)

	err = deps.UnmarshalJS([]byte(`"use k6 with k6/x/foo>3.0";
"use k6 with k6/x/foo>v0.1.0";
`))
	require.Error(t, err)
}

func Test_Dependencies_UnmarshalJS_real_script(t *testing.T) {
	t.Parallel()

	deps := make(k6deps.Dependencies)

	err := deps.UnmarshalJS([]byte(`
import exec from 'k6/x/exec';
import faker from "k6/x/faker"
import "k6/x/sql"

export default function () {
  console.log(exec.command("date"));
}	
`))
	require.NoError(t, err)

	require.Len(t, deps, 3)
	require.Equal(t, "k6/x/exec*", deps["k6/x/exec"].String())
	require.Equal(t, "k6/x/faker*", deps["k6/x/faker"].String())
	require.Equal(t, "k6/x/sql*", deps["k6/x/sql"].String())
}

func Test_Dependencies_UnmarshalJS_k6_path(t *testing.T) {
	t.Parallel()

	deps := make(k6deps.Dependencies)

	err := deps.UnmarshalJS([]byte(`
import Counter from 'k6/metrics';

const count = new Conter("foo")

export default function () {
  count.add(1)
}	
`))
	require.NoError(t, err)

	require.Len(t, deps, 1)
	require.Equal(t, "k6*", deps["k6"].String())
}
