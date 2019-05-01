package murmur2

import (
	crand "crypto/rand"
	"encoding/binary"
	"hash"
	"math/rand"
)

// New64a returns a MurmurHash64A hashing algorithm.
func New64a() hash.Hash64 {
	return &Murmur64A{
		seed: rand.Uint64(),
	}
}

// Murmur64A is an implementation of the MurmurHash64A algorithm by Austin Appleby.
type Murmur64A struct {
	sum  uint64
	seed uint64
}

// Write does not currently handle streaming writes.
func (m *Murmur64A) Write(b []byte) (int, error) {
	m.sum = Hash(b, m.seed)
	return len(b), nil
}

// Sum returns the hash's sum as a byte array.
func (m *Murmur64A) Sum(b []byte) []byte { return nil }

// Reset resets the Hash to its initial state.
func (m *Murmur64A) Reset()         { m.sum = 0 }

// Size returns the number of bytes Sum will return.
func (m *Murmur64A) Size() int      { return 8 }

// BlockSize returns the hash's underlying block size.
func (m *Murmur64A) BlockSize() int { return 8 }

// Sum64 returns the hash's 64 bit sum.
func (m *Murmur64A) Sum64() uint64  { return m.sum }

const m uint64 = 0xc6a4a7935bd1e995
const r = 47

// Based on MurmurHash64A by Austin Appleby
// https://github.com/aappleby/smhasher/blob/master/src/MurmurHash2.cpp#L96
func Hash(key []byte, seed uint64) uint64 {
	var h uint64
	var keyLen = uint64(len(key))
	h = seed ^ (keyLen * m)
	var i = 0

	for ; i < len(key)-7; i += 8 {
		var k = binary.LittleEndian.Uint64(key[i : i+8])
		k *= m
		k ^= k >> r
		k *= m

		h ^= k
		h *= m
	}

	bits := keyLen & 7
	if bits >= 7 {
		h ^= uint64(key[keyLen-bits+6]) << 48
	}
	if bits >= 6 {
		h ^= uint64(key[keyLen-bits+5]) << 40
	}
	if bits >= 5 {
		h ^= uint64(key[keyLen-bits+4]) << 32
	}
	if bits >= 4 {
		h ^= uint64(key[keyLen-bits+3]) << 24
	}
	if bits >= 3 {
		h ^= uint64(key[keyLen-bits+2]) << 16
	}
	if bits >= 2 {
		h ^= uint64(key[keyLen-bits+1]) << 8
	}
	if bits >= 1 {
		h ^= uint64(key[keyLen-bits+0])
		h *= m
	}

	h ^= h >> r
	h *= m
	h ^= h >> r

	return h
}

// ick... not sure how I feel about this but eliminates the seed value from new...
func init() {
	var b [8]byte
	_, err := crand.Read(b[:])
	if err != nil {
		panic(err.Error())
	}
	var seed = int64(binary.LittleEndian.Uint64(b[:]))
	rand.Seed(seed)
}
