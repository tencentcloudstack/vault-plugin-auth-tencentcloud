package clients

import (
	"fmt"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	"testing"
)

func TestCAMClient_GetRoleName(t *testing.T) {
	secretId := "xxxx"
	secretKey := "xxxx"
	token := "xxxx"

	creds, err := ChainedCredsToCli(secretId, secretKey, token)
	if err != nil {
		fmt.Printf("错误信息,%v", err)
	}
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	profile.HttpProfile.ReqTimeout = 90
	client, err := cam.NewClient(creds, regions.Ashburn, profile)
	if err != nil {
		fmt.Printf("错误信息,%v", err)
	}
	type fields struct {
		client *cam.Client
	}
	type args struct {
		roleId string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRoleName string
		wantErr      bool
	}{
		{
			name: "TestCAMClient_GetRoleName",
			fields: fields{
				client: client,
			},
			args: args{
				roleId: "4611686028425447636",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CAMClient{
				client: tt.fields.client,
			}
			gotRoleName, err := c.GetRoleName(tt.args.roleId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoleName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRoleName != tt.wantRoleName {
				t.Errorf("GetRoleName() gotRoleName = %v, want %v", gotRoleName, tt.wantRoleName)
			}
		})
	}
}
