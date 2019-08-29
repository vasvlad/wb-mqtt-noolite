package main

import (
	"strconv"

	"github.com/evgeny-boger/wbgo"
)

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

func (d *Desk) AddRange(id, title string, defaultValue byte, isWritable bool) (wbgo.Control, error) {
	controlArgs := wbgo.NewControlArgs().
		SetId(id).
		SetDescription(title).
		SetType("range").
		SetWritable(isWritable).
		SetMax(255).
		SetRawValue(strconv.Itoa(int(defaultValue)))
	return d.addControl(id, controlArgs)
}
