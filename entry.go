package simple_http

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

func RegisterServiceAPI(name string, api *ServiceAPI) {
}

func GetServiceAPI(name string) *ServiceAPI {
	return nil
}

func RequestServiceAPIByName(name string, input, header map[string]interface{}) (*APIResponse, error) {
	return nil, nil
}

func RequestServiceAPI(api *ServiceAPI, input map[string]interface{}, header map[string]string) (*APIResponse, error) {
	client := resty.New()
	client.SetTimeout(time.Duration(api.Timeout) * time.Millisecond)

	request := client.R()

	var (
		newInput  map[string]interface{} = input
		newHeader map[string]string      = header
	)

	if len(api.SignType) > 0 {
		sign := GetSignature(api.SignType)
		tmpInput, tmpHeader, err := sign(input, header, api.SignConfig)
		if err != nil {
			return nil, fmt.Errorf("signature error:%s", err.Error())
		}
		newInput, newHeader = tmpInput, tmpHeader
	}
	request.SetHeaders(api.StaticHeader)
	request.SetHeaders(newHeader)

	return nil, nil
}
