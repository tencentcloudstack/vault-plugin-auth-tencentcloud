package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/clients"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/tools"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

/*
   Test auth through http api
   curl \
    --request POST \
    --data {"region":"","roleName":"","vaultAddr":"","secret_id":"","secret_key":"","token":""} \
    http://127.0.0.1:8088/login
*/
func LoginServer(w http.ResponseWriter, req *http.Request) {
	s, _ := ioutil.ReadAll(req.Body)
	m := make(map[string]string)
	json.Unmarshal(s, &m)
	region := m["region"]
	roleName := m["roleName"]
	vaultAddr := m["vaultAddr"]
	secretId := m["secret_id"]
	secretKey := m["secret_key"]
	token := m["token"]
	loginData, err := getLoginData(secretId, secretKey, token, region, roleName)
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
	w.Write(body)
}

// GetLoginData
func GetLoginData(w http.ResponseWriter, req *http.Request) {
	s, _ := ioutil.ReadAll(req.Body)
	m := make(map[string]string)
	json.Unmarshal(s, &m)
	region := m["region"]
	roleName := m["roleName"]
	secretId := m["secret_id"]
	secretKey := m["secret_key"]
	token := m["token"]
	loginData, err := getLoginData(secretId, secretKey, token, region, roleName)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(loginData)
	w.Write(data)
}

func getLoginData(secretId string, secretKey string, token string,
	region string, roleName string) (map[string]interface{}, error) {
	credentialChain := []common.Provider{
		clients.NewConfigurationCredentialProvider(
			&clients.Configuration{secretId, secretKey, token}),
	}
	creds, err := common.NewProviderChain(credentialChain).GetCredential()
	if err != nil {
		return nil, err
	}
	return tools.GenerateLoginData(roleName, creds, region)

}
func main() {
	http.HandleFunc("/login", LoginServer)
	http.HandleFunc("/getLoginData", GetLoginData)
	fmt.Println("server start now ！！！！")
	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
