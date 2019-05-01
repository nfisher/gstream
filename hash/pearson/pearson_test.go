package pearson_test

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"math/rand"
	"testing"

	"github.com/nfisher/gstream/hash/murmur2"

	"github.com/nfisher/gstream/hash/pearson"
)

func Test_sum(t *testing.T) {
	td := map[string]struct {
		hash.Hash64
		in  string
		out uint64
	}{
		"string input": {h(), "Hello world!!!", binary.LittleEndian.Uint64([]byte{85, 241, 106, 61, 154, 39, 190, 155})},
		"empty input":  {h(), "", binary.LittleEndian.Uint64([]byte{0, 0, 0, 0, 0, 0, 0, 0})},
	}

	for n, tt := range td {
		t.Run(n, func(t *testing.T) {
			io.Copy(tt, bytes.NewBufferString(tt.in))
			sum := tt.Sum64()

			if sum != tt.out {
				t.Errorf("got sum = %v, want %v", sum, tt.out)
			}
		})
	}
}

var Sum uint64

func Benchmark_sum(b *testing.B) {
	small := []byte("hello world!!!!!")
	td := map[string]struct {
		h  hash.Hash64
		in []byte
	}{
		"pearson small": {pearson.New64(), small},
		"pearson large": {pearson.New64(), sumBenchInput},
		"mm2 small":     {murmur2.New64a(), small},
		"mm2 large":     {murmur2.New64a(), sumBenchInput},
	}

	for name, tc := range td {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = tc.h.Write(tc.in)
				Sum = tc.h.Sum64()
			}
		})
	}
}

var sumBenchInput = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam sit amet urna nec lectus pulvinar pharetra at nec neque. Nunc fermentum elit in ligula porttitor lobortis. Suspendisse placerat, purus consectetur auctor hendrerit, magna elit aliquet justo, nec varius magna nunc quis risus. Quisque lacinia nibh et massa mattis pretium. Sed nulla orci, pulvinar non massa quis, aliquet facilisis metus. Aenean a tempor elit. Cras ultricies porta mi vel scelerisque. Proin id elit blandit, tristique risus ut, mattis dolor. Nam egestas lectus ex, id fermentum mauris accumsan et. Sed ut est sem. Etiam tincidunt nec odio sit amet accumsan. Praesent consectetur, felis quis dignissim ultrices, nunc turpis consequat urna, at placerat nibh magna sed eros. Pellentesque neque magna, pellentesque eu tempor non, eleifend et justo. Nullam eleifend, dolor sed congue pulvinar, nulla velit viverra lorem, sit amet sodales lorem diam ut erat. Curabitur libero ex, vestibulum non massa vitae, auctor varius lectus. Quisque eleifend amet.`)

func h() hash.Hash64 {
	rand.Seed(12345)
	return pearson.New64()
}
