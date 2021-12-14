package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/clients"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/tools"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

func main() {
	region := os.Getenv("REGION")        // ex. 'na-ashburn'
	roleName := os.Getenv("ROLE_NAME")   // ex. 'firingrole001'
	vaultAddr := os.Getenv("VAULT_ADDR") // ex. 'http://127.0.0.1:8200'
	if region == "" {
		panic("REGION must be set")
	}
	if roleName == "" {
		panic("ROLE_NAME must be set")
	}
	if vaultAddr == "" {
		panic("VAULT_ADDR must be set")
	}
	// can get token Provider
	credentialChain := []common.Provider{
		clients.DefaultEnvProvider(),
	}
	creds, err := common.NewProviderChain(credentialChain).GetCredential()
	if err != nil {
		panic(err)
	}
	loginData, err := tools.GenerateLoginData(roleName, creds, region)
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(loginData)
	if err != nil {
		panic(err)
	}

	loginReq, err := http.NewRequest(http.MethodPost, vaultAddr+"/v1/auth/tencentcloud/login", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(loginReq)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("response status code: %d\n", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s", body)

}
