package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/cdarne/webhookd/internal/subprocess"
	"github.com/cdarne/webhookd/pkg/signature"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Server struct {
	http   *http.Server
	logger *log.Logger
}

func New(listenAddr, sharedSecret string, logger *log.Logger, command string, commandArgs []string) *Server {
	return &Server{
		http: &http.Server{
			Addr:         listenAddr,
			Handler:      handler(sharedSecret, logger, command, commandArgs),
			ErrorLog:     logger,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) SetupTLS(serverCert, serverKey, caCert string) error {
	tlsConfig, err := setupTLS(serverCert, serverKey, caCert)
	if err != nil {
		return err
	}
	s.http.TLSConfig = tlsConfig
	return nil
}

func (s *Server) Start() {
	go func() {
		var err error
		if s.http.TLSConfig != nil {
			err = s.http.ListenAndServeTLS("", "")
		} else {
			err = s.http.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			s.logger.Println(err)
		}
	}()
}

func (s *Server) Stop() error {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.http.SetKeepAlivesEnabled(false)
	return s.http.Shutdown(ctxShutDown)
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

func handler(sharedSecret string, logger *log.Logger, command string, commandArgs []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		hmacSignature := r.Header.Get("X-Shopify-Hmac-Sha256")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Printf("Unable to read request body: %s\n", err)
			return
		}

		if signature.ValidSignature(body, sharedSecret, hmacSignature) {
			logger.Printf("Processing webhook")
			go subprocess.Run(r, command, commandArgs, body, logger)
		} else {
			logger.Printf("Invalid webhook received")
		}
	})
}
