package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting website server")
	adminMacaroonPath := "/Users/a/.polar/networks/1/volumes/lnd/dave/data/chain/bitcoin/regtest/admin.macaroon"
	tlsCertPath := "/Users/a/.polar/networks/1/volumes/lnd/dave/tls.cert"
	ipAddr := "127.0.0.1:10004"
	client, err := CreateLNDClient(adminMacaroonPath, "", tlsCertPath, ipAddr)
	if err != nil {
		panic(err)
	}

	// Create website server
	httpServer, err := createServer(client)
	if err != nil {
		panic(err)
	}

	// Daemon wait until cancel
	c := make(chan os.Signal)
	done := make(chan bool, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		httpServer.Shutdown(context.Background())
		done <- true
	}()
	fmt.Println("Started...")
	<-done
	fmt.Println("Closing...")
}
