package server

import (
	"bytes"
	"fmt"
	"github.com/cdarne/webhookd/pkg/signature"
	"io"
	"io/ioutil"
	"net/http"
)

func VerifySignature(sharedSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hmacSignature := r.Header.Get("X-Shopify-Hmac-Sha256")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error while reading the request body: %s", err), http.StatusBadRequest)
			return
		}
		if signature.ValidSignature(body, sharedSecret, hmacSignature) {
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid HMAC signature", http.StatusUnauthorized)
		}
	})
}
