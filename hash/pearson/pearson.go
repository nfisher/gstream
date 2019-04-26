package pearson

import (
	"hash"
	"math/rand"
)

type Pearson struct {
	sz    int
	table []byte
}

func (p *Pearson) Sum(msg []byte) []byte {
	tLen := len(p.table)
	hh := make([]byte, p.sz, p.sz)

	for j := 0; j < p.sz; j++ {
		var z = int(msg[0])
		h := p.table[(z+j)%tLen]
		hh[j] = h
	}

	for i := 1; i < len(msg); i++ {
		for j := 0; j < p.sz; j++ {
			hh[j] = p.table[int(hh[j]^msg[i])]
		}
	}

	return hh
}

func (p *Pearson) Write(b []byte) (int, error) { return 0, nil }

func (p *Pearson) Reset() {}

func (p *Pearson) Size() int { return p.sz }

func (p *Pearson) BlockSize() int { return 0 }

func New(sz int, seed int64) hash.Hash {
	rand.Seed(seed)
	h := &Pearson{
		sz:    sz,
		table: table(256),
	}

	return h
}

func table(sz int) []byte {
	var b = make([]byte, 0, sz)
	for i := 0; i < sz; i++ {
		b = append(b, byte(i))
	}

	rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})

	return b
}