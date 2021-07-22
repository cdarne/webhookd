package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cdarne/webhookd/pkg/signature"
)

func VerifySignature(sharedSecret string, next handlerWithError) handlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		hmacSignature := r.Header.Get("X-Shopify-Hmac-Sha256")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error while reading the request body: %s", err)
		}
		if signature.ValidSignature(body, sharedSecret, hmacSignature) {
			// "rewind" the body to be readable by `next`
			r.Body = ioutil.NopCloser(bytes.NewReader(body))
			return next(w, r)
		} else {
			return NewHTTPError(http.StatusUnauthorized, errors.New("invalid HMAC signature"))
		}
	}
}
