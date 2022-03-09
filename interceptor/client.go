package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

func CreateLNDClient(macaroonPath, macaroonHex, tlsCertPath, ipAddr string) (lnrpc.LightningClient, error) {
	//tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	//macaroonPath := path.Join(usr.HomeDir, ".lnd/admin.macaroon")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fmt.Println("Cannot get node tls credentials", err)
		return nil, err
	}

	var macaroonData []byte
	if macaroonHex != "" {
		macBytes, err := hex.DecodeString(macaroonHex)
		if err != nil {
			return nil, err
		}
		macaroonData = macBytes
	} else if macaroonPath != "" {
		macBytes, err := ioutil.ReadFile(macaroonPath)
		if err != nil {
			return nil, err
		}
		macaroonData = macBytes // make it available outside of the else if block
	} else {
		return nil, fmt.Errorf("LND macaroon is missing")
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macaroonData); err != nil {
		return nil, err
	}

	/*
		macaroonBytes, err := ioutil.ReadFile(macaroonPath)
		if err != nil {
			fmt.Println("Cannot read macaroon file", err)
			return nil, err
		}

		mac := &macaroon.Macaroon{}
		if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
			fmt.Println("Cannot unmarshal macaroon", err)
			return nil, err
		}
	*/

	macCred, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macCred),
	}

	conn, err := grpc.Dial(ipAddr, opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return nil, err
	}
	client := lnrpc.NewLightningClient(conn)

	return client, nil
}

func CreateLNDPaymentClient(macaroonPath, macaroonHex, tlsCertPath, ipAddr string) (routerrpc.RouterClient, error) {
	// Hardcoded paths

	//tlsCertPath := path.Join(usr.HomeDir, ".lnd/tls.cert")
	//macaroonPath := path.Join(usr.HomeDir, ".lnd/admin.macaroon")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fmt.Println("Cannot get node tls credentials", err)
		return nil, err
	}

	var macaroonData []byte
	if macaroonHex != "" {
		macBytes, err := hex.DecodeString(macaroonHex)
		if err != nil {
			return nil, err
		}
		macaroonData = macBytes
	} else if macaroonPath != "" {
		macBytes, err := ioutil.ReadFile(macaroonPath)
		if err != nil {
			return nil, err
		}
		macaroonData = macBytes // make it available outside of the else if block
	} else {
		return nil, fmt.Errorf("LND macaroon is missing")
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macaroonData); err != nil {
		return nil, err
	}

	/*
		macaroonBytes, err := ioutil.ReadFile(macaroonPath)
		if err != nil {
			fmt.Println("Cannot read macaroon file", err)
			return nil, err
		}

		mac := &macaroon.Macaroon{}
		if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
			fmt.Println("Cannot unmarshal macaroon", err)
			return nil, err
		}
	*/

	macCred, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macCred),
	}

	conn, err := grpc.Dial(ipAddr, opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return nil, err
	}
	client := routerrpc.NewRouterClient(conn)

	return client, nil
}

func createGrpcInterceptor(client lnrpc.LightningClient) error {

	go func() {
		ctx, _ := context.WithTimeout(context.Background(), time.Minute*5)
		rpcMiddlewareClient, err := client.RegisterRPCMiddleware(ctx)
		if err != nil {
			panic(err)
		}

		fmt.Println("Created middleware stream")

		// Register interceptor immediately
		err = rpcMiddlewareClient.Send(&lnrpc.RPCMiddlewareResponse{
			MiddlewareMessage: &lnrpc.RPCMiddlewareResponse_Register{
				Register: &lnrpc.MiddlewareRegistration{
					MiddlewareName:           "subscribe",
					CustomMacaroonCaveatName: "subscribe",
					ReadOnlyMode:             false,
				},
			},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println("Registered middleware stream")

		// Listen for responses
		fmt.Println("Listening to middleware stream")
		go func() {
			for {
				resp, err := rpcMiddlewareClient.Recv()
				if err != nil {
					panic(err)
				}
				fmt.Println("Got middleware response")
				fmt.Println(resp)

				fmt.Println("sending response")
				err = rpcMiddlewareClient.Send(&lnrpc.RPCMiddlewareResponse{
					RefMsgId: resp.GetMsgId(),
					MiddlewareMessage: &lnrpc.RPCMiddlewareResponse_Feedback{
						Feedback: &lnrpc.InterceptFeedback{
							Error:           "",
							ReplaceResponse: false,
						},
					},
				})
				if err != nil {
					panic(err)
				}
				fmt.Println("Sent grpc response")
			}
		}()
	}()

	time.Sleep(time.Duration(time.Second * 1))

	return nil
}

func bakeMacaroon(client lnrpc.LightningClient) (string, error) {
	ctx := context.Background()
	//secret := uint64(123)

	// TODO real custom caveat
	resp, err := client.BakeMacaroon(ctx, &lnrpc.BakeMacaroonRequest{
		//RootKeyId:                secret,
		AllowExternalPermissions: true,
		Permissions: []*lnrpc.MacaroonPermission{
			/*
				{
					Entity: "info",
					Action: "read",
				},
			*/
			{
				Entity: "offchain",
				Action: "read",
			},
			{
				Entity: "offchain",
				Action: "write",
			},
		},
	})
	if err != nil {
		return "", err
	}

	// Parse the mac to add a caveat to it
	freshMac := resp.GetMacaroon()
	freshMacBytes, err := hex.DecodeString(freshMac)
	if err != nil {
		return "", err
	}
	fmt.Println("Got bytes for mac")

	formattedMac := macaroon.Macaroon{}
	err = formattedMac.UnmarshalBinary(freshMacBytes)
	if err != nil {
		fmt.Println("Could not unmarshall mac")
		return "", err
	}
	fmt.Println("Unmarshalled mac")
	fmt.Println(formattedMac.Id())

	fmt.Println("Going to add caveats")
	formattedMac.AddFirstPartyCaveat([]byte("lnd-custom subscribe"))
	//formattedMac.AddFirstPartyCaveat([]byte("sats 1000"))
	//formattedMac.AddFirstPartyCaveat([]byte("iteration 10s"))

	fmt.Println(formattedMac.Caveats())

	reformattedMac, err := formattedMac.MarshalBinary()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(reformattedMac), nil
	//return freshMac, nil
}
