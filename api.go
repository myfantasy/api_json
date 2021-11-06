package api_json

import (
	"context"
	"encoding/json"

	"github.com/myfantasy/authentication"
	"github.com/myfantasy/authorization"
	"github.com/myfantasy/compress"
	"github.com/myfantasy/mfs"
	"github.com/myfantasy/mft"

	ajt "github.com/myfantasy/api_json_types"
)

type GetCompressionFunc func(preferCompress []compress.CompressionType,
) (outCompType compress.CompressionType, outPreferCompress []compress.CompressionType)

type Api struct {
	Compressor *compress.Generator

	ApisMap map[ajt.ObjectType]map[ajt.Action]ajt.Api

	GetCompression GetCompressionFunc

	AuthenticationChecker map[string]authentication.AuthenticationChecker
	PermissionChecker     authorization.PermissionChecker

	mx mfs.RWTMutex
}

func (api *Api) AddApi(a ajt.Api) {
	api.mx.Lock()
	defer api.mx.Unlock()

	if api.ApisMap == nil {
		api.ApisMap = make(map[ajt.ObjectType]map[ajt.Action]ajt.Api, 0)
	}

	for _, v := range a.AllowedCommands() {
		if _, ok := api.ApisMap[v.ObjectType]; !ok {
			api.ApisMap[v.ObjectType] = make(map[ajt.Action]ajt.Api)
		}

		api.ApisMap[v.ObjectType][v.Action] = a
	}
}

func (api *Api) AddAuthenticationChecker(ac authentication.AuthenticationChecker) {
	api.mx.Lock()
	defer api.mx.Unlock()

	if api.AuthenticationChecker == nil {
		api.AuthenticationChecker = make(map[string]authentication.AuthenticationChecker)
	}

	api.AuthenticationChecker[ac.Type()] = ac
}

func (api *Api) RLock(ctx context.Context) bool {
	return api.mx.RTryLock(ctx)
}
func (api *Api) RUnlock() {
	api.mx.RUnlock()
}

func (api *Api) MarshalResponce(ctx context.Context, compType compress.CompressionType, resp *Responce,
) (outCompType compress.CompressionType, bodyResponce []byte) {
	var err *mft.Error
	body := resp.Marchal()
	bodyResponce = body
	if compType != compress.NoCompression {
		outCompType, bodyResponce, err = api.Compressor.Compress(ctx, true, compType, body, nil)
	}
	if err != nil {
		panic(mft.GenerateErrorE(20400040, err, compType))
	}

	return outCompType, bodyResponce
}

