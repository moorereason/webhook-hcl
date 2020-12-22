// +build ignore

package config

// Config is the configuration for a given instance of webhook.
type Config struct {
	IP          string
	Port        int
	Secure      bool
	HTTPMethods []string
	Hooks       []*Hook
}

// Hook is the configuration for a given webhook.
type Hook struct {
	ID          string
	Request     *Request
	Constraints *[]Constraint
	Task        *Task
	Response    *Response
}

// Request contains the request configuration.
type Request struct {
	IncomingPayloadContentType string
	JSONStringParameters       []string
}

type Constraint struct {
	Raw string
	Ok  bool
}
