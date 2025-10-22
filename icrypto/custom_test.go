package icrypto

import (
	"fmt"
	"testing"
)

func TestProcessKey(t *testing.T) {
	p := NewCaesarUtil(nil, nil, "", nil)
	ori := "grape connect shuffle guide cradle coast climb report target weird prefer pistol"
	fmt.Println(ori)
	enc := p.EncodeMultipleTimes(ori, true)
	fmt.Println(enc)
	dec := p.EncodeMultipleTimes(enc, false)
	fmt.Println(dec)
	fmt.Println(ori == dec)
}
