package simple_http

type ServiceAPI struct {
	Host         string                 `json:"host"`
	SignType     string                 `json:"sign_type"`
	SignConfig   map[string]interface{} `json:"sign_config"`
	Path         string                 `json:"path"`
	ContentType  string                 `json:"content_type"`
	Method       string                 `json:"method"`
	Timeout      int64                  `json:"timeout"`
	SuccessCode  string                 `json:"success_code"`
	MessageKey   string                 `json:"message_key"`
	CodeKey      string                 `json:"code_key"`
	DataKey      string                 `json:"data_key"`
	StaticHeader map[string]string      `json:"header"`
}

// APIResponse
type APIResponse struct {
	Code           string
	Message        string
	Data           string
	HttpStatusCode int64
	OriginBody     []byte
}

// ServiceAPIFactory
type ServiceAPIFactory interface {
	GetAPI() *ServiceAPI
}
