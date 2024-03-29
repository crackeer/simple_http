package simple_http

// ServiceAPI
type ServiceAPI struct {
	Host           string                 `json:"host"`
	Sign           string                 `json:"sign"`
	SignConfig     map[string]interface{} `json:"sign_config"`
	Path           string                 `json:"path"`
	ContentType    string                 `json:"content_type"`
	Method         string                 `json:"method"`
	Timeout        int64                  `json:"timeout"`
	SuccessCode    string                 `json:"success_code"`
	MessageKey     string                 `json:"message_key"`
	CodeKey        string                 `json:"code_key"`
	DataKey        string                 `json:"data_key"`
	DisableExtract bool                   `json:"disable_extract"`
	StaticHeader   map[string]string      `json:"header"`
}

// APIResponse
type APIResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   bool        `json:"error"`
}
