package main

import (
	"strconv"

	"github.com/contactless/wb-mqtt-noolite/noolite"
	"github.com/evgeny-boger/wbgo"
)

type DimmerDesk struct {
	Desk
	addr  [4]byte
	ch    byte
	on    bool
	level byte
}

func (rd *DimmerDesk) initialize() error {
	rd.updateStatus()
	_, err := rd.AddSwitch("power", "State", rd.on, true)
	if err != nil {
		return err
	}
	_, err = rd.AddRange("level", "Level", 255, true)
	rd.events["power"] = rd.toogle
	rd.events["level"] = rd.changeLevel
	rd.d.loops = append(rd.d.loops, rd.updateStatus)
	return rd.switchToDimmer()
}

func (rd *DimmerDesk) switchToDimmer() error {
	/*
	https://github.com/SergejPr/NooLite-F/issues/1#issuecomment-367813106

	CMD:129 FM:16 D0:xx D1:0 D2:127 D3:0

	bits for D0:
	0: save the module state when power is down
	1: dimmer mode
	2: allow accept noolite commands
	3-4: extra button configuration (00 - switch mode, 01 - button, 10 - breaker??, 11 - disable extra button)
	5: state after power up (has effect only if the saving of module state is disabled)
	6: retranslation of the  nooLite commands
	7: ?
	*/
	req := new(noolite.Request)
	req.Ch = 1
	req.Ctr = 8
	req.Mode = noolite.NooLiteFTX
	req.Cmd = 129
	req.Fmt = 16
	req.D0 = 22
	req.D2 = 127
	req.ID0 = rd.addr[0]
	req.ID1 = rd.addr[1]
	req.ID2 = rd.addr[2]
	req.ID3 = rd.addr[3]
	err := rd.d.conn.Write(req)
	if err != nil {
		return err
	}
	_, err = rd.d.conn.Read()
	if err != nil {
		return err
	}
	return nil
}

func (rd *DimmerDesk) changeLevel(e wbgo.ControlOnValueEvent) {
	newLevel, err := strconv.Atoi(e.RawValue)
	if err != nil {
		wbgo.Error.Printf("Error on parse new level: %s", err)
		return
	}
	req := new(noolite.Request)
	req.Ch = 1
	req.Ctr = 8
	req.Mode = noolite.NooLiteFTX
	req.Cmd = 6
	req.ID0 = rd.addr[0]
	req.ID1 = rd.addr[1]
	req.ID2 = rd.addr[2]
	req.ID3 = rd.addr[3]
	req.D0 = byte(newLevel)
	err = rd.d.conn.Write(req)
	if err != nil {
		wbgo.Error.Printf("Error on sedn command to noolite-f relay: %s", err)
		return
	}
	_, err = rd.d.conn.Read()
	if err != nil {
		wbgo.Error.Printf("Error on sedn command to noolite-f relay: %s", err)
		return
	}
	rd.level = byte(newLevel)
}

func (rd *DimmerDesk) toogle(e wbgo.ControlOnValueEvent) {
	var cmd byte
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
	req.D3 = rd.level
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
	return

}

func (rd *DimmerDesk) updateStatus() {
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
	oldLevel := rd.level
	rd.level = resp.D3
	if oldLevel != rd.level {
		ctrl, ok := rd.controllers["level"]
		if !ok {
			return
		}
		err = rd.d.Access(func(tx wbgo.DriverTx) error {
			ctrl.SetTx(tx)
			err = ctrl.SetOnValue(strconv.Itoa(int(rd.level)))()
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			wbgo.Error.Printf("Error on update value: %s\n", err)
		}
	}
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
