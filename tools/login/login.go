package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/tools"
)

func main() {
	region := os.Getenv("REGION")                // ex. 'na-ashburn'
	roleName := os.Getenv("ROLE_NAME")           // ex. 'firingrole001'
	vaultAddr := os.Getenv("VAULT_ADDR")         // ex. 'http://127.0.0.1:8200'
	sid := os.Getenv("TENCENTCLOUD_SECRET_ID")   // ex. 'xxx'
	skey := os.Getenv("TENCENTCLOUD_SECRET_KEY") // ex. 'xxx'
	token := os.Getenv("TENCENTCLOUD_TOKEN")     // ex. 'xxx'
	if region == "" {
		panic("REGION must be set")
	}
	if roleName == "" {
		panic("ROLE_NAME must be set")
	}
	if vaultAddr == "" {
		panic("VAULT_ADDR must be set")
	}
	if sid == "" {
		panic("sid must be set")
	}
	if skey == "" {
		panic("skey must be set")
	}
	if token == "" {
		panic("token must be set")
	}

	loginData := tools.GenerateLoginDataV2(roleName, region, sid, skey, token)
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
