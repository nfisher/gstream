package gstream_test

import (
	"encoding/binary"
	"errors"
	"github.com/nfisher/gstream/hash/pearson"
	"hash"
	"math"
	"math/rand"
	"testing"
)

func Test_count(t *testing.T) {
	td := map[string]struct{
		*CountMin
		key string
		count uint64
	}{
		"no entry": {cm(1, ), "none", 0},
		"matched entry": {cm(1, "match"), "match", 1},
		"multi hash no entry": {cm(4), "none", 0},
		"multi hash matched entry": {cm(4, "match"), "match", 1},
	}

	for n, tc := range td {
		t.Run(n, func(t *testing.T) {
			actual, err := tc.Count(tc.key)
			if err != nil {
				t.Fatalf("Count(%v) err = %v, want nil", tc.key, err)
			}

			if actual != tc.count {
				t.Errorf("Count(%v) = %v, want %v", tc.key, actual, tc.count)
			}
		})
	}
}

var AddCount uint64
var AddErr error
func Benchmark_add(b *testing.B) {
	var err error
	cm := cm(8)
	for i := 0; i < b.N; i++ {
		err = cm.Add("hello world")
	}
	AddErr = err
	AddCount, _ = cm.Count("hello world")
}

var CountBench uint64
func Benchmark_count(b *testing.B) {
	var count uint64
	cm := cm(8, "hello world")
	for i := 0; i < b.N; i++ {
		count, _ = cm.Count("hello world")
	}
	CountBench = count
}


func cm(d int, keys ...string) *CountMin {
	rand.Seed(12345)
	cm := New(1000, d)
	for _, k := range keys {
		_ = cm.Add(k)
	}
	return cm
}

// New creates a new CountMin struct for counting distinct stuff.
func New(w, d int) *CountMin {
	sz := w * d
	cm := &CountMin{
		w:     uint64(w),
		d:     d,
		table: make([]uint64, sz, sz),
		hash:  make([]hash.Hash, d, d),
	}

	for i := 0; i < d; i++ {
		cm.hash[i] = pearson.New(HashSumSize)
	}

	return cm
}

const HashSumSize = 8
var ErrInvalidInt = errors.New("invalid int64 marshalled from checksum")

type CountMin struct {
	// w is the width and number of cells per function in the sketch.
	w uint64
	// d is the depth and number of hash functions in the sketch.
	d     int
	// table is the w*d table of values.
	table []uint64
	// hash is the d hash functions to be applied.
	hash  []hash.Hash
}

func (cm *CountMin) Add(s string) error {
	bytes := []byte(s)
	w := cm.w
	for i, h := range cm.hash {
		b := h.Sum(bytes)
		u, n := binary.Uvarint(b)
		if n < 1 {
			return ErrInvalidInt
		}
		idx := u % w + uint64(i) * w
		cm.table[idx]++
	}

	return nil
}

func (cm *CountMin) Count(s string) (uint64, error) {
	var count uint64 = math.MaxUint64
	bytes := []byte(s)
	w := cm.w
	for i, h := range cm.hash {
		b := h.Sum(bytes)
		u, n := binary.Uvarint(b)
		if n < 1 {
			return 0, ErrInvalidInt
		}

		idx := u % w + uint64(i) * w
		cur := cm.table[idx]
		if cur < count {
			count = cur
		}
	}

	return count, nil
}
