package api_json

import (
	"context"
	"encoding/json"

	ajt "github.com/myfantasy/api_json_types"
	"github.com/myfantasy/compress"
	"github.com/myfantasy/mfs"
	"github.com/myfantasy/mft"
)

type CallFunction func(ctx context.Context, compType compress.CompressionType, bodyRequest []byte,
) (outCompType compress.CompressionType, bodyResponce []byte)

type ApiProvider struct {
	Compressor *compress.Generator

	CallFunc CallFunction

	GetCompression GetCompressionFunc

	PreferCompressLastResponce []compress.CompressionType

	AuthType        string
	UserNameRequest string
	SecretInfo      json.RawMessage

	mx mfs.RWTMutex
}

func (api *ApiProvider) RLock(ctx context.Context) bool {
	return api.mx.RTryLock(ctx)
}
func (api *ApiProvider) RUnlock() {
	api.mx.RUnlock()
}

func (api *ApiProvider) UpdatePreferCompressLast(preferCompressLast []compress.CompressionType) {
	// is rlocks there
	if len(preferCompressLast) == 0 {
		return
	}
	if len(preferCompressLast) == len(api.PreferCompressLastResponce) {
		ok := true
		for i := 0; i < len(preferCompressLast); i++ {
			ok = preferCompressLast[i] == api.PreferCompressLastResponce[i]
			if !ok {
				break
			}
		}
		if ok {
			return
		}
	}
	go func() {
		api.mx.Lock()
		defer api.mx.Unlock()
		api.PreferCompressLastResponce = preferCompressLast
	}()
}

func (sa *ApiProvider) AllowedCommands() []ajt.CommandDescription {
	return nil
}
func (ap *ApiProvider) DoRequest(ctx context.Context, reqCmd *ajt.CommandRequest) *ajt.CommandResponce {
	var req Request
	var respCmd ajt.CommandResponce
	if !ap.RLock(ctx) {
		respCmd.Error = mft.GenerateError(20400200)
		return &respCmd
	}
	defer ap.RUnlock()
	req.AuthType = ap.AuthType
	req.UserNameRequest = ap.UserNameRequest
	req.SecretInfo = ap.SecretInfo

	req.CommandRequest = *reqCmd

	callCompType := compress.NoCompression
	var callPreferCompress []compress.CompressionType
	if ap.GetCompression != nil {
		callCompType, callPreferCompress = ap.GetCompression(ap.PreferCompressLastResponce)
	}
	req.PreferCompress = callPreferCompress

	body := req.Marchal()
	bodyComp := body
	if callCompType != compress.NoCompression {
		callCompType, bodyComp, respCmd.Error = ap.Compressor.Compress(ctx, true, callCompType, bodyComp, nil)
		if respCmd.Error != nil {
			respCmd.Error = mft.GenerateErrorE(20400201, respCmd.Error, callCompType)
			return &respCmd
		}
	}

	outCompType, bodyResponce := ap.CallFunc(ctx, callCompType, bodyComp)

	uncBodyResponce := bodyResponce
	if outCompType != compress.NoCompression {
		_, uncBodyResponce, respCmd.Error = ap.Compressor.Restore(ctx, outCompType, uncBodyResponce, nil)
		if respCmd.Error != nil {
			respCmd.Error = mft.GenerateErrorE(20400202, respCmd.Error, outCompType)
			return &respCmd
		}
	}

	var resp Responce
	respCmd.Error = resp.Unmarchal(uncBodyResponce)
	if respCmd.Error != nil {
		respCmd.Error = mft.GenerateErrorE(20400203, respCmd.Error)
		return &respCmd
	}

	ap.UpdatePreferCompressLast(resp.PreferCompress)

	return &resp.CommandResponce
}
