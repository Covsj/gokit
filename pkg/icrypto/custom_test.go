package icrypto

import (
	"fmt"
	"testing"
)

func TestProcessKey(t *testing.T) {
	p := NewCaesarUtil(nil, nil, "", nil)
	ori := "b0W786t56iWc6iy5BhdbWi68y6bw54guFghe6wC"
	fmt.Println(ori)
	enc := p.EncodeMultipleTimes(ori, true)
	fmt.Println(enc)
	dec := p.EncodeMultipleTimes(enc, false)
	fmt.Println(dec)
	fmt.Println(ori == dec)
}
