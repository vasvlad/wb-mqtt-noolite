package main

import (
	"github.com/vasvlad/wb-mqtt-noolite"
	"github.com/contactless/wbgong"

)

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
	rd.on = !rd.on
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
