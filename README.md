# satslayer

## Prereq

This relies on macaroon interceptors, which have to be turned on the user's LND config. 

```
  --rpcmiddleware.enable
```

Run that when you launch LND, or hit advanced options in Polar for the node, prefill with default options, then add that line somewhere.

## User side server

User side interceptor server, runs the API's for sending payment and the interceptor daemon for LND macaroon requests.

### API 

API: 

POST localhost:8080/sendpayment

Body not implemented yet

### Running

Change hard coded lines: 

interceptor/client.go 

```
	// Hardcoded paths
	tlsCertPath := "/Users/a/.polar/networks/1/volumes/lnd/carol/tls.cert"
	ipAddr := "127.0.0.1:10003"
```

interceptor/main.go

```
	adminMacaroonPath := "/Users/a/.polar/networks/1/volumes/lnd/carol/data/chain/bitcoin/regtest/admin.macaroon"
```

```
cd interceptor
go run *.go
```

## Website side server

Website side server, runs the API's for getting an invoice and getting notified of paid transactions. Also runs the code for pulling payments from users.

### API

API: 

#### Get invoice

GET localhost:8081/getinvoice
Returns invoice string

#### Subscribe transaction

TODO TURN TO WEBSOCKET

POST localhost:8081/subscribetx
Returns nothing *yet*


### Running

```
cd backend
go run *.go
```
