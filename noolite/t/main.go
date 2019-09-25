package main

import "fmt"
import "github.com/contactless/wb-mqtt-noolite/noolite"
import "github.com/evgeny-boger/wbgo"
import "math/rand"
import "sync"

import "time"
import "unsafe"

type callback func(*noolite.Response) error

type MTRF64 struct {
	inBuff          chan *noolite.Request
	nlCallbacks     [64]callback
	nlfUsedChannels [64]bool
	nlUsedChannels  [64]bool
	nlfCallbacks    map[uint32]callback
	conn            *noolite.Connection
	Devices         []*Device
	mu              sync.Mutex
}

func SbytesToUint32(in [4]byte) uint32 {
	return *((*uint32)(unsafe.Pointer(&in)))
}

func Uint32To4Bytes(in uint32) [4]byte {
	return *((*[4]byte)(unsafe.Pointer(&in)))
}

func NewMTRF64(addr string) *MTRF64 {
	m := new(MTRF64)
	m.conn, _ = noolite.NewConnection(addr) //TODO
	m.inBuff = make(chan *noolite.Request, 64)
	m.nlfCallbacks = make(map[uint32]callback)
	go m.run()
	return m
}

func (m *MTRF64) AddNooliteFDevice() {
	req := new(noolite.Request)
	req.Ch = byte(rand.Intn(63))
	req.Mode = noolite.NooLiteFTX
	req.Cmd = noolite.BindCmd
	m.inBuff <- req
}

func (m *MTRF64) AddNooliteSensor() {
	req := new(noolite.Request)
	req.Mode = noolite.NooLiteRx
	req.Ctr = 3
	req.Ch = byte(rand.Intn(63))
	m.inBuff <- req
}

type Device struct {
	IsNewProtocol bool //false - NooLite, true - NooLite-F
	IsTx          bool //sensor (false) or device (true)
	Channel       byte
	Addr          uint32  //Address if NooLite-F
	Type          byte    //Device type, if avaible
	Status        [4]byte //Status, if avaible
	m             *MTRF64
}

func (d *Device) Switch() {
	if d.IsTx {
		req := new(noolite.Request)
		req.Ch = d.Channel
		req.Cmd = 4
		if d.IsNewProtocol {
			addr := Uint32To4Bytes(d.Addr)
			req.ID0 = addr[0]
			req.ID1 = addr[1]
			req.ID2 = addr[2]
			req.ID3 = addr[3]

			req.Mode = noolite.NooLiteFTX
		} else {
			req.Mode = noolite.NooLiteTX
		}
		d.m.inBuff <- req
	}
}

func (m *MTRF64) run() {
	go func() {
		var readCount int32
		needRead := make(chan struct{})
		for {
			select {
			case in := <-m.inBuff:
				println("in buf")
				m.mu.Lock()
				_ = m.conn.Write(in)
				m.mu.Unlock()
				readCount++
			case <-needRead:
				m.mu.Lock()
				resp, err := m.conn.Read()
				m.mu.Unlock()
				if err != nil {
					continue
				}
				if resp.Togl > 0 {
					readCount = int32(resp.Togl)
				}
				if resp.Ctr == 3 {
					d := new(Device)
					d.Channel = resp.Ch
					d.IsNewProtocol = resp.Mode > noolite.NooLiteRx
					d.IsTx = resp.Mode != noolite.NooLiteRx
					d.Type = resp.D0
					d.m = m
					if d.IsNewProtocol {
						addr := [4]byte{resp.ID0, resp.ID1, resp.ID2, resp.ID3}
						d.Addr = SbytesToUint32(addr)
					}
					m.Devices = append(m.Devices, d)
					fmt.Printf("%+v\n", d)
					fmt.Printf("%+v\n", m.Devices)
				}
			default:
				if readCount > 0 {
					needRead <- struct{}{}
				}
			}
		}

	}()
}

func main() {
	wbgo.SetDebuggingEnabled(true)
	m := NewMTRF64("/dev/ttyUSB0")
	m.AddNooliteFDevice()
	//	m.AddNooliteSensor()
	for {
		time.Sleep(time.Second)
		fmt.Printf("%+v\n", m.Devices)
		for _, dev := range m.Devices {
			dev.Switch()
		}
	}
}
