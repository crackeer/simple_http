package simple_http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const (
	contentTypeJSON = "application/json"
	contentTypeForm = "application/x-www-form-urlencoded"
)

var (
	apiMap    map[string]*ServiceAPI
	apiLocker *sync.RWMutex
	logger    *logrus.Logger
)

func init() {
	apiMap = map[string]*ServiceAPI{}
	apiLocker = &sync.RWMutex{}
}

// SetLogger
//
//	@param l
func SetLogger(l *logrus.Logger) {
	apiLocker.Lock()
	defer apiLocker.Unlock()
	logger = l
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

	var (
		newInput  map[string]interface{} = input
		newHeader map[string]string      = header

		response *resty.Response
		err      error
	)

	request := client.R()
	if logger != nil {
		request.EnableTrace()
		defer func() {
			trace := request.TraceInfo()
			entry := logger.WithFields(map[string]interface{}{
				"host":         api.Host,
				"path":         api.Path,
				"method":       api.Method,
				"content_type": api.ContentType,
				"header":       newHeader,
				"input":        newInput,
				"remote_addr":  trace.RemoteAddr.String(),
				"cost":         trace.TotalTime / time.Duration(time.Millisecond),
			})
			if err != nil {
				entry.Error(err.Error())
			} else {
				entry.WithField("response", string(response.Body()))
				entry.Info("success")
			}
		}()
	}

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
	response, err = request.Execute(api.Method, api.Path)
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
