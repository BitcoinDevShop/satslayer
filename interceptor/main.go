package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
)

func main() {
	adminMacaroonPath := "/Users/a/.polar/networks/1/volumes/lnd/carol/data/chain/bitcoin/regtest/admin.macaroon"
	tlsCertPath := "/Users/a/.polar/networks/1/volumes/lnd/carol/tls.cert"
	ipAddr := "127.0.0.1:10003"
	client, err := CreateLNDClient(adminMacaroonPath, "", tlsCertPath, ipAddr)
	if err != nil {
		panic(err)
	}
	// Use macaroon to start another client that makes a request to LND
	routerClient, err := CreateLNDPaymentClient(adminMacaroonPath, "", tlsCertPath, ipAddr)
	if err != nil {
		panic(err)
	}

	// Connect to user node
	ctx := context.Background()
	getInfoResp, err := client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		fmt.Println("Cannot get info from node:", err)
		panic(err)
	}
	fmt.Println(getInfoResp)

	// Start interceptor
	err = createGrpcInterceptor(client)
	if err != nil {
		panic(err)
	}

	// Create user server
	httpServer, err := createServer(routerClient)
	if err != nil {
		panic(err)
	}

	// Create new macaroon
	// TODO move this to server via pay
	macaroon, err := bakeMacaroon(client)
	if err != nil {
		panic(err)
	}

	fmt.Println("created macaroon")
	fmt.Println(macaroon)

	// Use macaroon to start another client that makes a request to LND
	permissionedClient, err := CreateLNDPaymentClient("", macaroon, tlsCertPath, ipAddr)
	if err != nil {
		panic(err)
	}

	fmt.Println("Created permissioned client")
	sendCtx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	if err != nil {
		panic(err)
	}
	defer cancel()

	for {
		resp, err := permissionedClient.SendPaymentV2(sendCtx, &routerrpc.SendPaymentRequest{
			PaymentRequest: "lnbcrt10u1psuut8app5nnwegkppq3h0a6gwc43qxxe0nuxfqs78hl4szdq3aw0m7cnanwqsdqqcqzpgxq9z0rgqsp5uj2qm8xch2wds3hy27gc8cjttf3thk8zhjdmf2cv534q0jlwfexq9q8pqqqssq4jdsm9f2jf3trc0kvyza2rsd957hfvnkeqwpmlzpr38hkgeh9rwru3afqmtq4pydjg4k5v6ndpz87mqdy3dp4rs8rp3cnhunscxm6hcp7ltf84",
			TimeoutSeconds: 100,
			FeeLimitMsat:   10000000,
		})
		if err != nil {
			panic(err)
		}
		go func() {
			for {
				payResp, err := resp.Recv()
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("payment in progress...")

				if payResp.Status == lnrpc.Payment_SUCCEEDED {
					fmt.Println("Payment succeeded")
					break
				}
			}
		}()
		time.Sleep(2 * time.Second)
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
