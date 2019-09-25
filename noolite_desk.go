package main

import (
	"github.com/contactless/wb-mqtt-noolite/noolite"
	"github.com/evgeny-boger/wbgo"
)

type NooliteBindDesk struct {
	Desk
}

func (nbd *NooliteBindDesk) Initialize() error {
	_, err := nbd.AddButton("add_tx", "Add NooLite-F device")
	if err != nil {
		return err
	}
	_, err := nbd.AddButton("add_sensor", "Add NooLite sensor")
	nbd.events["add_sensor"] = nbd.addSensor
	nbd.events["add_tx"] = nbd.createTX
	return nil
}

func (nbd *NooliteBindDesk) createTX(event wbgo.ControlOnValueEvent) {
	req := new(noolite.Request)
	req.Ch = 1
	req.Mode = noolite.NooLiteFTX
	req.Cmd = noolite.BindCmd
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
	if resp.D0 == noolite.NooLiteFRelay {
		go nbd.d.CreateNooliteF(resp.Ch, resp.ID0, resp.ID1, resp.ID2, resp.ID3, true)
	} else if resp.D0 == noolite.NooLiteFDimmer || resp.D0 == noolite.NooLiteFDimmerRelay {
		go nbd.d.CreateNooliteF(resp.Ch, resp.ID0, resp.ID1, resp.ID2, resp.ID3, false)
	}
}

func (nbd *NooliteBindDesk) addSensor(event wbgo.ControlOnValueEvent) {
	req := new(noolite.Request)
	req.Ch = 1 //Todo select channel
	req.Mode = 1
	req.Ctr = 3
	err := ndb.d.conn.Write(req)
	if err != nil {
		wbgo.Error.Printf("Error on bind NooLite sensor: %s\n", err.Error())
		return

	}
}
