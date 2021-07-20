package main

import (
	"context"
	"flag"
	"github.com/cdarne/webhookd/internal/server"
	"log"
	"os"
	"os/signal"
)

var caCert = flag.String("ca-cert", "", "CA certificate path.")
var serverCert = flag.String("server-cert", "", "Server certificate path.")
var serverKey = flag.String("server-key", "", "Server key path.")
var listenAddr = flag.String("listen-addr", "127.0.0.1:8080", "Listen address and port.")
var sharedSecret = flag.String("shared-secret", "", "Shared secret used to verify HMAC signatures.")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	logger := log.New(os.Stdout, "webhookd: ", log.LstdFlags)
	logger.Println("Server is starting...")
	server := server.New(*listenAddr, *sharedSecret, logger)
	if useSSL() {
		err := server.SetupTLS(*serverCert, *serverKey, *caCert)
		if err != nil {
			logger.Fatalln(err)
		}
	}

	server.Start()
	logger.Println("Server is ready to handle requests at", *listenAddr)

	<-ctx.Done()
	// stop handling the Interrupt signal. This restores the default go behaviour (exit) in case of a second Interrupt
	stop()

	logger.Println("Server is shutting down")
	if err := server.Stop(); err != nil {
		logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	logger.Println("Server stopped")
}

func useSSL() bool {
	return *caCert != "" && *serverCert != "" && *serverKey != ""
}
