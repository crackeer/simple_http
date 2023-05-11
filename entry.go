package simple_http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

const (
	contentTypeJSON = "application/json"
	contentTypeForm = "application/x-www-form-urlencoded"
)

var (
	apiMap    map[string]*ServiceAPI
	apiLocker *sync.RWMutex
)

func init() {
	apiMap = map[string]*ServiceAPI{}
	apiLocker = &sync.RWMutex{}
}

// RegisterServiceAPI
//
//	@param name
//	@param api
func RegisterServiceAPI(name string, api *ServiceAPI) {
	apiLocker.Lock()
	defer apiLocker.Unlock()
	apiMap[name] = api
}

// GetServiceAPI
//
//	@param name
//	@return *ServiceAPI
func GetServiceAPI(name string) *ServiceAPI {
	if api, ok := apiMap[name]; ok {
		return api
	}
	return nil
}

// RequestServiceAPIByName
//
//	@param name
//	@param input
//	@param header
//	@return *APIResponse
//	@return error
func RequestServiceAPIByName(name string, input map[string]interface{}, header map[string]string) (*APIResponse, error) {
	api := GetServiceAPI(name)
	if api == nil {
		return nil, fmt.Errorf("api config `%s` not found", name)
	}
	return RequestServiceAPI(api, input, header)
}

// RequestServiceAPI
//
//	@param api
//	@param input
//	@param header
//	@return *APIResponse
//	@return error
func RequestServiceAPI(api *ServiceAPI, input map[string]interface{}, header map[string]string) (*APIResponse, error) {
	response, err := DoRequest(api, input, header)
	if err != nil {
		return nil, err
	}

	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("http error: %v", response.Status())
	}

	return extractResponse(response.Body(), api), nil
}

// DoRequest
//
//	@param api
//	@param input
//	@param header
//	@return *resty.Response
//	@return error
func DoRequest(api *ServiceAPI, input map[string]interface{}, header map[string]string) (*resty.Response, error) {
	client := resty.New()
	client.SetBaseURL(api.Host)
	if api.Timeout > 0 {
		client.SetTimeout(time.Duration(api.Timeout) * time.Millisecond)
	}

	request := client.R()

	var (
		newInput  map[string]interface{} = input
		newHeader map[string]string      = header
	)

	if len(api.SignName) > 0 {
		sign := GetSign(api.SignName)
		if sign == nil {
			return nil, fmt.Errorf("sign `%s` not supported", api.SignName)
		}
		tmpInput, tmpHeader, err := sign(input, header, api.SignConfig)
		if err != nil {
			return nil, fmt.Errorf("signature error:%s", err.Error())
		}
		newInput, newHeader = tmpInput, tmpHeader
	}
	request.SetHeaders(api.StaticHeader)
	request.SetHeaders(newHeader)
	for {
		if api.Method == http.MethodGet {
			request = request.SetQueryParams(Map2MapString(newInput))
			break
		}

		if api.Method == http.MethodPost {
			if api.Method == contentTypeJSON {
				request.SetBody(newInput)
				break
			}
			request.SetMultipartFormData(Map2MapString(newInput))
		}
	}

	return executeRequest(request, api.Method, api.Path)

}

func executeRequest(req *resty.Request, method string, fullURL string) (*resty.Response, error) {
	var response *resty.Response
	var err error
	switch method {
	case http.MethodGet:
		response, err = req.Get(fullURL)
	case http.MethodPost:
		response, err = req.Post(fullURL)
	}
	return response, err
}

func extractResponse(bytes []byte, api *ServiceAPI) *APIResponse {

	if api.DisableExtract {
		var data interface{}
		if err := json.Unmarshal(bytes, &data); err != nil {
			data = string(bytes)
		}
		return &APIResponse{
			Code:    "",
			Message: "",
			Data:    data,
			Error:   false,
		}
	}

	code := gjson.GetBytes(bytes, api.CodeKey).String()
	message := gjson.GetBytes(bytes, api.MessageKey).String()
	if code != api.SuccessCode {
		return &APIResponse{
			Code:    code,
			Message: message,
			Data:    nil,
			Error:   true,
		}
	}
	data := gjson.GetBytes(bytes, api.DataKey).Value()
	return &APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
		Error:   false,
	}
}
