package api_json

import (
	"context"

	ajt "github.com/myfantasy/api_json_types"
	"github.com/myfantasy/mft"
)

const (
	ServiceApiObjectType ajt.ObjectType = "service_api"
	PingAction           ajt.Action     = "ping"
)

type ServiceApi struct {
}

func (sa *ServiceApi) AllowedCommands() []ajt.CommandDescription {
	res := []ajt.CommandDescription{
		{
			ObjectType:  ServiceApiObjectType,
			Action:      PingAction,
			Description: "Call service with empty answer",
		},
	}
	return res
}
func (sa *ServiceApi) DoRequest(ctx context.Context, req *ajt.CommandRequest) *ajt.CommandResponce {
	var resp ajt.CommandResponce
	if req.ObjectType == ServiceApiObjectType && req.Action == PingAction {
		return &resp
	}

	resp.Error = mft.GenerateError(20400400, ServiceApiObjectType, PingAction)
	return &resp
}

func Ping(ctx context.Context, api ajt.Api) *mft.Error {
	var req ajt.CommandRequest
	req.ObjectType = ServiceApiObjectType
	req.Action = PingAction

	resp := api.DoRequest(ctx, &req)

	return resp.Error
}
func PingAsUser(ctx context.Context, api ajt.Api, user string) *mft.Error {
	var req ajt.CommandRequest
	req.ObjectType = ServiceApiObjectType
	req.Action = PingAction
	req.User = user

	resp := api.DoRequest(ctx, &req)

	return resp.Error
}
