// Package uuid generates random version 4 UUIDs with crypto/rand, so the
// service needs no extra dependency in the shared go.mod.
package uuid

import (
	"crypto/rand"
	"fmt"
)

// Generator implements ports.IDGenerator.
type Generator struct{}

// NewID returns a new random UUID, for example "6f1c2a7e-3b4d-4c5e-8f90-1a2b3c4d5e6f".
func (Generator) NewID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])      // crypto/rand.Read never returns an error since Go 1.24
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
