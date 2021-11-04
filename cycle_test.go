package api_json

import (
	"context"
	"testing"
	"time"

	"github.com/myfantasy/authentication"
	"github.com/myfantasy/authentication/sat"
	"github.com/myfantasy/authorization/saz"
	"github.com/myfantasy/compress"
	"github.com/myfantasy/mft"
)

func TestCycle(t *testing.T) {
	var gcf GetCompressionFunc
	gcf = func(preferCompress []compress.CompressionType,
	) (outCompType compress.CompressionType, outPreferCompress []compress.CompressionType) {
		outCompType = compress.Zip
		outPreferCompress = append(outPreferCompress, compress.Zip, compress.NoCompression)
		return outCompType, outPreferCompress
	}
	api := &Api{
		Compressor: compress.GeneratorCreate(7),

		GetCompression: gcf,

		AuthenticationChecker: map[string]authentication.AuthenticationChecker{
			"simple": &sat.SimpleAuthenticationChecker{
				Users: map[string]sat.User{
					"admin": {
						Name:       "admin",
						Pwd:        "123",
						PwdIsEnc:   false,
						IsDisabled: false,
					},
					"test1": {
						Name:       "test1",
						Pwd:        "",
						PwdIsEnc:   false,
						IsDisabled: false,
					},
					"test2": {
						Name:       "test2",
						Pwd:        "",
						PwdIsEnc:   false,
						IsDisabled: true,
					},
				},
			},
		},
		PermissionChecker: &saz.SimplePermissionChecker{
			Users: map[string]saz.User{
				"admin": {
					Name:    "admin",
					IsAdmin: true,
				},
			},
		},
	}

	api.AddApi(&ServiceApi{})

	ap := &ApiProvider{
		CallFunc: func(ctx context.Context, compType compress.CompressionType,
			bodyRequest []byte, waitDuration time.Duration,
		) (outCompType compress.CompressionType, bodyResponce []byte, err *mft.Error) {
			outCompType, bodyResponce = api.Do(ctx, compType, bodyRequest)
			return outCompType, bodyResponce, err
		},
		Compressor: compress.GeneratorCreate(5),

		GetCompression: gcf,

		//PreferCompressLastResponce []compress.CompressionType

		AuthType:        "simple",
		UserNameRequest: "admin",
		SecretInfo:      (&sat.Request{Pwd: "123"}).ToSecretInfo(),
	}

	err := PingAsUser(context.Background(), ap, "abra")

	if err != nil {
		t.Fatal(err)
	}

	ap.UserNameRequest = "test1"
	err = PingAsUser(context.Background(), ap, "abra")

	if err == nil {
		t.Fatal("should be: 'api_json.Api.Do: impersonate for user `test1` as `abra` fail'")
	}
	if err.Code != 20400058 {
		t.Fatal(err)
	}
}
