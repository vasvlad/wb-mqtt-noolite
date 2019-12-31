package main

import "github.com/vasvlad/wb-mqtt-noolite"

type NooliteConnection interface {
	Write(req *noolite.Request) error
	Read() (*noolite.Response, error)
	Close()
}
