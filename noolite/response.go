package noolite

import (
	"errors"
	"reflect"
	"unsafe"
)

type Response struct {
	st   byte
	Mode byte
	Ctr  byte
	Togl byte
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

var (
	ErrWrongST  = errors.New("wrong st")
	ErrWrongSP  = errors.New("wrong sp")
	ErrWrongCRC = errors.New("wrong crc")
)

func (r Response) validate() error {
	if r.st != 173 {
		return ErrWrongST
	}
	if r.sp != 174 {
		return ErrWrongSP
	}
	crc := r.st + r.Mode + r.Ctr + r.Togl + r.Ch + r.Cmd + r.Fmt
	crc += r.D0 + r.D1 + r.D2 + r.D3 + r.ID0 + r.ID1 + r.ID2 + r.ID3
	if r.crc != crc {
		return ErrWrongCRC
	}
	return nil
}

func bytesToResponse(raw []byte) (*Response, error) {
	if len(raw) != 17 {
	}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&raw))
	r := (*Response)(unsafe.Pointer(sh.Data))
	err := r.validate()
	if err != nil {
		return nil, err
	}
	return r, nil
}
