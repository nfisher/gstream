package pearson_test

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"math/rand"
	"testing"

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

func h() hash.Hash64 {
	rand.Seed(12345)
	return pearson.New64()
}
