package pearson

import (
	crand "crypto/rand"
	"encoding/binary"
	"hash"
	"math/rand"
)

func New64() hash.Hash64 {
	p := &Pearson{
		sz:    8,
		table: shuffledTable(tableSize),
	}
	p.Reset()

	return p
}

// Pearson is a variable length pearson hashing algorithm.
type Pearson struct {
	sz    int
	table []byte
	hh    []byte
	wrote int
}

func (p *Pearson) Sum(msg []byte) []byte {
	return nil
}

func (p *Pearson) Sum64() uint64 {
	return binary.LittleEndian.Uint64(p.hh)
}

func (p *Pearson) Write(msg []byte) (int, error) {
	// TODO: ensure multiple calls works correctly.
	tLen := len(p.table)

	for j := 0; j < p.sz; j++ {
		var z int
		if len(msg) > 0 {
			z = int(msg[0])
		}
		h := p.table[(z+j)%tLen]
		p.hh[j] = h
	}

	for i := 1; i < len(msg); i++ {
		for j := 0; j < p.sz; j++ {
			p.hh[j] = p.table[int(p.hh[j]^msg[i])]
		}
	}

	return len(msg), nil
}

func (p *Pearson) Reset() {
	p.hh = make([]byte, p.sz, p.sz)
	p.wrote = 0
}

func (p *Pearson) Size() int { return p.sz }

func (p *Pearson) BlockSize() int { return 1 }

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

func shuffledTable(sz int) []byte {
	var b = make([]byte, 0, sz)
	for i := 0; i < sz; i++ {
		b = append(b, byte(i))
	}

	rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})

	return b
}

const tableSize = 256
