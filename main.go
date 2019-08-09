package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/contactless/wb-mqtt-noolite/noolite"
	"github.com/contactless/wbgo"
)

type NooliteModel struct {
	wbgo.ModelBase
	nl, nlf    *BaseNooliteDevice
	connection *noolite.Connection
}

func (nm *NooliteModel) Start() error {
	nm.nl = &BaseNooliteDevice{false, nm}
	nm.nlf = &BaseNooliteDevice{true, nm}
	nm.Observer.OnNewDevice(nm.nl)
	nm.Observer.OnNewDevice(nm.nlf)
	var err error
	nm.connection, err = noolite.NewConnection("/dev/ttyUSB0")
	if err != nil {
		return err
	}
	return nil
}

func (nm *NooliteModel) Stop() {
	nm.Observer.RemoveDevice(nm.nl)
	nm.Observer.RemoveDevice(nm.nlf)
}

type BaseNooliteDevice struct {
	isNF bool
	nm   *NooliteModel
}

func (bnd BaseNooliteDevice) Name() string {
	if bnd.isNF {
		return "noolite-f"
	}
	return "noolite"
}

func (bnd BaseNooliteDevice) Title() string {
	if bnd.isNF {
		return "NooLite-F"
	}
	return "NooLite"
}

func (bnd BaseNooliteDevice) Observe(o wbgo.DeviceObserver) {
	o.OnNewControl(bnd, wbgo.Control{
		Name:        "add_device",
		Title:       "Add Device",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
}

func (bnd BaseNooliteDevice) AcceptValue(name, value string) {
	fmt.Println(bnd.Name(), "AcceptValue", name, "=", value)
}

func (bnd BaseNooliteDevice) AcceptOnValue(name string, value string) bool {
	//fmt.Println(bnd.Name(), "AcceptOnValue", name, "=", value)
	if value == "1" {
		bnd.createDevice()
	}
	return true
}

func (bnd BaseNooliteDevice) createDevice() {
	nld := &NooLiteDevice{nm: bnd.nm}
	bnd.nm.Observer.OnNewDevice(nld)
}

func (bnd BaseNooliteDevice) IsVirtual() bool {
	return true
}

type NooLiteDevice struct {
	nm     *NooliteModel
	binded bool
	addr   [4]byte
	ch     byte
	isNF   bool
}

func (nld *NooLiteDevice) Name() string {
	return fmt.Sprintf("NooLiteF-%X", nld.addr)
}

func (nld *NooLiteDevice) Title() string {
	return fmt.Sprintf("noolitef-%X", nld.addr)
}

func (nld *NooLiteDevice) Observe(o wbgo.DeviceObserver) {
	o.OnNewControl(nld, wbgo.Control{
		Name:        "level",
		Title:       "Level",
		Type:        "range",
		Max:         100,
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "state",
		Title:       "State",
		Type:        "switch",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "color",
		Title:       "Color",
		Type:        "rgb",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "slowup",
		Title:       "Slowup",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "slowdown",
		Title:       "Slowdown",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "slowswitch",
		Title:       "slowswitch",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "slowstop",
		Title:       "slowstop",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "shadowlevel",
		Title:       "ShadowLevel",
		Type:        "range",
		Max:         100,
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "bind",
		Title:       "bind",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
	o.OnNewControl(nld, wbgo.Control{
		Name:        "unbind",
		Title:       "unbind",
		Type:        "pushbutton",
		Writability: wbgo.DefaultWritability,
	})
}

func (nld *NooLiteDevice) AcceptValue(name string, value string) {
	panic("not implemented")
}

func (nld *NooLiteDevice) AcceptOnValue(name, value string) bool {
	fmt.Println(nld.Name(), "AcceptOnValue", name, "=", value)
	switch name {
	case "level":
	case "state":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		if value == "1" {
			req.Cmd = 2
		} else if value == "0" {
			req.Cmd = 0
		}
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		_, err = nld.nm.connection.Read()
		if err != nil {
			return false
		}
	case "color":
	case "slowup":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		req.Cmd = 3
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		_, err = nld.nm.connection.Read()
		if err != nil {
			return false
		}
	case "slowdown":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		req.Cmd = 1
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		_, err = nld.nm.connection.Read()
		if err != nil {
			return false
		}
	case "slowswitch":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		req.Cmd = 18
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		_, err = nld.nm.connection.Read()
		if err != nil {
			return false
		}
	case "slowstop":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		req.Cmd = 10
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		_, err = nld.nm.connection.Read()
		if err != nil {
			return false
		}

	case "shadowlevel":
	case "bind":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		req.Cmd = 15
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		resp, err := nld.nm.connection.Read()
		if err != nil {
			return false
		}
		nld.addr = [4]byte{resp.ID0, resp.ID1, resp.ID2, resp.ID3}
	case "unbind":
		req := new(noolite.Request)
		req.Ch = 11
		req.Mode = noolite.NooLiteFTX
		req.Cmd = 9
		err := nld.nm.connection.Write(req)
		if err != nil {
			return false
		}
		_, err = nld.nm.connection.Read()
		if err != nil {
			return false
		}
	}
	return true
}

func (nld *NooLiteDevice) IsVirtual() bool {
	return false
}

func main() {
	wbgo.Debug = log.New(os.Stdout, "", log.LstdFlags)
	wbgo.SetDebuggingEnabled(true)
	model := &NooliteModel{}
	mqttClient := wbgo.NewPahoMQTTClient("ws://192.168.88.104:18883", "test_dev", false)
	driver := wbgo.NewDriver(model, mqttClient)
	driver.SetAutoPoll(true)
	driver.SetPollInterval(time.Second)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	err := driver.Start()
	if err != nil {
		panic(err)
	}
	<-c
	driver.Stop()
}
