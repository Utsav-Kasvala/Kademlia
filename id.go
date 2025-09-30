package main

import (
	"fmt"
	"math/bits"
	"math/rand"
)

const (
	IDBits = 3 // small ID space for demo
	K      = 1  // bucket size
	Alpha  = 1  // concurrency
)

type ID uint16

func RandomID() ID {
	return ID(rand.Intn(1 << IDBits))
}

func IDFromKey(key string) ID {
	var sum uint32
	for i := 0; i < len(key); i++ {
		sum += uint32(key[i])
	}
	return ID(sum % (1 << IDBits))
}

func (id ID) Distance(other ID) ID {
	return id ^ other
}

func (id ID) String() string {
	return fmt.Sprintf("%04x", uint16(id))
}

func bucketIndex(self, id ID) int {
	dist := self.Distance(id)
	if dist == 0 {
		return -1
	}
	// prefix matching
	return bits.Len(uint(dist)) - 1
}
