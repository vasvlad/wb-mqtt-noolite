package noolite

import (
	"testing"
)

func TestRequestToBytes(t *testing.T) {
	req := new(Request)
	req.Mode = 1
	req.Ctr = 2
	req.res = 3
	req.Ch = 4
	req.Cmd = 5
	req.Fmt = 6
	req.D0 = 7
	req.D1 = 8
	req.D2 = 9
	req.D3 = 10
	req.ID0 = 11
	req.ID1 = 12
	req.ID2 = 13
	req.ID3 = 14
	raw := req.toBytes()
	if raw[0] != 171 {
		t.Error("Wrong start byte")
	}
	if raw[16] != 172 {
		t.Error("Wrong stop byte")
	}
	if raw[15] != 20 {
		t.Errorf("Wrong crc byte, got %d, want %d", raw[15], 20)
	}
	for i := 1; i < 15; i++ {
		if raw[i] != byte(i) {
			t.Errorf("Wrong byte %d, want %d, got %d", i, i, raw[i])
		}
	}
}
