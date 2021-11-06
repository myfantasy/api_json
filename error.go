package api_json

import "github.com/myfantasy/mft"

// Errors codes and description
var Errors map[int]string = map[int]string{
	20400000: "api_json.Request.Marchal: fail to marshal",
	20400010: "api_json.Request.Unmarchal: fail to unmarshal",
	20400020: "api_json.Responce.Marchal: fail to marshal",
	20400030: "api_json.Responce.Unmarchal: fail to unmarshal",
	20400040: "api_json.Api.MarshalResponce: fail to compress compress type `%v`",
	20400050: "api_json.Api.Do: R lock fail",
	20400051: "api_json.Api.Do: Unknown object type `%v`",
	20400052: "api_json.Api.Do: Unknown action `%v`",
	20400053: "api_json.Api.Do: api not init",
	20400054: "api_json.Api.Do: authentication type not allowed `%v`",
	20400055: "api_json.Api.Do: authentication internal error for user `%v`",
	20400056: "api_json.Api.Do: authentication fail for user `%v`",
	20400057: "api_json.Api.Do: impersonate for user `%v` as `%v` internal error",
	20400058: "api_json.Api.Do: impersonate for user `%v` as `%v` fail",
	20400059: "api_json.Api.Do: fail to restore compress type `%v`",
	20400060: "api_json.Api.Do: fail to unmarshal request (compress type `%v`)",

	20400200: "api_json.ApiProvider.DoRequest: R lock fail",
	20400201: "api_json.ApiProvider.DoRequest: fail to compress; compress type `%v`",
	20400202: "api_json.ApiProvider.DoRequest: fail to restore; compress type `%v`",
	20400203: "api_json.ApiProvider.DoRequest: unmarshal responce fail",
	20400204: "api_json.ApiProvider.DoRequest: fail to call CallFunc; object type `%v` action `%v`",

	20400400: "api_json.ServiceApi.DoRequest: Unknown object type `%v` action `%v`",
}

func init() {
	mft.AddErrorsCodes(Errors)
}
