package noolite

import (
	"testing"
)

func TestBytesToResponse(t *testing.T) {
	raw := []byte{}
	_, err := bytesToResponse(raw)
	if err == nil {
		t.Error("not error on short slice!")
	}
	raw = make([]byte, 17)
	_, err = bytesToResponse(raw)
	if err != ErrWrongST {
		t.Error("wrong ST but not error!")
	}
	raw[0] = 173
	_, err = bytesToResponse(raw)
	if err != ErrWrongSP {
		t.Error("wrong SP but not error!")
	}
	raw[16] = 174
	_, err = bytesToResponse(raw)
	if err != ErrWrongCRC {
		t.Error("wrong CRC but not error!")
	}
	raw[15] = 173
	_, err = bytesToResponse(raw)
	if err != nil {
		t.Errorf("Valid package but have error: %v", err)
	}

}
