package gstream_test

import "testing"

func Test_failing(t *testing.T) {
	t.Fail()
}

// New creates a new CountMin struct for counting distinct stuff.
func New(w, d int) CountMin {
	sz := w * d
	return CountMin{
		w: w,
		d: d,
		counters: make([]int, sz, sz),
		hash: make([]hashfn, d, d),
	}
}

type hashfn func(string) int

type CountMin struct {
	w int
	d int
	counters []int
	hash []hashfn
}

func Add(s string) {

}

func Min(s string) int {
	return 0
}
