package atlas

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodePath(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("c", "6")
	n2 := n1.AddLink("a", "5")
	n3 := n2.AddLink("b", "7")

	require.Equal(t, map[string]string{
		"a": "5",
		"b": "7",
		"c": "6",
	}, n3.Path())
}

func TestNodeAddLink(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("a", "5")
	n2 := n1.AddLink("c", "6")
	n3 := n2.AddLink("b", "7")
	nempty := r.AddLink("", "")

	require.Equal(t, n3.Path(), map[string]string{
		"a": "5",
		"b": "7",
		"c": "6",
	})
	require.True(t, r != n1)
	require.True(t, r != n2)
	require.True(t, r != n3)
	require.True(t, n2 != n1)
	require.True(t, n3 != n1)
	require.True(t, n2 != n3)
	require.True(t, r == nempty)
	require.True(t, n2 == r.AddLink("c", "6").AddLink("a", "5"))
	require.True(t, n2 == r.AddLink("c", "6").AddLink("a", "5").AddLink("a", "5"))
}

func TestNodeValueByKey(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("c", "6").AddLink("a", "5")
	n2 := r.AddLink("b", "7")

	v, ok := n1.ValueByKey("c")
	require.True(t, ok)
	assert.Equal(t, "6", v)

	_, ok = n1.ValueByKey("b")
	require.False(t, ok)

	v, ok = n2.ValueByKey("b")
	require.True(t, ok)
	assert.Equal(t, "7", v)
}

func TestNodeDeleteKey(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("c", "6").AddLink("a", "5")
	n2 := n1.AddLink("b", "7").AddLink("d", "9")

	n3 := n1.DeleteKey("c")
	n4 := n2.DeleteKey("c")

	assert.Equal(t, r, r.DeleteKey("doesnotexist"))
	assert.Equal(t, n4, n4.DeleteKey("doesnotexist"))

	expn1 := map[string]string{"a": "5", "c": "6"}
	assert.Equal(t, expn1, n1.Path())

	expn2 := map[string]string{"a": "5", "c": "6", "b": "7", "d": "9"}
	assert.Equal(t, expn2, n2.Path())

	expn3 := map[string]string{"a": "5"}
	assert.Equal(t, expn3, n3.Path())

	expn4 := map[string]string{"a": "5", "b": "7", "d": "9"}
	assert.Equal(t, expn4, n4.Path())

	assert.Equal(t, r, n3.DeleteKey("a"))
}

func TestNodeLen(t *testing.T) {
	t.Parallel()

	r := New()
	n1 := r.AddLink("c", "6")
	n2 := n1.AddLink("b", "7").AddLink("d", "9")

	assert.Equal(t, r.Len(), 0)
	assert.Equal(t, n1.Len(), 1)
	assert.Equal(t, n2.Len(), 3)
}

func TestNodeContains(t *testing.T) {
	t.Parallel()

	r := New()

	n1 := r.AddLink("a", "5").
		AddLink("c", "6").
		AddLink("b", "7")

	n2 := r.AddLink("b", "7").AddLink("c", "6")
	n3 := r.AddLink("b", "7").AddLink("c", "4")
	n4 := r.AddLink("a", "5")

	n5 := r.AddLink("d", "9").
		AddLink("b", "7").
		AddLink("c", "4")

	n6 := r.AddLink("e", "5").
		AddLink("b", "7").
		AddLink("c", "4")

	assert.True(t, r.Contains(r))   // {} | {}
	assert.True(t, n1.Contains(n1)) // A5,C6,B7 | A5,C6,B7

	assert.True(t, n2.Contains(r))    // B7,C6 | {}
	require.False(t, n2.Contains(n1)) // B7,C6 | A5,C6,B7
	assert.False(t, r.Contains(n2))   // {} | B7,C6

	require.True(t, n1.Contains(n4))  // A5,C6,B7 | A5
	require.True(t, n1.Contains(n2))  // A5,C6,B7 | B7,C6
	require.False(t, n1.Contains(n3)) // A5,C6,B7 | B7,C4

	require.False(t, n3.Contains(n5)) // B7,C4 | A5,C6,B7
	require.False(t, n3.Contains(n5)) // B7,C4 | D9,B7,C4
	require.False(t, n3.Contains(n2)) // B7,C4 | B7,C6

	require.False(t, n6.Contains(n5)) // E5,B7,C4 | D9,B7,C4
}

func TestNodeIsRoot(t *testing.T) {
	t.Parallel()

	r := New()
	assert.True(t, r.IsRoot())
	subnode := r.AddLink("key1", "val1")
	assert.False(t, subnode.IsRoot())
}

func ExampleNode_Data() {
	node := New().
		AddLink("foo", "1").
		AddLink("bar", "2").
		AddLink("baz", "3")

	for iter := node; !iter.IsRoot(); {
		prev, key, value := iter.Data()
		fmt.Printf("%s: %s\n", key, value)
		iter = prev
	}

	// Output:
	// bar: 2
	// baz: 3
	// foo: 1
}

