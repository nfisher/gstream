package countmin_test

import (
	"encoding/csv"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/nfisher/gstream/countmin"
)

func Test_inner_product(t *testing.T) {
	td := map[string]struct {
		k1      []string
		k2      []string
		product uint64
	}{
		"should equal 0 with empty sketch": {[]string{}, []string{}, 0},
		"should equal 0 with s2 empty":     {[]string{"hello"}, []string{}, 0},
		"should equal 0 with s1 empty":     {[]string{}, []string{"hello"}, 0},
		"should equal positive product":    {repeat("hello", 10), []string{"hello", "weak"}, 10},
	}

	for name, tc := range td {
		t.Run(name, func(t *testing.T) {
			s1 := countmin.New(1024, 4)
			add(s1, tc.k1...)
			s2 := countmin.NewWithSeeds(1024, 4, s1.Seeds)
			add(s2, tc.k2...)

			product, _ := countmin.InnerProduct(s1, s2)
			if product != tc.product {
				t.Errorf("product = %v, want %v", product, tc.product)
			}
		})
	}
}

func repeat(s string, n int) []string {
	var a []string
	for i := 0; i < n; i++ {
		a = append(a, s)
	}
	return a
}

func add(sketch *countmin.Sketch, keys ...string) {
	for _, k := range keys {
		sketch.Update(k, 1)
	}
}

func Test_merge_errors(t *testing.T) {
	td := map[string]struct {
		sketches []*countmin.Sketch
		err      error
	}{
		"empty not allowed":   {[]*countmin.Sketch{}, countmin.ErrCountOfSketchesInMerge},
		"at least 2 required": {[]*countmin.Sketch{countmin.New(2, 2)}, countmin.ErrCountOfSketchesInMerge},
		"differing depth":     {[]*countmin.Sketch{countmin.New(2, 2), countmin.New(2, 4)}, countmin.ErrMixedDepthIncompatible},
		"differing width":     {[]*countmin.Sketch{countmin.New(2, 2), countmin.New(4, 2)}, countmin.ErrMixedWidthIncompatible},
		// TODO: err on mismatched Seeds.
	}

	for name, tc := range td {
		t.Run(name, func(t *testing.T) {
			_, err := countmin.Merge(tc.sketches...)
			if err != tc.err {
				t.Errorf("err = %v, want %v", err, tc.err)
			}
		})
	}
}

func Test_merge(t *testing.T) {
	sketches := []*countmin.Sketch{
		cm(4, "hello"),
		cm(4, "bye"),
	}

	combined, err := countmin.Merge(sketches...)
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

func Test_with_FIFA_data(t *testing.T) {
	r, err := fifaReader()
	if err != nil {
		t.Fatal(err.Error())
	}

	sketch := countmin.New(1000, 4)

	m := make(map[string]uint64)
	var record []string
	for {
		record, err = r.Read()
		if err != nil {
			break
		}

		name := record[FifaClub]
		sketch.Update(name, 1)
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
		actual := sketch.PointEst(club)
		expected := m[club]
		if actual != expected {
			miscounts++
		}
	}

	percentage := float64(miscounts) / float64(total)
	if percentage > 0.07 {
		t.Errorf("got %0.4v%% (%v) wrong, want < 7%%", percentage*100.0, miscounts)
	}
}

func Test_sum(t *testing.T) {
	sketch := cm(4, "match", "pants", "shorts")
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
		"no entry":        {cm(4), "none", 0},
		"matched entry":   {cm(4, "match"), "match", 1},
		"unmatched entry": {cm(4, "match"), "none", 0},
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
	sketch := cm(8)
	for i := 0; i < b.N; i++ {
		sketch.Update("hello world", 1)
	}
	AddCount = sketch.PointEst("hello world")
}

var CountBench uint64

func Benchmark_point_estimate(b *testing.B) {
	var count uint64
	sketch := cm(8, "hello world")
	for i := 0; i < b.N; i++ {
		count = sketch.PointEst("hello world")
	}
	CountBench = count
}

func cm(d int, keys ...string) *countmin.Sketch {
	rand.Seed(1556608494)
	sketch := countmin.New(1024, d)
	for _, k := range keys {
		sketch.Update(k, 1)
	}
	return sketch
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
