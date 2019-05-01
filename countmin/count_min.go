package countmin

import (
	"hash"
	"io"
	"math"
	"strings"
)

// New creates a new CountMin struct for approximately counting distinct string elements.
func New(w, d int, fn func() hash.Hash64) *CountMin {
	sz := w * d
	cm := &CountMin{
		w:     uint64(w),
		d:     d,
		table: make([]uint64, sz, sz),
		hash:  make([]hash.Hash64, d, d),
	}

	for i := 0; i < d; i++ {
		cm.hash[i] = fn()
	}

	return cm
}

type CountMin struct {
	// w is the width and number of cells per function in the sketch.
	w uint64
	// d is the depth and number of hash functions in the sketch.
	d int
	// table is the w*d table of values.
	table []uint64
	// hash is the d hash functions to be applied.
	hash []hash.Hash64
}

func (cm *CountMin) Add(s string) {
	w := cm.w
	for i, h := range cm.hash {
		buf := strings.NewReader(s)
		_, err := io.Copy(h, buf)
		if err != nil {
			panic(err.Error())
		}
		u := h.Sum64()
		h.Reset()
		idx := u%w + uint64(i)*w
		cm.table[idx]++
	}
}

func (cm *CountMin) Count(s string) uint64 {
	w := cm.w
	var count uint64 = math.MaxUint64
	for i, h := range cm.hash {
		buf := strings.NewReader(s)
		_, err := io.Copy(h, buf)
		if err != nil {
			panic(err.Error())
		}
		u := h.Sum64()
		h.Reset()

		idx := u%w + uint64(i)*w
		cur := cm.table[idx]
		if cur < count {
			count = cur
		}
	}

	if count == math.MaxUint64 {
		count = 0
	}

	return count
}
