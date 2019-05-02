// PointEst-Min Sketches based on G. Cormode 2003,2004.
// https://www.cs.rutgers.edu/~muthu/countmin.c
package countmin

import (
	"errors"
	"hash"
	"io"
	"math"
	"math/rand"
	"strings"

	"github.com/nfisher/gstream/hash/murmur2"
)

// New creates a new Sketch struct for approximately counting distinct string elements.
func New(w, d int) *Sketch {
	var seeds []uint64
	for i := 0; i < d; i++ {
		seeds = append(seeds, rand.Uint64())
	}

	return NewWithSeeds(w, d, seeds)
}

// NewWithSeeds creates a new Sketch struct for approximately counting distinct elements using the provided Seeds.
func NewWithSeeds(w, d int, seeds []uint64) *Sketch {
	sz := w * d
	cm := &Sketch{
		Width: uint64(w),
		Depth: d,
		Table: make([]uint64, sz, sz),
		hash:  make([]hash.Hash64, d, d),
		Seeds: seeds,
	}

	for i := 0; i < d; i++ {
		cm.hash[i] = murmur2.New64aWithSeed(seeds[i])
	}

	return cm
}

func InnerProduct(sketch1 *Sketch, sketch2 *Sketch) uint64 {
	products := make([]uint64, sketch1.Depth, sketch1.Depth)
	var min uint64 = math.MaxUint64

	//	for , s1 := sketch1.

	for _, v := range products {
		if v < min {
			min = v
		}
	}

	return min
}

// Merge combines the values into 2 or more Sketch structures into one.
func Merge(sketches ...*Sketch) (*Sketch, error) {
	if sketches == nil || len(sketches) < 2 {
		return nil, ErrCountOfSketchesInMerge
	}

	first := sketches[0]

	for i := 1; i < len(sketches); i++ {
		err := isCompatible(first, sketches[i])
		if err != nil {
			return nil, err
		}
	}

	w := int(first.Width)
	d := first.Depth
	h := first.hash

	combined := New(w, d)
	combined.hash = h
	// copy Seeds...

	for i := range combined.Table {
		for j := range sketches {
			combined.Table[i] += sketches[j].Table[i]
		}
	}

	return combined, nil
}

// Sketch is a CountMin sketch to probabilistically count keys.
type Sketch struct {
	// Width is the width and number of cells per function in the sketch.
	Width uint64
	// Depth is the depth and number of hash functions in the sketch.
	Depth int
	// Table is the Width*Depth Table of values.
	Table []uint64
	// Seeds are used when creating the hashing functions.
	Seeds []uint64
	// hash is the Depth hash functions to be applied.
	hash []hash.Hash64
}

// Sum provides a total sum of the backing Table.
func (cm *Sketch) Sum() uint64 {
	var sum uint64
	for _, v := range cm.Table {
		sum += v
	}
	return sum
}

// Update the Sketch key with the provided delta.
func (cm *Sketch) Update(s string, delta int) {
	w := cm.Width
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
			cm.Table[idx] += uint64(delta)
		} else {
			cm.Table[idx] -= uint64(delta)
		}
	}
}

// PointEst provides a point estimate for the given string.
func (cm *Sketch) PointEst(s string) uint64 {
	w := cm.Width
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
		cur := cm.Table[idx]
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

func isCompatible(sketch1 *Sketch, sketch2 *Sketch) error {
	if sketch1.Depth != sketch2.Depth {
		return ErrMixedDepthIncompatible
	}

	if sketch1.Width != sketch2.Width {
		return ErrMixedWidthIncompatible
	}

	if len(sketch1.Seeds) != len(sketch2.Seeds) {
		return ErrCountOfSeedsIncompatible
	}

	for i, v := range sketch1.Seeds {
		if v != sketch2.Seeds[i] {
			return ErrSeedValuesIncompatible
		}
	}

	return nil
}

var (
	ErrMixedDepthIncompatible   = errors.New("cannot merge structures of differing depths")
	ErrMixedWidthIncompatible   = errors.New("cannot merge structures of differing widths")
	ErrCountOfSketchesInMerge   = errors.New("at least 2 structures are required to merge")
	ErrCountOfSeedsIncompatible = errors.New("the number of Seeds should match Depth")
	ErrSeedValuesIncompatible   = errors.New("the seed values are incompatible")
)
