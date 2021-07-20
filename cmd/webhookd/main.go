package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/cdarne/webhookd/pkg/signature"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var caCert = flag.String("ca-cert", "", "CA certificate path.")
var serverCert = flag.String("server-cert", "", "Server certificate path.")
var serverKey = flag.String("server-key", "", "Server key path.")
var listenAddr = flag.String("listen-addr", "127.0.0.1:8080", "Listen address and port.")
var sharedSecret = flag.String("shared-secret", "", "Shared secret used to verify HMAC signatures.")

func handler(logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		hmacSignature := r.Header.Get("X-Shopify-Hmac-Sha256")
		logger.Printf("Headers: %+v\n", r.Header)
		logger.Println("HMAC signature", hmacSignature)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Printf("Unable to read request body: %s\n", err)
			return
		}

		if signature.ValidSignature(body, *sharedSecret, hmacSignature) {
			logger.Println("Valid webhook signature :)")
		} else {
			logger.Printf("Invalid webhook signature :(")
		}
	})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	logger := log.New(os.Stdout, "webhookd: ", log.LstdFlags)
	logger.Println("Server is starting...")
	server := setupServer(logger)

	startServer(logger, server)
	logger.Println("Server is ready to handle requests at", *listenAddr)

	<-ctx.Done()
	// stop handling the Interrupt signal. This restores the default go behaviour (exit) in case of a second Interrupt
	stop()

	logger.Println("Server is shutting down")
	if err := shutdownServer(server); err != nil {
		logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	logger.Println("Server stopped")
}

func setupServer(logger *log.Logger) *http.Server {
	server := &http.Server{
		Addr:         *listenAddr,
		Handler:      handler(logger),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if useSSL() {
		tlsConfig, err := setupTLS(*serverCert, *serverKey, *caCert)
		if err != nil {
			logger.Fatalln(err)
		}
		server.TLSConfig = tlsConfig
	}

	return server
}

func startServer(logger *log.Logger, server *http.Server) {
	go func() {
		var err error
		if useSSL() {
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logger.Fatalln(err)
		}
	}()
}

func shutdownServer(server *http.Server) error {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	return server.Shutdown(ctxShutDown)
}

func useSSL() bool {
	return *caCert != "" && *serverCert != "" && *serverKey != ""
}

func setupTLS(certFile, keyFile, CAFile string) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = make([]tls.Certificate, 1)
	tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(CAFile)
	if err != nil {
		return nil, err
	}
	ca := x509.NewCertPool()
	ok := ca.AppendCertsFromPEM(b)
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate: %q", CAFile)
	}
	tlsConfig.ClientCAs = ca
	return tlsConfig, nil
}