func TestNodeGetData(t *testing.T) {
	t.Parallel()

	r := New()

	parent, key, val := r.Data()
	assert.True(t, parent == r)
	assert.Equal(t, "", key)
	assert.Equal(t, "", val)

	subnode1 := r.AddLink("key_m", "val_m")
	parent, key, val = subnode1.Data()
	assert.True(t, parent == r)
	assert.Equal(t, "key_m", key)
	assert.Equal(t, "val_m", val)

	subnode2 := subnode1.AddLink("key_a", "val_a")
	parent, key, val = subnode2.Data()
	assert.True(t, parent == subnode1)
	assert.Equal(t, "key_a", key)
	assert.Equal(t, "val_a", val)

	subnode3 := subnode2.AddLink("key_z", "val_z")
	parent, key, val = subnode3.Data()
	// This is sorted after key_m, so its path should be key_a->key_m->key_z->root
	assert.Equal(t, "key_a", key)
	assert.Equal(t, "val_a", val)
	assert.True(t, parent.Contains(subnode1))
	assert.Equal(t,
		map[string]string{"key_a": "val_a", "key_m": "val_m", "key_z": "val_z"},
		subnode3.Path(),
	)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// from https://stackoverflow.com/a/22892986/5427244
func randSeq() string {
	b := make([]byte, 100)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec
	}
	return string(b)
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	// this just test that adding stuff is not racy
	r := New()
	values := make([]string, 10000)
	keys := make([]string, 15)
	for i := 0; i < len(keys); i++ {
		keys[i] = randSeq()
	}
	for i := 0; i < len(values); i++ {
		values[i] = randSeq()
	}
	concurrency := 128
	repetitions := 10240
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func(wid int) {
			defer wg.Done()
			n := r
			for j := 0; j < repetitions; j++ {
				index := wid + j
				n = n.AddLink(keys[index%len(keys)], values[index%len(keys)])
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkNodeConcurrencyBad(b *testing.B) {
	ixrand := func(nvals int) int {
		return rand.Int() % nvals //nolint:gosec
	}
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			keys := make([]string, 15)
			for i := 0; i < len(keys); i++ {
				keys[i] = randSeq()
			}
			for i := 0; i < len(values); i++ {
				values[i] = randSeq()
			}
			r := New()
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					r.AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					).AddLink(
						keys[ixrand(len(keys))],
						values[ixrand(len(values))],
					)
				}
			})
		})
	}
}

func BenchmarkNodeRealistic(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			values := make([]string, n)
			for i := 0; i < len(values); i++ {
				values[i] = randSeq()
			}
			r := New().
				AddLink("labelone", "valueone").
				AddLink("labeltthree", "valuetthree").
				AddLink("labelfour", "valuefour").
				AddLink("labelfive", "valuefive").
				AddLink("labelsix", "valuefive").
				AddLink("labelseven", "valuefive").
				AddLink("labeleigth", "valuefive").
				AddLink("labeltwo", "valuetwo")
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				i := 0
				for p.Next() {
					i++
					n := r.AddLink(
						"badkey",
						values[i%len(values)],
					)
					if i%2 == 0 {
						n = n.AddLink("labelsix", "someOtheStrangeValue")
					}
					switch i % 7 {
					case 0, 1, 2:
						n.AddLink("okayLabel", "200")
					case 3, 4:
						n.AddLink("okayLabel", "400")
					case 5:
						n.AddLink("okayLabel", "500")
					case 6:
						n.AddLink("okayLabel", "0")
					}
				}
			})
		})
	}
}

func BenchmarkNodeContainsPositive(b *testing.B) {
	r := New()
	n := r.AddLink("labelone", "valueone").
		AddLink("labeltwo", "valuetwo").
		AddLink("labeltthree", "valuetthree").
		AddLink("labelfour", "valuefour").
		AddLink("labelfive", "valuefive").
		AddLink("labelsix", "valuefive").
		AddLink("labelseven", "valuefive").
		AddLink("labeleigth", "valuefive")

	n2 := r.AddLink("labelone", "valueone").
		AddLink("labeltthree", "valuetthree").
		AddLink("labelfour", "valuefour").
		AddLink("labelfive", "valuefive").
		AddLink("labelsix", "valuefive").
		AddLink("labelseven", "valuefive").
		AddLink("labeleigth", "valuefive")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			n.Contains(n2)
		}
	})
}

func BenchmarkNodeContainsNegative(b *testing.B) {
	r := New()
	n := r.AddLink("labelone", "valueone").
		AddLink("labeltwo", "valuetwo").
		AddLink("labeltthree", "valuetthree").
		AddLink("labelfour", "valuefour").
		AddLink("labelfive", "valuefive").
		AddLink("labelsix", "valuefive").
		AddLink("labelseven", "valuefive").
		AddLink("labeleigth", "valuefive")

	n2 := r.AddLink("labelone", "valueone").
		AddLink("labeltthree", "valuetthree").
		AddLink("labelfour", "valuefour").
		AddLink("labelfive", "valuefives"). // wrong value
		AddLink("labelsix", "valuefive").
		AddLink("labelseven", "valuefive").
		AddLink("labeleigth", "valuefive")

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			n.Contains(n2)
		}
	})
}

func BenchmarkNodeValueByKey(b *testing.B) {
	r := New()
	n := r.AddLink("labelone", "valueone").
		AddLink("labeltwo", "valuetwo").
		AddLink("labeltthree", "valuetthree").
		AddLink("labelfour", "valuefour").
		AddLink("labelfive", "valuefive").
		AddLink("labelsix", "valuefive").
		AddLink("labelseven", "valuefive").
		AddLink("labeleigth", "valuefive")

	for name, key := range map[string]string{
		"goodKeyLow":  "labelone",
		"badKeyLow":   "labelone2",
		"goodKeyMid":  "labelfour",
		"badKeyMid":   "labelfour2",
		"goodKeyHigh": "labeleight",
		"badKeyHigh":  "labeleight2",
	} {
		key := key
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(p *testing.PB) {
				for p.Next() {
					n.ValueByKey(key)
				}
			})
		})
	}
}
