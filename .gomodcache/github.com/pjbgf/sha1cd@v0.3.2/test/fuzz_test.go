//go:build gofuzz
// +build gofuzz

package test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/pjbgf/sha1cd"
	"github.com/pjbgf/sha1cd/cgo"
)

func FuzzDeviationDetection(f *testing.F) {
	f.Add([]byte{})

	g := sha1cd.New().(sha1cd.CollisionResistantHash)
	c := cgo.New().(sha1cd.CollisionResistantHash)

	f.Fuzz(func(t *testing.T, in []byte) {
		g.Reset()
		c.Reset()

		g.Write(in)
		c.Write(in)

		gv, gc := g.CollisionResistantSum(nil)
		cv, cc := c.CollisionResistantSum(nil)

		if bytes.Compare(gv, cv) != 0 || gc != cc {
			t.Fatalf("input: %q\n go result: %q %v\ncgo result: %q %v",
				hex.EncodeToString(in), hex.EncodeToString(gv), gc, hex.EncodeToString(cv), cc)
		}
	})
}