func (api *Api) Do(ctx context.Context, compType compress.CompressionType, bodyRequest []byte,
) (outCompType compress.CompressionType, bodyResponce []byte) {
	var resp Responce
	outCompType = compress.NoCompression
	bodyIn := bodyRequest
	if compType != compress.NoCompression {
		_, bodyIn, resp.CommandResponce.Error = api.Compressor.Restore(ctx, compType, bodyRequest, nil)
		if resp.CommandResponce.Error != nil {
			resp.CommandResponce.Error =
				mft.GenerateErrorE(20400059, resp.CommandResponce.Error, compType)
			return api.MarshalResponce(ctx, outCompType, &resp)
		}
	}

	var req Request
	resp.CommandResponce.Error = req.Unmarchal(bodyIn)
	if resp.CommandResponce.Error != nil {
		resp.CommandResponce.Error =
			mft.GenerateErrorE(20400060, resp.CommandResponce.Error, compType)
		return api.MarshalResponce(ctx, outCompType, &resp)
	}

	if !api.RLock(ctx) {
		resp.CommandResponce.Error = mft.GenerateError(20400050)
		return api.MarshalResponce(ctx, outCompType, &resp)
	}
	defer api.RUnlock()

	if len(api.AuthenticationChecker) > 0 {
		at, ok := api.AuthenticationChecker[req.AuthType]
		if !ok {
			resp.CommandResponce.Error = mft.GenerateError(20400054, req.AuthType)
			return api.MarshalResponce(ctx, outCompType, &resp)
		}

		ok, userName, err := at.Check(ctx, req.UserNameRequest, req.SecretInfo)
		if err != nil {
			resp.CommandResponce.Error = mft.GenerateErrorE(20400055, err, req.UserNameRequest)
			return api.MarshalResponce(ctx, outCompType, &resp)
		}
		if !ok {
			resp.CommandResponce.Error = mft.GenerateError(20400056, req.UserNameRequest)
			return api.MarshalResponce(ctx, outCompType, &resp)
		}

		if req.CommandRequest.User == "" {
			req.CommandRequest.User = userName
		} else if req.CommandRequest.User == userName {
			// DO Nothing
		} else {
			if api.PermissionChecker != nil {
				allow, err := api.PermissionChecker.CheckPermission(ctx,
					authorization.UserName(userName),
					authorization.ObjectTypeAuthorization,
					authorization.ImpersonateAction,
					req.CommandRequest.User)
				if err != nil {
					resp.CommandResponce.Error = mft.GenerateErrorE(20400057, err, userName, req.CommandRequest.User)
					return api.MarshalResponce(ctx, outCompType, &resp)
				}
				if !allow {
					resp.CommandResponce.Error = mft.GenerateError(20400058, userName, req.CommandRequest.User)
					return api.MarshalResponce(ctx, outCompType, &resp)
				}
			}
		}
	}

	if api.ApisMap == nil {
		resp.CommandResponce.Error = mft.GenerateError(20400053)
		return api.MarshalResponce(ctx, outCompType, &resp)
	}

	mAct, ok := api.ApisMap[req.CommandRequest.ObjectType]
	if !ok || mAct == nil {
		resp.CommandResponce.Error = mft.GenerateError(20400051, req.CommandRequest.ObjectType)
		return api.MarshalResponce(ctx, outCompType, &resp)
	}

	a, ok := mAct[req.CommandRequest.Action]
	if !ok || a == nil {
		resp.CommandResponce.Error = mft.GenerateError(20400052, req.CommandRequest.Action)
		return api.MarshalResponce(ctx, outCompType, &resp)
	}

	rsp := a.DoRequest(ctx, &req.CommandRequest)
	resp.CommandResponce = *rsp

	if api.GetCompression != nil {
		outCompType, resp.PreferCompress = api.GetCompression(req.PreferCompress)
	}

	return api.MarshalResponce(ctx, outCompType, &resp)
}

type Request struct {
	UserNameRequest string                     `json:"user_name,omitempty"`
	SecretInfo      json.RawMessage            `json:"secret_info,omitempty"`
	AuthType        string                     `json:"auth_type,omitempty"`
	CommandRequest  ajt.CommandRequest         `json:"request,omitempty"`
	PreferCompress  []compress.CompressionType `json:"prefer_compress,omitempty"`
}
type Responce struct {
	CommandResponce ajt.CommandResponce        `json:"responce,omitempty"`
	PreferCompress  []compress.CompressionType `json:"prefer_compress,omitempty"`
}

func (r *Request) Marchal() json.RawMessage {
	body, er0 := json.Marshal(r)
	if er0 != nil {
		panic(mft.GenerateErrorE(20400000, er0))
	}
	return body
}
func (r *Request) Unmarchal(body json.RawMessage) *mft.Error {
	er0 := json.Unmarshal(body, r)
	if er0 != nil {
		return mft.GenerateErrorE(20400010, er0)
	}
	return nil
}
func (r *Responce) Marchal() json.RawMessage {
	body, er0 := json.Marshal(r)
	if er0 != nil {
		panic(mft.GenerateErrorE(20400020, er0))
	}
	return body
}
func (r *Responce) Unmarchal(body json.RawMessage) *mft.Error {
	er0 := json.Unmarshal(body, r)
	if er0 != nil {
		return mft.GenerateErrorE(20400030, er0)
	}
	return nil
}
