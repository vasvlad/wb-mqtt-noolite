package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/contactless/wb-mqtt-noolite/noolite"
	"github.com/evgeny-boger/wbgo"
)

type Driver struct {
	wbgo.DeviceDriver
	desks  map[string]*Desk
	conn   *noolite.Connection
	loops  []func()
	ticker *time.Ticker
	wbgo.DeviceFilter
}

func NewDriver(driverArgs *wbgo.DriverArgs) (*Driver, error) {
	d := new(Driver)
	d.desks = make(map[string]*Desk)
	var err error
	d.DeviceDriver, err = wbgo.NewDriverBase(driverArgs)
	if err != nil {
		return nil, err
	}
	if err := d.StartLoop(); err != nil {
		return nil, err
	}
	d.conn, err = noolite.NewConnection("/dev/ttyUSB0") //TODO
	if err != nil {
		return nil, err
	}
	d.OnDriverEvent(d.HandleEvent)
	d.WaitForReady()
	d.DeviceFilter = wbgo.NewDeviceListFilter("nlf+", "noolite-f")
	d.SetFilter(d)
	d.WaitForReady()
	d.ticker = time.NewTicker(5 * time.Second)
	go func() {
		for {
			_, ok := <-d.ticker.C
			if !ok {
				break
			}
			for _, f := range d.loops {
				go f()
			}
		}
	}()
	return d, nil
}

func (d *Driver) MatchTopic(topic string) bool {
	if strings.Contains(topic, "/devices/nlf-") {
		fmt.Println(topic)
		var kind string
		var id0, id1, id2, id3, ch byte
		fmt.Sscanf(topic, "/devices/nlf-%s_%X-%X_%X_%X_%X", &kind, &ch, &id0, &id1, &id2, &id3)
		if kind == "r" {
			id := fmt.Sprintf("nlf-r_%X-%X_%X_%X_%X", ch, id0, id1, id2, id3)
			title := fmt.Sprintf("NooLite-F Relay %X%X%X%X", id0, id1, id2, id3)
			wbgo.Debug.Printf("Try to create relayDesk %s[%s]\n", title, id)
			desk, err := d.AddDesk(id, title)
			wbgo.Debug.Printf("Try to create relayDesk %s[%+v]\n", err, desk)
			if err != nil {
				wbgo.Debug.Printf("Error on crate pane: %s", err)
				return false
			}
			rd := &RelayDesk{*desk, [4]byte{id0, id1, id2, id3}, ch, false}
			wbgo.Debug.Printf("Try to initialize NL-F relay desk\n")
			err = rd.initialize()
			if err != nil {
				wbgo.Error.Printf("Error on initialize relay pane: %s", err)
				return false
			}
		}
		return true
	}
	return false
}

func (d *Driver) AddDesk(id, title string) (*Desk, error) {
	err := d.AccessAsync(func(tx wbgo.DriverTx) error {
		deviceArgs := wbgo.NewLocalDeviceArgs().SetId(id).
			SetTitle(title).SetVirtual(false).SetDoLoadPrevious(false)
		dev, err := tx.CreateDevice(deviceArgs)()
		wbgo.Debug.Println("Create device")
		if err != nil {
			wbgo.Error.Printf("Error on create device: %s\n", err)
			return err
		}
		desk := &Desk{dev, make(map[string]wbgo.Control), d, make(map[string]func(e wbgo.ControlOnValueEvent))}
		d.desks[id] = desk
		return nil
	})()

	if err != nil {
		return nil, err
	}
	return d.desks[id], nil
}

func (d *Driver) HandleEvent(e wbgo.DriverEvent) {
	switch event := e.(type) {
	case wbgo.ControlOnValueEvent:
		desk := d.desks[event.Control.GetDevice().GetId()]
		if desk != nil {
			handler := desk.events[event.Control.GetId()]
			if handler != nil {
				wbgo.Debug.Printf("Call handler for %s->%s\n", desk.GetId(), event.Control.GetId())
				handler(event)
			}
		}
	}
}

type NooliteBindDesk struct {
	Desk
}

func (nbd *NooliteBindDesk) Initialize() error {
	_, err := nbd.AddButton("add_tx", "Add NooLite-F device")
	if err != nil {
		return err
	}
	nbd.events["add_tx"] = nbd.createTX
	_, err = nbd.AddButton("add_rx", "Add NooLite-F sensor")
	if err != nil {
		return err
	}
	return nil
}

func (nbd *NooliteBindDesk) createTX(event wbgo.ControlOnValueEvent) {
	req := new(noolite.Request)
	req.Ch = 1
	req.Mode = noolite.NooLiteFTX
	req.Cmd = 15
	err := nbd.d.conn.Write(req)
	if err != nil {
		wbgo.Error.Printf("Error on bind NooLite-F TX: %s\n", err.Error())
		return
	}
	resp, err := nbd.d.conn.Read()
	if err != nil {
		wbgo.Error.Printf("Error on bind NooLite-F TX: %s\n", err.Error())
		return
	}
	wbgo.Debug.Println("Create relay desk")
	go nbd.createRelayDesk(resp)
}

