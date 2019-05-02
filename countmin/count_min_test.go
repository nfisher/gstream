package countmin_test

import (
	"encoding/csv"
	"hash"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/nfisher/gstream/countmin"
	"github.com/nfisher/gstream/hash/murmur2"
	"github.com/nfisher/gstream/hash/pearson"
)

func Test_merge_errors(t *testing.T) {
	td := map[string]struct {
		cms []*countmin.Sketch
		err error
	}{
		"at least 2 required": {[]*countmin.Sketch{countmin.New(2, 2)}, countmin.ErrNotEnough},
		"differing depth":     {[]*countmin.Sketch{countmin.New(2, 2), countmin.New(2, 4)}, countmin.ErrMixedDepth},
		"differing width":     {[]*countmin.Sketch{countmin.New(2, 2), countmin.New(4, 2)}, countmin.ErrMixedWidth},
		// TODO: err on mismatched seeds.
	}

	for name, tc := range td {
		t.Run(name, func(t *testing.T) {
			_, err := countmin.Merge(tc.cms...)
			if err != tc.err {
				t.Errorf("err = %v, want %v", err, tc.err)
			}
		})
	}
}

func Test_merge(t *testing.T) {
	cms := []*countmin.Sketch{
		cm(murmur2.New64a, 4, "hello"),
		cm(murmur2.New64a, 4, "bye"),
	}

	combined, err := countmin.Merge(cms...)
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}

	if combined.Sum() != 8 {
		t.Errorf("Sum = %v, want 8", combined.Sum())
	}

	if combined.PointEst("hello") != 1 {
		t.Errorf("PointEst(hello) = %v, want 1", combined.PointEst("hello"))
	}

	if combined.PointEst("bye") != 1 {
		t.Errorf("PointEst(bye) = %v, want 1", combined.PointEst("bye"))
	}
}

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

			cm := countmin.NewWithHash64(1000, 4, tc.nh)

			m := make(map[string]uint64)
			var record []string
			for {
				record, err = r.Read()
				if err != nil {
					break
				}

				name := record[FifaClub]
				cm.Update(name, 1)
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
				actual := cm.PointEst(club)
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

func Test_sum(t *testing.T) {
	sketch := cm(murmur2.New64a, 4, "match", "pants", "shorts")
	actual := sketch.Sum()
	var expected uint64 = 12
	if actual != expected {
		t.Errorf("Sum = %v, want %v", actual, expected)
	}
}

func Test_point_estimate(t *testing.T) {
	td := map[string]struct {
		*countmin.Sketch
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
			actual := tc.PointEst(tc.key)
			if actual != tc.count {
				t.Errorf("PointEst(%v) = %v, want %v", tc.key, actual, tc.count)
			}
		})
	}
}

var AddCount uint64

func Benchmark_add(b *testing.B) {
	cm := cm(murmur2.New64a, 8)
	for i := 0; i < b.N; i++ {
		cm.Update("hello world", 1)
	}
	AddCount = cm.PointEst("hello world")
}

var CountBench uint64

func Benchmark_point_estimate(b *testing.B) {
	var count uint64
	cm := cm(murmur2.New64a, 8, "hello world")
	for i := 0; i < b.N; i++ {
		count = cm.PointEst("hello world")
	}
	CountBench = count
}

func cm(fn func() hash.Hash64, d int, keys ...string) *countmin.Sketch {
	rand.Seed(1556608494)
	cm := countmin.NewWithHash64(1024, d, fn)
	for _, k := range keys {
		cm.Update(k, 1)
	}
	return cm
}

func fifaReader() (*csv.Reader, error) {
	f, err := os.Open("../testdata/fifa.csv")
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)

	return r, nil
}

const (
	FifaNationality = 5
	FifaClub        = 9
)
