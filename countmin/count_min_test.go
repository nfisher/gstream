package countmin_test

import (
	"encoding/csv"
	"github.com/nfisher/gstream/countmin"
	"hash"
	"math/rand"
	"os"
	"sort"
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

			cm := countmin.New(1000, 4, tc.nh)

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
	f, err := os.Open("../testdata/fifa.csv")
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)

	return r, nil
}

func Test_count(t *testing.T) {
	td := map[string]struct {
		*countmin.CountMin
		key   string
		count uint64
	}{
		"no entry":        {cm(murmur2.New64a, 4), "none", 0},
		"matched entry":   {cm(murmur2.New64a, 4, "match"), "match", 1},
		"unmatched entry": {cm(murmur2.New64a, 4, "match"), "none", 0},
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
	cm := cm(murmur2.New64a, 8)
	for i := 0; i < b.N; i++ {
		cm.Add("hello world")
	}
	AddCount = cm.Count("hello world")
}

var CountBench uint64

func Benchmark_count(b *testing.B) {
	var count uint64
	cm := cm(murmur2.New64a, 8, "hello world")
	for i := 0; i < b.N; i++ {
		count = cm.Count("hello world")
	}
	CountBench = count
}

func cm(fn func() hash.Hash64, d int, keys ...string) *countmin.CountMin {
	rand.Seed(1556608494)
	cm := countmin.New(1024, d, fn)
	for _, k := range keys {
		cm.Add(k)
	}
	return cm
}

