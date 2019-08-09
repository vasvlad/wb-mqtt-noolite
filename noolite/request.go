package noolite

import (
	"reflect"
	"unsafe"
)

type Request struct {
	st   byte
	Mode byte
	Ctr  byte
	res  byte //unused, always zero
	Ch   byte
	Cmd  byte
	Fmt  byte
	D0   byte
	D1   byte
	D2   byte
	D3   byte
	ID0  byte
	ID1  byte
	ID2  byte
	ID3  byte
	crc  byte
	sp   byte
}

func (r *Request) prepare() {
	r.st = 171
	r.sp = 172
	r.crc = r.st + r.Mode + r.Ctr + r.res + r.Ch + r.Cmd + r.Fmt
	r.crc += r.D0 + r.D1 + r.D2 + r.D3 + r.ID0 + r.ID1 + r.ID2 + r.ID3
}

func (r *Request) toBytes() []byte {
	r.prepare()
	sh := &reflect.SliceHeader{}
	sh.Cap = 17
	sh.Len = 17
	sh.Data = uintptr(unsafe.Pointer(r))
	data := ((*[]byte)(unsafe.Pointer(sh)))
	return *data
}
