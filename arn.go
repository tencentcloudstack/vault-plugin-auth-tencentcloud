package vault_plugin_auth_tencentcloud

import (
	"errors"
	"fmt"
	"strings"
)

const (
	arnRoleType        arnType = iota // roleName
	arnAssumedRoleType                // assumed-role
)

const (
	roleName    = "roleName"
	assumedRole = "assumed-role"
)

type arnType int

// toString
func (t arnType) String() string {
	switch t {
	case arnRoleType:
		return roleName
	case arnAssumedRoleType:
		return assumedRole
	default:
		return ""
	}
}

type arn struct {
	Uin      string
	RoleName string
	RoleId   string
	Full     string
	Type     arnType
}

// check member
func (a *arn) IsMemberOf(possibleParent *arn) bool {
	if possibleParent.Type != arnRoleType && possibleParent.Type != arnAssumedRoleType {
		return false
	}
	if possibleParent.Uin != a.Uin {
		return false
	}
	if possibleParent.RoleName != a.RoleName {
		return false
	}
	return true
}

func parseARN(a string) (*arn, error) {
	// camArn should look like one of the following:
	// 1. qcs::cam::uin/<uin>:roleName/<RoleName>
	// 2. qcs::sts:<uin>:assumed-role/<RoleId>
	// if we get something like 2, then we want to transform that back to what
	// most people would expect, which is qcs::cam::uin/<uin>:roleName/<RoleName>
	if a == "" {
		return nil, fmt.Errorf("no arn provided")
	}
	parsed := &arn{
		Full: a,
	}
	outerFields := strings.Split(a, ":")
	if len(outerFields) != 6 && len(outerFields) != 5 {
		return nil, fmt.Errorf("unrecognized arn: contains %d colon-separated fields, expected 6 or 5", len(outerFields))
	}
	if outerFields[0] != "qcs" {
		return nil, errors.New(`unrecognized arn: does not begin with "qcs:"`)
	}
	if outerFields[2] != "cam" && outerFields[2] != "sts" {
		return nil, fmt.Errorf("unrecognized service: %v, not cam or sts", outerFields[2])
	}
	if outerFields[2] == "cam" {
		uinFields := strings.Split(outerFields[4], "/")
		if len(uinFields) < 2 {
			return nil, fmt.Errorf("unrecognized arn: %q contains fewer than 2 slash-separated uinFields", outerFields[4])
		}
		parsed.Uin = uinFields[1]
		roleFiles := strings.Split(outerFields[5], "/")
		if len(roleFiles) == 2 {
			parsed.Type = arnRoleType
			if roleFiles[0] == roleName {
				parsed.RoleName = roleFiles[1]
			} else {
				return nil, errors.New("the caller's arn does not match the role's arn")
			}
		} else {
			return nil, fmt.Errorf("unrecognized arn: %q contains fewer than 2 slash-separated roleFiles", outerFields[4])
		}
	} else if outerFields[2] == "sts" {
		parsed.Uin = outerFields[3]
		roleFiles := strings.Split(outerFields[4], "/")
		if len(roleFiles) == 2 {
			parsed.Type = arnAssumedRoleType
			if roleFiles[0] == assumedRole {
				parsed.RoleId = roleFiles[1]
			} else {
				return nil, errors.New("the caller's arn does not match the role's arn")
			}
		} else {
			return nil, fmt.Errorf("unrecognized arn: %q contains fewer than 2 slash-separated roleFiles", outerFields[4])
		}
	}
	return parsed, nil
}

// toString
func (a *arn) String() string {
	return a.Full
}
