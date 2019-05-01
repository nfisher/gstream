package murmur2_test

import (
	"testing"

	"github.com/nfisher/gstream/hash/murmur2"
)

func Test_hash(t *testing.T) {
	td := map[string]struct {
		expected uint64
	}{
		"hello world":    {2153148270702525707},
		"good night all": {1458475013093136911},
		"1234567":        {105322671949869309},
		"12345678":       {16884033550786057081},
	}

	for name, tc := range td {
		t.Run(name, func(t *testing.T) {
			k := []byte(name)
			v := murmur2.Hash(k, 12345)
			if v != tc.expected {
				t.Errorf("got Sum = %v, want %v", v, tc.expected)
			}
		})
	}
}
