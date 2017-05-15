package security

import (
	"testing"
	"encoding/hex"
	"fmt"
	"8583/utils"
)

func TestEncodeHex(t *testing.T) {
	str := hex.EncodeToString(([]byte)("test"))
	if str != "74657374" {
		t.Error(str)
	}
}

func TestDecodeHex(t *testing.T) {
	d,err := hex.DecodeString("a098")
	if err != nil{
		t.Error(err)
	}
	fmt.Printf("%x\n",d)
	t.Log(utils.EncodeToString(d))
}