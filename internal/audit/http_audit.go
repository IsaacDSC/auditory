package audit

type HttpAudit struct {
	Request  RequestAudit  `json:"request"`
	Response ResponseAudit `json:"response"`
}

type RequestAudit struct {
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Query   string              `json:"query"`
}

type ResponseAudit struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}
