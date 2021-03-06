package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/cdarne/webhookd/internal/server"
	"github.com/cdarne/webhookd/internal/subprocess"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	listenAddr := flag.String("listen-addr", "127.0.0.1:8080", "Listen address and port.")
	sharedSecret := flag.String("shared-secret", "", "Shared secret used to verify HMAC signatures.")

	concurrency := flag.Int("concurrency", runtime.NumCPU(), "Number of program instances that are allowed to run concurrently.")
	queueSize := flag.Int("queue-size", 1000, "The length of the queue of webhooks waiting to be processed.")

	caCert := flag.String("ca-cert", "", "CA certificate path.")
	serverCert := flag.String("server-cert", "", "Server certificate path.")
	serverKey := flag.String("server-key", "", "Server key path.")

	flag.Parse()

	logger := log.New(os.Stdout, "webhookd: ", log.LstdFlags)

	args := flag.Args()
	if len(args) < 1 {
		logger.Fatalln("Missing command argument. Aborting...")
	}
	command := args[0]
	commandArgs := args[1:]

	logger.Println("Server is starting...")

	runner := subprocess.NewRunner(logger, *concurrency, *queueSize)
	runner.Start()

	handler := server.Logging(logger, server.VerifySignature(*sharedSecret, server.SpawnProcess(command, commandArgs, runner)))
	server := server.New(*listenAddr, handler, logger)
	if useSSL(*serverCert, *serverKey, *caCert) {
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
	runner.Stop()
	logger.Println("Server stopped")
}

func useSSL(serverCert, serverKey, caCert string) bool {
	return serverCert != "" && serverKey != "" && caCert != ""
}
