package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/evgeny-boger/wbgo"
)

func main() {
	serial := flag.String("serial", "/dev/ttyUSB0", "serial port address")
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker url")
	debug := flag.Bool("debug", false, "Enable debugging")
	flag.Parse()
	if *debug {
		wbgo.SetDebuggingEnabled(true)
	}
	client := wbgo.NewPahoMQTTClient(*broker, "noolite-f")
	driverArgs := wbgo.NewDriverArgs()
	driverArgs.SetId("noolite-f")
	driverArgs.SetMqtt(client)
	driverArgs.SetReownUnknownDevices(true)
	driver, err := NewDriver(driverArgs, *serial)
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
