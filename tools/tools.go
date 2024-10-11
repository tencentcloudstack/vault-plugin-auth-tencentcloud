package tools

import (
	"fmt"
	"net/http"
	"net/url"
)

// GenerateLoginDataV2 Generates the necessary data to send to the Vault server for generating a token.
func GenerateLoginDataV2(role, region, secretId, secretKey, token string) map[string]interface{} {
	return map[string]interface{}{
		"role":       role,
		"region":     region,
		"secret_id":  secretId,
		"secret_key": secretKey,
		"token":      token,
	}
}

/*
RequestCapturer fulfills the Proxy method of http.Transport, so can be used to replace
the Proxy method on any transport method to simply capture the request.
Its Proxy method always returns an error so the request won't actually be fired.
This is useful for quickly finding out what final request a client is sending.
*/
type RequestCapturer struct {
	request *http.Request
}

// Proxy
func (r *RequestCapturer) Proxy(req *http.Request) (*url.URL, error) {
	r.request = req
	return nil, fmt.Errorf("throwing an error so we won't actually execute the request")
}

// GetCapturedRequest
func (r *RequestCapturer) GetCapturedRequest() (*http.Request, error) {
	if r.request == nil {
		return nil, fmt.Errorf("no request captured")
	}
	return r.request, nil
}
