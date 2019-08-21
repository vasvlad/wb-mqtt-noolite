package noolite

import (
	"syscall"

	"github.com/evgeny-boger/wbgo"
	"github.com/schleibinger/sio"
)

//Connection - struct for connect to MTRF-64
type Connection struct {
	uart *sio.Port
}

//NewConnection - return new connection to MTRF-64
func NewConnection(dev string) (*Connection, error) {
	c := new(Connection)
	var err error
	c.uart, err = sio.Open(dev, syscall.B9600)
	if err != nil {
		return nil, err
	}
	err = c.initialize()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Connection) initialize() error {
	req := &Request{
		Mode: 4,
	}
	err := c.Write(req)
	if err != nil {
		return err
	}
	_, err = c.Read()
	if err != nil {
		return err
	}
	return nil
}

//Write - write request to MTRF-64 and lock until read
func (c *Connection) Write(req *Request) error {
	buf := req.toBytes()
	_, err := c.uart.Write(req.toBytes())
	if err != nil {
		return err
	}
	wbgo.Debug.Printf("Sended %+x\n\t%+v\n", buf, req)
	return nil
}

//Read - read response from MTRF-64 and unlock for write
func (c *Connection) Read() (*Response, error) {
	var buf = make([]byte, 17)
	_, err := c.uart.Read(buf)
	if err != nil {
		return nil, err
	}
	resp, err := bytesToResponse(buf)
	if err != nil {
		return nil, err
	}
	wbgo.Debug.Printf("Recieved %+x\n\t%+v\n", buf, resp)
	return resp, nil
}

//Close - close connection to MTRF-64
func (c *Connection) Close() {
	c.uart.Close()
}
