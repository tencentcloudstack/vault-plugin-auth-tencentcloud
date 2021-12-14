package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	stsLocal "github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/tencentcloud/sts/v20180813"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

// Generates the necessary data to send to the Vault server for generating a token.
// This is useful for other API clients to use.
// If "" is passed in for SecretId,SecretKey,Token
// attempts to use credentials set as env vars or available through instance metadata.
func GenerateLoginData(role string, creds common.CredentialIface, region string) (map[string]interface{}, error) {
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	capturer := &RequestCapturer{}
	transport := &http.Transport{}
	transport.Proxy = capturer.Proxy
	if region == "" {
		region = regions.Ashburn
	}
	client, err := stsLocal.NewClient(creds, region, profile)
	if err != nil {
		return nil, err
	}
	client.WithHttpTransport(transport)
	client.GetCallerIdentity(stsLocal.NewGetCallerIdentityRequest())
	getCallerIdentityRequest, err := capturer.GetCapturedRequest()
	if err != nil {
		return nil, err
	}
	u := base64.StdEncoding.EncodeToString([]byte(getCallerIdentityRequest.URL.String()))
	b, err := json.Marshal(getCallerIdentityRequest.Header)
	if err != nil {
		return nil, err
	}
	headers := base64.StdEncoding.EncodeToString(b)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"role":                     role,
		"identity_request_url":     u,
		"identity_request_headers": headers,
	}, nil
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
