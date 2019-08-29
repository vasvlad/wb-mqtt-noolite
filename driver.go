package main

import (
	"fmt"
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
	d.OnDriverEvent(d.HandleEvent)
	d.WaitForReady()
	d.DeviceFilter = wbgo.NewDeviceListFilter("noolite-f", "+/controls/power")
	d.SetFilter(d)
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
				go handler(event)
			}
		}
	case wbgo.NewExternalDeviceEvent:
		id := event.Device.GetId()
		title := event.Device.GetTitle()
		if strings.Contains(id, "nlf-r_") {
			go func() {
				//NooLite-F Relay
				var ch, id0, id1, id2, id3 byte
				_, err := fmt.Sscanf(id, "nlf-r_%X-%X_%X_%X_%X", &ch, &id0, &id1, &id2, &id3)
				if err != nil {
					wbgo.Debug.Printf("Error parse id: %s", err)
					return
				}
				wbgo.Debug.Printf("Try to create relayDesk %s[%s]\n", title, id)
				desk, err := d.AddDesk(id, title)
				wbgo.Debug.Printf("Try to create relayDesk %s[%+v]\n", err, d)
				if err != nil {
					wbgo.Debug.Printf("Error on crate pane: %s", err)
					return
				}
				rd := &RelayDesk{*desk, [4]byte{id0, id1, id2, id3}, ch, false}
				wbgo.Debug.Printf("Try to initialize NL-F relay desk\n")
				go func() {
					err = rd.initialize()
					if err != nil {
						wbgo.Error.Printf("Error on initialize relay pane: %s", err)
					}
				}()
			}()
		}
	case wbgo.StopEvent:
		d.WaitForReady()
		d.conn.Close()
	}
}
