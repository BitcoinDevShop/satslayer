package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
)

type serverLnd struct {
	lndClient routerrpc.RouterClient
}

func createServer(client routerrpc.RouterClient) (*http.Server, error) {
	s := serverLnd{
		lndClient: client,
	}

	r := mux.NewRouter()
	r.HandleFunc("/sendpayment", s.SendPaymentHandler).Methods(http.MethodPost, http.MethodOptions)
	//http.Handle("/", r)
	//http.ListenAndServe(":8000", r)

	httpServer := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: r,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			fmt.Println(err.Error())
		}
	}()

	return httpServer, nil
}

type SendPayment struct {
	PullInterval int    `json:"pull_interval"`
	Invoice      string `json:"invoice"`
}

func (s *serverLnd) SendPaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	// get body
	var sendPaymentRequest SendPayment
	err := json.NewDecoder(r.Body).Decode(&sendPaymentRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	sendCtx, _ := context.WithTimeout(context.Background(), time.Minute*2)
	//defer cancel()

	resp, err := s.lndClient.SendPaymentV2(sendCtx, &routerrpc.SendPaymentRequest{
		PaymentRequest: sendPaymentRequest.Invoice,
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
				break
			}
			fmt.Println("payment in progress...")

			if payResp.Status == lnrpc.Payment_SUCCEEDED {
				fmt.Println("Payment succeeded")
				break
			}
		}
	}()

	fmt.Println("Received send payment request")

	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("success"))
}
