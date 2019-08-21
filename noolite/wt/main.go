package main

import "github.com/contactless/wb-mqtt-noolite/noolite"
import "github.com/evgeny-boger/wbgo"
import "log"
import "os"
import "strconv"

func main() {
	wbgo.SetDebugLogger(log.New(os.Stdout, "NOOLITE:", log.LstdFlags), true)
	wbgo.SetDebuggingEnabled(true)
	wbgo.Debug.Println("Started")
	c, err := noolite.NewConnection(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer c.Close()
	req := new(noolite.Request)
	//resp := new(noolite.Response)
	ich, _ := strconv.Atoi(os.Args[3])
	req.Ch = byte(ich)
	switch os.Args[2] {
	case "bind_tx":
		req.Mode = 2
		req.Cmd = 15
		err = c.Write(req)
		if err != nil {
			panic(err)
		}
		_, err = c.Read()
		if err != nil {
			panic(err)
		}
	case "bind_rx":
		req.Mode = 3
		req.Ctr = 3
		err = c.Write(req)
		if err != nil {
			panic(err)
		}
		_, err = c.Read()
		if err != nil {
			panic(err)
		}
	case "toogle":
		req.Mode = 2
		req.Cmd = 4
		err = c.Write(req)
		if err != nil {
			panic(err)
		}
		_, err = c.Read()
		if err != nil {
			panic(err)
		}
	case "minute":
		req.Mode = 2
		req.Cmd = 25
		req.Fmt = 5
		req.D0 = 60 / 5
		err = c.Write(req)
		if err != nil {
			panic(err)
		}
		_, err = c.Read()
		if err != nil {
			panic(err)
		}
	case "read_state":
		req.Mode = 2
		req.Cmd = 128
		err = c.Write(req)
		if err != nil {
			panic(err)
		}
		_, err = c.Read()
		if err != nil {
			panic(err)
		}
	case "read":
		for {
			c.Read()
		}
	}

}
