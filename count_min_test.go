package gstream_test

import (
	"encoding/csv"
	"hash"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/nfisher/gstream/hash/murmur2"
	"github.com/nfisher/gstream/hash/pearson"
)

const (
	FifaNationality = 5
	FifaClub        = 9
)

func Test_open(t *testing.T) {
	td := map[string]struct {
		nh func() hash.Hash64
	}{
		"pearson": {pearson.New64},
		"murmur2": {murmur2.New64a},
	}

	for name, tc := range td {
		t.Run(name, func(t *testing.T) {
			r, err := fifaReader()
			if err != nil {
				t.Fatal(err.Error())
			}

			cm := New(1000, 4, tc.nh)

			m := make(map[string]uint64)
			var record []string
			for {
				record, err = r.Read()
				if err != nil {
					break
				}

				name := record[FifaClub]
				cm.Add(name)
				v := m[name]
				v++
				m[name] = v
			}

			var clubs sort.StringSlice
			for k := range m {
				clubs = append(clubs, k)
			}
			sort.Sort(sort.Reverse(clubs))
			total := len(m)

			var miscounts int
			for _, club := range clubs {
				actual := cm.Count(club)
				expected := m[club]
				if actual != expected {
					miscounts++
				}
			}

			percentage := float64(miscounts) / float64(total)
			if percentage > 0.10 {
				t.Errorf("got %0.4v%% (%v) wrong, want < 10%%", percentage*100.0, miscounts)
			}
		})
	}
}

func fifaReader() (*csv.Reader, error) {
	f, err := os.Open("testdata/fifa.csv")
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)

	return r, nil
}

func Test_count(t *testing.T) {
	td := map[string]struct {
		*CountMin
		key   string
		count uint64
	}{
		"no entry":                   {cm(1), "none", 0},
		"matched entry":              {cm(1, "match"), "match", 1},
		"unmatched entry":            {cm(1, "match"), "none", 0},
		"multi hash no entry":        {cm(4), "none", 0},
		"multi hash matched entry":   {cm(4, "match"), "match", 1},
		"multi hash unmatched entry": {cm(4, "match"), "none", 0},
		// TODO: find keys with collisions in only 1 of the hash families.
	}

	for n, tc := range td {
		t.Run(n, func(t *testing.T) {
			actual := tc.Count(tc.key)
			if actual != tc.count {
				t.Errorf("Count(%v) = %v, want %v", tc.key, actual, tc.count)
			}
		})
	}
}

var AddCount uint64

func Benchmark_add(b *testing.B) {
	cm := cm(8)
	for i := 0; i < b.N; i++ {
		cm.Add("hello world")
	}
	AddCount = cm.Count("hello world")
}

var CountBench uint64

func Benchmark_count(b *testing.B) {
	var count uint64
	cm := cm(8, "hello world")
	for i := 0; i < b.N; i++ {
		count = cm.Count("hello world")
	}
	CountBench = count
}

func cm(d int, keys ...string) *CountMin {
	rand.Seed(1556608494)
	cm := New(1024, d, pearson.New64)
	for _, k := range keys {
		cm.Add(k)
	}
	return cm
}

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
