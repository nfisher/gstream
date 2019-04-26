package pearson_test

import (
	"bytes"
	"hash"
	"math/rand"
	"testing"

	"github.com/nfisher/gstream/hash/pearson"
)

func Test_sum(t *testing.T) {
	td := map[string]struct{
		hash.Hash
		in []byte
		out []byte
	}{
		"1 byte hash": {h(1), []byte("Hello world!!!"), []byte{85}},
		"16 byte hash": {h(16), []byte("Hello world!!!"), []byte{85, 241, 106, 61, 154, 39, 190, 155, 203, 108, 86, 4, 188, 121, 163, 215}},
	}

	for n, tt := range td {
		t.Run(n, func(t *testing.T) {
			sum := tt.Sum(tt.in)
			if len(sum) != tt.Size() {
				t.Errorf("got sz=%v, want %v", len(sum), tt.Size())
			}

			if !bytes.Equal(sum, tt.out) {
				t.Errorf("got sum=%v, want %v", sum, tt.out)
			}
		})
	}
}

var Sum []byte
func Benchmark_sum_1024(b *testing.B) {
	hash16 := h(16)
	for i := 0; i < b.N; i++ {
		Sum = hash16.Sum(sumBenchInput)
	}
}

func Benchmark_sum_16(b *testing.B) {
	small := []byte("hello world!!!!!")
	hash16 := h(16)
	for i := 0; i < b.N; i++ {
		Sum = hash16.Sum(small)
	}
}

var sumBenchInput = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam sit amet urna nec lectus pulvinar pharetra at nec neque. Nunc fermentum elit in ligula porttitor lobortis. Suspendisse placerat, purus consectetur auctor hendrerit, magna elit aliquet justo, nec varius magna nunc quis risus. Quisque lacinia nibh et massa mattis pretium. Sed nulla orci, pulvinar non massa quis, aliquet facilisis metus. Aenean a tempor elit. Cras ultricies porta mi vel scelerisque. Proin id elit blandit, tristique risus ut, mattis dolor. Nam egestas lectus ex, id fermentum mauris accumsan et. Sed ut est sem. Etiam tincidunt nec odio sit amet accumsan. Praesent consectetur, felis quis dignissim ultrices, nunc turpis consequat urna, at placerat nibh magna sed eros. Pellentesque neque magna, pellentesque eu tempor non, eleifend et justo. Nullam eleifend, dolor sed congue pulvinar, nulla velit viverra lorem, sit amet sodales lorem diam ut erat. Curabitur libero ex, vestibulum non massa vitae, auctor varius lectus. Quisque eleifend amet.`)

func h(sz int) hash.Hash {
	rand.Seed(12345)
	return pearson.New(sz)
}
