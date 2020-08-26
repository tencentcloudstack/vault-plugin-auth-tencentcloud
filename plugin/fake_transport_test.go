package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/custom"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

type fakeTransport struct {
	arn         Arn
	accountId   string
	userId      string
	principalId string
	userType    string
}

func (f fakeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	action := request.Header.Get("X-TC-Action")

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)

	switch action {
	case "GetCallerIdentity":
		response := custom.NewGetCallerIdentityResponse()

		response.Response = &struct {
			Arn *string `json:"Arn,omitempty" name:"Arn"`

			AccountId *string `json:"AccountId,omitempty" name:"AccountId"`

			UserId *string `json:"UserId,omitempty" name:"UserId"`

			PrincipalId *string `json:"PrincipalId,omitempty" name:"PrincipalId"`

			Type *string `json:"Type,omitempty" name:"Type"`

			RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
		}{
			Arn:         common.StringPtr(string(f.arn)),
			AccountId:   &f.accountId,
			UserId:      &f.userId,
			PrincipalId: &f.principalId,
			Type:        &f.userType,
			RequestId:   common.StringPtr("test-111-2222-33333-444444"),
		}

		respBytes, err := json.Marshal(response)
		if err != nil {
			return nil, err
		}

		_, _ = recorder.Write(respBytes)
	default:
		return nil, fmt.Errorf("unknown action %s", action)
	}

	return recorder.Result(), nil
}
