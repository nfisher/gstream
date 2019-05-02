// PointEst-Min Sketches based on G. Cormode 2003,2004.
// https://www.cs.rutgers.edu/~muthu/countmin.c
package countmin

import (
	"errors"
	"hash"
	"io"
	"math"
	"strings"

	"github.com/nfisher/gstream/hash/murmur2"
)

var (
	ErrMixedDepth = errors.New("cannot merge structures of differing depths")
	ErrMixedWidth = errors.New("cannot merge structures of differing widths")
	ErrNotEnough  = errors.New("at least 2 structures are required to merge")
)

// NewWithHash64 creates a new Sketch struct for approximately counting distinct string elements.
func NewWithHash64(w, d int, fn func() hash.Hash64) *Sketch {
	sz := w * d
	cm := &Sketch{
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

// New creates a new Sketch struct for approximately counting distinct string elements.
func New(w, d int) *Sketch {
	return NewWithHash64(w, d, murmur2.New64a)
}

func isCompatible(cm1 *Sketch, cm2 *Sketch) error {
	if cm1.d != cm2.d {
		return ErrMixedDepth
	}
	if cm1.w != cm2.w {
		return ErrMixedWidth
	}
	// TODO: compare seeds
	return nil
}

// Merge combines the values into 2 or more Sketch structures into one.
func Merge(cms ...*Sketch) (*Sketch, error) {
	if len(cms) < 2 {
		return nil, ErrNotEnough
	}

	first := cms[0]

	for i := 1; i < len(cms); i++ {
		err := isCompatible(first, cms[i])
		if err != nil {
			return nil, err
		}
	}

	w := int(first.w)
	d := first.d
	h := first.hash

	combined := New(w, d)
	combined.hash = h

	for i := range combined.table {
		for j := range cms {
			combined.table[i] += cms[j].table[i]
		}
	}

	return combined, nil
}

type Sketch struct {
	// w is the width and number of cells per function in the sketch.
	w uint64
	// d is the depth and number of hash functions in the sketch.
	d int
	// table is the w*d table of values.
	table []uint64
	// hash is the d hash functions to be applied.
	hash []hash.Hash64
}

func (cm *Sketch) Sum() uint64 {
	var sum uint64
	for _, v := range cm.table {
		sum += v
	}
	return sum
}

// Update the Sketch key with the provided delta.
func (cm *Sketch) Update(s string, delta int) {
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

		if delta > 0 {
			cm.table[idx] += uint64(delta)
		} else {
			cm.table[idx] -= uint64(delta)
		}
	}
}

func (cm *Sketch) PointEst(s string) uint64 {
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

func (cm *Sketch) PointMed() uint64 {
	return 0
}
