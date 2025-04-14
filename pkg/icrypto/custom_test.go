package icrypto

import (
	"fmt"
	"testing"
)

func TestProcessKey(t *testing.T) {
	p := NewCaesarUtil(nil, nil, "", nil)
	ori := "sA^vWg3zGMRV2S~DS9Bt0xNraE^L59OR162sByD"
	fmt.Println(ori)
	enc := p.EncodeMultipleTimes(ori, true)
	fmt.Println(enc)
	dec := p.EncodeMultipleTimes(enc, false)
	fmt.Println(dec)
	fmt.Println(ori == dec)
}
