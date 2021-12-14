package vault_plugin_auth_tencentcloud

import (
	"testing"
)

func TestParseRoleArn(t *testing.T) {
	// qcs::cam::uin/100021543***:roleName/****
	arn := "qcs::cam::uin/1000215438890:roleName/elk"
	result, err := parseARN(arn)
	if err != nil {
		t.Fatal(err)
	}
	if result.Uin != "1000215438890" {
		t.Fatalf("got %s but expected %s", result.Uin, "1000215438890")
	}
	if result.Type != arnRoleType {
		t.Fatalf("got %d but expected %d", result.Type, arnRoleType)
	}
	if result.RoleName != "elk" {
		t.Fatalf("got %s but wanted %s", result.RoleName, "elk")
	}
	if result.RoleId != "" {
		t.Fatalf("got %s but wanted %s", result.RoleId, "")
	}
}

func TestParseAssumedRoleArn(t *testing.T) {
	// qcs::sts:1000262***:assumed-role/461168601842741***
	arn := "qcs::sts:1000215438890:assumed-role/4611686018427418890"
	result, err := parseARN(arn)
	if err != nil {
		panic(err)
	}
	if result.Uin != "1000215438890" {
		t.Fatalf("got %s but expected %s", result.Uin, "1000215438890")
	}
	if result.Type != arnAssumedRoleType {
		t.Fatalf("got %d but expected %d", result.Type, arnAssumedRoleType)
	}
	if result.RoleName != "" {
		t.Fatalf("got %s but wanted %s", result.RoleName, "")
	}
	if result.RoleId != "4611686018427418890" {
		t.Fatalf("got %s but wanted %s", result.RoleId, "4611686018427418890")
	}
}

func TestParseEmpty(t *testing.T) {
	arn := ""
	_, err := parseARN(arn)
	if err == nil {
		t.Fatal("expected an err")
	}
}
