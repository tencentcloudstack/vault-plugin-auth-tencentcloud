package cli

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/custom"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

type dumpTransport struct {
	req *http.Request
}

func (d *dumpTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	d.req = request

	return nil, errors.New("no real network api call")
}

func DumpCallerIdentityRequest(accessKey, secretKey, secretToken, region string) (requestUrl string, requestBody []byte, signedHeader http.Header, err error) {
	credential := common.NewCredential(accessKey, secretKey)
	credential.Token = secretToken

	cpf := profile.NewClientProfile()

	apiClient, err := custom.NewClient(credential, region, cpf)
	if err != nil {
		return "", nil, nil, err
	}

	transport := new(dumpTransport)

	apiClient.WithHttpTransport(transport)

	request := custom.NewGetCallerIdentityRequest()

	_, _ = apiClient.GetCallerIdentity(request)

	requestBody, err = ioutil.ReadAll(transport.req.Body)
	if err != nil {
		return "", nil, nil, err
	}

	return request.GetUrl(), requestBody, transport.req.Header, nil
}
