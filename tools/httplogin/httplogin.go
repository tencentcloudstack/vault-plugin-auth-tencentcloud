package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/tools"
	"io/ioutil"
	"log"
	"net/http"
)

/*
Test auth through http api

	curl \
	 --request POST \
	 --data {"region":"","role_name":"","vault_addr":"","secret_id":"","secret_key":"","token":""} \
	 http://127.0.0.1:8088/login
*/
func LoginServer(w http.ResponseWriter, req *http.Request) {
	s, _ := ioutil.ReadAll(req.Body)
	m := make(map[string]string)
	json.Unmarshal(s, &m)
	region := m["region"]
	roleName := m["role_name"]
	vaultAddr := m["vault_addr"]
	secretId := m["secret_id"]
	secretKey := m["secret_key"]
	token := m["token"]
	loginData := tools.GenerateLoginDataV2(roleName, region, secretId, secretKey, token)
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
	roleName := m["role_name"]
	secretId := m["secret_id"]
	secretKey := m["secret_key"]
	token := m["token"]
	loginData := tools.GenerateLoginDataV2(roleName, region, secretId, secretKey, token)
	data, _ := json.Marshal(loginData)
	w.Write(data)
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
