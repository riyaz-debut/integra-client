package utils

type Response struct {
	Status     int         `json:"status,omitempty"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Role	   string	   `json:"role,omitempty"`
	OrgData	   interface{} `json:"org_data,omitempty"`
	Err        error       `json:"err,omitempty"`
	HttpStatus interface{} `json:"http_status,omitempty"`
}
