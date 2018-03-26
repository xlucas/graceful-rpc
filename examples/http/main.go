package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"stash.ovh.net/playground/rpc"
)

var clientLogger = log.New(os.Stdout, "client|", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
var serverLogger = log.New(os.Stdout, "server|", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

type Args struct {
	A int
	B int
}

type Operations struct{}

func (op *Operations) Long(args *Args, reply *int) error {
	defer serverLogger.Print("leaving RPC method")
	serverLogger.Print("entering RPC method")
	time.Sleep(5 * time.Second)
	*reply = args.A + args.B
	return nil
}

var (
	clientExit = make(chan struct{}, 1)
	serverExit = make(chan struct{}, 1)
	result     int
)

func main() {
	if err := rpc.Register(&Operations{}); err != nil {
		serverLogger.Fatalf("unable to register methods: %s", err)
	}

	rpc.HandleHTTP()

	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		serverLogger.Fatalf("unable to listen: %s", err)
	}

	go func() {
		if err := http.Serve(lis, nil); err != nil && err != rpc.ErrServerClosed {
			serverLogger.Printf("unable to serve request: %s", err)
		}
	}()

	client, err := rpc.DialHTTP("tcp", ":5000")
	if err != nil {
		clientLogger.Fatalf("unable to dial: %s", err)
	}

	// Call a server's long-running operation
	go func() {
		clientLogger.Print("calling RPC method")
		err := client.Call("Operations.Long", &Args{A: 1, B: 2}, &result)
		if err != nil {
			clientLogger.Fatalf("operation error: %s", err)
		}
		clientExit <- struct{}{}
	}()

	// Trigger server graceful shutdown
	time.Sleep(2 * time.Second)
	go func() {
		serverLogger.Print("calling shutdown")
		if err := rpc.Shutdown(context.Background()); err != nil {
			serverLogger.Fatalf("shutdown failed: %s", err)
		}
		serverExit <- struct{}{}
	}()

	<-clientExit
	<-serverExit

	clientLogger.Printf("RPC method result: %d", result)
}
