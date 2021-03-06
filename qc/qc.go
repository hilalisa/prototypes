package main

import (
	"flag"
	"fmt"

	"github.com/progrium/prototypes/libmux/mux"
	"github.com/progrium/prototypes/qrpc"
)

const addr = "localhost:4242"

func main() {

	sess, err := mux.DialWebsocket(addr)
	if err != nil {
		panic(err)
	}
	client := &qrpc.Client{Session: sess}

	flag.Parse()

	var resp string
	var args interface{}
	if flag.Arg(1) != "" {
		args = flag.Arg(1)
	}
	err = client.Call(flag.Arg(0), args, &resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("resp: %#v\n", resp)
}
