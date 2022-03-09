package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type serverLnd struct {
	lndClient lnrpc.LightningClient
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func createServer(client lnrpc.LightningClient) (*http.Server, error) {
	s := serverLnd{
		lndClient: client,
	}

	r := mux.NewRouter()
	r.HandleFunc("/getinvoice", s.GetInvoiceHandler).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/subscribetx", s.SubscribeTxHandler)

	httpServer := &http.Server{
		Addr:    "0.0.0.0:8081",
		Handler: r,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			fmt.Println(err.Error())
		}
	}()

	return httpServer, nil
}

func (s *serverLnd) GetInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	fmt.Println("Received get invoice")

	resp, err := s.lndClient.AddInvoice(context.Background(), &lnrpc.Invoice{
		Value: 1000,
		IsAmp: true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	resp.GetPaymentRequest()

	// TODO
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.GetPaymentRequest()))
}

func (s *serverLnd) SubscribeTxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	subSocket, err := s.lndClient.SubscribeInvoices(context.Background(), &lnrpc.InvoiceSubscription{})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("listening to socket")

	for {
		inv, err := subSocket.Recv()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("received invoice")
		if inv.GetState() == lnrpc.Invoice_SETTLED {
			fmt.Println("paid invoice success")
			err = connection.WriteMessage(1, []byte("true"))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
