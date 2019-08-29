package main

import (
	"fmt"

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
	if resp.D2 == 0 {
		wbgo.Debug.Println("Create relay desk")
		go nbd.createRelayDesk(resp)
	} else {
		wbgo.Debug.Println("Create dimmer desk")
		go nbd.createDimmerDesk(resp)
	}
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

func (nbd *NooliteBindDesk) createDimmerDesk(resp *noolite.Response) {
	id := fmt.Sprintf("nlf-d_%X-%X_%X_%X_%X", resp.Ch, resp.ID0, resp.ID1, resp.ID2, resp.ID3)
	title := fmt.Sprintf("NooLite-F Dimmer %X%X%X%X", resp.ID0, resp.ID1, resp.ID2, resp.ID3)
	wbgo.Debug.Printf("Try to create dimmerDesk %s[%s]\n", title, id)
	d, err := nbd.d.AddDesk(id, title)
	wbgo.Debug.Printf("Try to create DimmerDesk %s[%+v]\n", err, d)
	if err != nil {
		wbgo.Debug.Printf("Error on crate pane: %s", err)
		return
	}
	rd := &DimmerDesk{*d, [4]byte{resp.ID0, resp.ID1, resp.ID2, resp.ID3}, resp.Ch, false, 255}
	wbgo.Debug.Printf("Try to initialize NL-F dimmer desk\n")
	err = rd.initialize()
	if err != nil {
		wbgo.Error.Printf("Error on initialize dimmer pane: %s", err)
	}
}
