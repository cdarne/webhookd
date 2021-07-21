package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Server struct {
	http   *http.Server
	logger *log.Logger
}

func New(listenAddr string, handler http.Handler, logger *log.Logger) *Server {
	return &Server{
		http: &http.Server{
			Addr:         listenAddr,
			Handler:      handler,
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

// ClientError is an error whose details to be shared with client.
type ClientError interface {
	Error() string
	Status() int
}

// HTTPError implements ClientError interface.
type HTTPError struct {
	error
	StatusCode int
}

func (e *HTTPError) Status() int {
	return e.StatusCode
}

func NewHTTPError(err error, statusCode int) *HTTPError {
	return &HTTPError{
		error:      err,
		StatusCode: statusCode,
	}
}

type handlerWithError func(http.ResponseWriter, *http.Request) error

func (h handlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var status int
	clientError, ok := err.(ClientError)
	if !ok {
		status = http.StatusInternalServerError
	} else {
		status = clientError.Status()
	}
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