func (nbd *NooliteBindDesk) createRelayDesk(resp *noolite.Response) {
	id := fmt.Sprintf("nlf-r_%X-%X_%X_%X_%X", resp.Ch, resp.ID0, resp.ID1, resp.ID2, resp.ID3)
	title := fmt.Sprintf("NooLite-F Relay %X%X%X%X", resp.ID0, resp.ID1, resp.ID2, resp.ID3)
	wbgo.Debug.Printf("Try to create relayDesk %s[%s]\n", title, id)
	d, err := nbd.d.AddDesk(id, title)
	wbgo.Debug.Printf("Try to create relayDesk %s[%+v]\n", err, d)
	if err != nil {
		wbgo.Debug.Printf("Error on crate pane: %s", err)
		return
	}
	rd := &RelayDesk{*d, [4]byte{resp.ID0, resp.ID1, resp.ID2, resp.ID3}, resp.Ch, false}
	wbgo.Debug.Printf("Try to initialize NL-F relay desk\n")
	err = rd.initialize()
	if err != nil {
		wbgo.Error.Printf("Error on initialize relay pane: %s", err)
	}
}

type RelayDesk struct {
	Desk
	addr [4]byte
	ch   byte
	on   bool
}

func (rd *RelayDesk) initialize() error {
	rd.updateStatus()
	_, err := rd.AddSwitch("power", "State", rd.on, true)
	if err != nil {
		return err
	}
	rd.events["power"] = rd.toogle
	rd.d.loops = append(rd.d.loops, rd.updateStatus)
	return nil
}

func (rd *RelayDesk) toogle(e wbgo.ControlOnValueEvent) {
	var cmd byte = 0
	if e.RawValue == "1" {
		cmd = 2
	}
	if (rd.on && cmd == 2) || (!rd.on && cmd == 0) {
		return
	}
	req := new(noolite.Request)
	req.Ch = 1
	req.Ctr = 8
	req.Mode = noolite.NooLiteFTX
	req.Cmd = cmd
	req.ID0 = rd.addr[0]
	req.ID1 = rd.addr[1]
	req.ID2 = rd.addr[2]
	req.ID3 = rd.addr[3]
	err := rd.d.conn.Write(req)
	if err != nil {
		wbgo.Error.Printf("Error on sedn command to noolite-f relay: %s", err)
		return
	}
	_, err = rd.d.conn.Read()
	if err != nil {
		wbgo.Error.Printf("Error on sedn command to noolite-f relay: %s", err)
		return
	}
	return

}

func (rd *RelayDesk) updateStatus() {
	req := new(noolite.Request)
	req.Ch = 1
	req.Ctr = 8
	req.Mode = noolite.NooLiteFTX
	req.Cmd = 128 //ReadState
	req.ID0 = rd.addr[0]
	req.ID1 = rd.addr[1]
	req.ID2 = rd.addr[2]
	req.ID3 = rd.addr[3]
	err := rd.d.conn.Write(req)
	if err != nil {
		wbgo.Error.Printf("Error on sedn command to noolite-f relay: %s", err)
		return
	}
	resp, err := rd.d.conn.Read()
	if err != nil {
		wbgo.Error.Printf("Error on sedn command to noolite-f relay: %s", err)
		return
	}
	old := rd.on
	rd.on = resp.D2 == 1
	if rd.on != old {
		ctrl, ok := rd.controllers["power"]
		if !ok {
			return
		}
		err = rd.d.Access(func(tx wbgo.DriverTx) error {
			ctrl.SetTx(tx)
			err = ctrl.SetOnValue(rd.on)()
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			wbgo.Error.Printf("Error on update value: %s\n", err)
		}
	}
}

type Desk struct {
	wbgo.LocalDevice
	controllers map[string]wbgo.Control
	d           *Driver
	events      map[string]func(e wbgo.ControlOnValueEvent)
}

func (d *Desk) addControl(id string, controlArgs *wbgo.ControlArgs) (wbgo.Control, error) {
	err := d.d.AccessAsync(func(tx wbgo.DriverTx) error {
		d.SetTx(tx)
		ctrl, err := d.CreateControl(controlArgs)()
		if err != nil {
			return err
		}
		d.controllers[id] = ctrl
		return nil
	})()
	if err != nil {
		return nil, err
	}
	return d.controllers[id], nil
}

func (d *Desk) AddSwitch(id, title string, defaultValue, isWritable bool) (wbgo.Control, error) {
	controlArgs := wbgo.NewControlArgs().
		SetId(id).
		SetDescription(title).
		SetType("switch").
		SetWritable(isWritable)
	if defaultValue {
		controlArgs.SetRawValue("1")
	} else {
		controlArgs.SetRawValue("0")
	}
	return d.addControl(id, controlArgs)
}

func (d *Desk) AddButton(id, title string) (wbgo.Control, error) {
	controlArgs := wbgo.NewControlArgs().
		SetId(id).
		SetDescription(title).
		SetType("pushbutton").
		SetWritable(true)
	return d.addControl(id, controlArgs)
}

func main() {
	wbgo.SetDebugLogger(log.New(os.Stdin, "", log.LstdFlags), false)
	wbgo.SetDebuggingEnabled(true)
	client := wbgo.NewPahoMQTTClient("tcp://192.168.88.100:1883", "noolite-f")
	driverArgs := wbgo.NewDriverArgs()
	driverArgs.SetId("noolite-f")
	driverArgs.SetMqtt(client)
	driverArgs.SetUseStorage(true)
	driverArgs.SetStoragePath("/tmp/storage")
	driverArgs.SetReownUnknownDevices(true)
	driver, err := NewDriver(driverArgs)
	if err != nil {
		wbgo.Error.Fatalf("%+v", err)
	}
	desk, err := driver.AddDesk("noolite-f", "NooLite-F Control")
	if err != nil {
		wbgo.Error.Fatal(err)
	}
	cd := &NooliteBindDesk{*desk}
	err = cd.Initialize()
	if err != nil {
		wbgo.Error.Fatal(err)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	<-ch
	driver.StopLoop()
	driver.Close()
}
