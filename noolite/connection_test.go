package noolite

import (
	"errors"
	"reflect"
	"testing"
)

type FakeUart struct {
	awaitReq    []byte
	successRead bool
	needFail    bool
}

func (fu FakeUart) Read(d []byte) (count int, err error) {
	if fu.successRead {
		d[0] = 173
		d[16] = 174
		if !fu.needFail {
			d[15] = 173
		}
		return 17, nil
	}
	return 0, errors.New("need error")
}

func (fu FakeUart) Write(d []byte) (count int, err error) {
	count = len(d)
	if !reflect.DeepEqual(d, fu.awaitReq) {
		return 0, errors.New("not equal")
	}
	return count, nil
}

func (fu FakeUart) Close() error {
	return nil
}

func TestWrite(t *testing.T) {
	c := new(Connection)
	fu := new(FakeUart)
	c.uart = fu

	req := new(Request)

	fu.awaitReq = req.toBytes()

	err := c.Write(req)
	if err != nil {
		t.Error("Fail on write!")
	}

	fu.awaitReq = []byte{}

	err = c.Write(req)
	if err == nil {
		t.Error("Await error, but it's nil")
	}
}

func TestRead(t *testing.T) {
	c := new(Connection)
	fu := new(FakeUart)
	c.uart = fu

	fu.successRead = true
	fu.needFail = false

	_, err := c.Read()
	if err != nil {
		t.Error("fail on read")
	}

	fu.needFail = true
	_, err = c.Read()

	if err == nil {
		t.Error("want crc error, but nil")
	}

	fu.successRead = false
	if err == nil {
		t.Error("want read error, but nil")
	}

}
