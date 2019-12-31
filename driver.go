package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/vasvlad/wb-mqtt-noolite"
	"github.com/contactless/wbgong"
)

//Driver - common noolite and noolite-f driver
type Driver struct {
	wbgo.DeviceDriver
	desks  map[string]*Desk
	conn   NooliteConnection
	loops  []func()
	ticker *time.Ticker
}

//NewDriver - function constructor of noolite driver
func NewDriver(driverArgs *wbgo.DriverArgs, serial string) (*Driver, error) {
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
	d.conn, err = noolite.NewConnection(serial)
	if err != nil {
		return nil, err
	}
	d.OnDriverEvent(d.handleEvent)
	d.WaitForReady()
	deviceFilter := wbgo.NewDeviceListFilter("noolite-f", "+/controls/power")
	d.SetFilter(deviceFilter)
	d.WaitForReady()
	d.ticker = time.NewTicker(10 * time.Second)
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

func (d *Driver) addDesk(id, title string) (*Desk, error) {
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

func (d *Driver) handleEvent(e wbgo.DriverEvent) {
	switch event := e.(type) {
	case wbgo.ControlOnValueEvent:
		d.changeControlEvent(event)
	case wbgo.NewExternalDeviceEvent:
		d.addExternalDevice(event)
	case wbgo.StopEvent:
		d.WaitForReady()
		d.conn.Close()
	}
}

func (d *Driver) changeControlEvent(event wbgo.ControlOnValueEvent) {
	desk := d.desks[event.Control.GetDevice().GetId()]
	if desk != nil {
		handler := desk.events[event.Control.GetId()]
		if handler != nil {
			wbgo.Debug.Printf("Call handler for %s->%s\n", desk.GetId(), event.Control.GetId())
			go handler(event)
		}
	}
}

func (d *Driver) addExternalDevice(event wbgo.NewExternalDeviceEvent) {
	id := event.Device.GetId()
	title := event.Device.GetTitle()
	if strings.Contains(id, "nlf-r_") {
		go d.addExternalNLFRelay(id, title)
	} else if strings.Contains(id, "nlf-d_") {
		go d.addExternalNLFDimmer(id, title)
	}
}

func (d *Driver) addExternalNLFRelay(id, title string) {
	var ch, id0, id1, id2, id3 byte
	_, err := fmt.Sscanf(id, "nlf-r_%X-%X_%X_%X_%X", &ch, &id0, &id1, &id2, &id3)
	if err != nil {
		wbgo.Debug.Printf("Error parse id: %s", err)
		return
	}
	d.CreateNooliteF(ch, id0, id1, id2, id3, true)
}

func (d *Driver) addExternalNLFDimmer(id, title string) {
	var ch, id0, id1, id2, id3 byte
	_, err := fmt.Sscanf(id, dimmerIDMask, &ch, &id0, &id1, &id2, &id3)
	if err != nil {
		wbgo.Debug.Printf("Error parse id: %s", err)
		return
	}
	d.CreateNooliteF(ch, id0, id1, id2, id3, false)
}

const (
	relayIDMask     = "nlf-r_%X-%X_%X_%X_%X"
	dimmerIDMask    = "nlf-d_%X-%X_%X_%X_%X"
	relayTitleMask  = "NooLite-F Relay %X%X%X%X"
	dimmerTitleMask = "NooLite-F Dimmer %X%X%X%X"
)

//CreateNooliteF - create new noolite-f device (realy or dimmer)
func (d *Driver) CreateNooliteF(ch, id0, id1, id2, id3 byte, isRelay bool) {
	var idMask, titleMask string
	if isRelay {
		idMask = relayIDMask
		titleMask = relayTitleMask
	} else {
		idMask = dimmerIDMask
		titleMask = dimmerTitleMask
	}
	id := fmt.Sprintf(idMask, ch, id0, id1, id2, id3)
	title := fmt.Sprintf(titleMask, id0, id1, id2, id3)
	wbgo.Debug.Printf("Try to create %s[%s]\n", title, id)
	desk, err := d.addDesk(id, title)
	if err != nil {
		wbgo.Error.Printf("Error on create %s: %s", title, err)
		return
	}
	var init interface {
		initialize() error
	}
	addr := [4]byte{id0, id1, id2, id3}
	if isRelay {
		init = &RelayDesk{*desk, addr, ch, false}
	} else {
		init = &DimmerDesk{*desk, addr, ch, false, 255}
	}
	go func() {
		err := init.initialize()
		if err != nil {
			wbgo.Error.Printf("Error on initialize %s: %s", title, err)
		}
	}()
}
