package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type Config struct {
	Servers []Server `hcl:"server,block"`
}

type Server struct {
	IP          *string   `hcl:"ip"`
	Port        *int      `hcl:"port"`
	Secure      *bool     `hcl:"secure"`
	HTTPMethods *[]string `hcl:"http_methods"`
	RawHooks    hcl.Body  `hcl:",remain"` // See https://hcl.readthedocs.io/en/latest/go_decoding_gohcl.html#partial-decoding

	Hooks []Hook
}

type HooksConfig struct {
	Hooks []Hook `hcl:"hook,block"`
}

type Hook struct {
	ID            string   `hcl:"id,label"`
	Request       *Request `hcl:"request,block"`
	PreExecConfig hcl.Body `hcl:",remain"`

	// Request     *Request
	Constraints *[]bool
	Task        Task
	Response    *Response
}

type Request struct {
	IncomingPayloadContentType *string   `hcl:"content_type"`
	JSONStringParameters       *[]string `hcl:"json_parameters"`
}

type PreExecConfig struct {
	Constraints    *[]bool  `hcl:"constraints"`
	Task           Task     `hcl:"task,block"`
	PostExecConfig hcl.Body `hcl:",remain"`
}

type Task struct {
	ExecuteCommand           []string           `hcl:"cmd"`
	CommandWorkingDirectory  *string            `hcl:"workdir"`
	PassEnvironmentToCommand *map[string]string `hcl:"env_vars"`
	PassFile                 *PassFile          `hcl:"pass_file,block"`
	File                     *File              `hcl:"create_file,block"`
	// CaptureCommandOutput        *bool              `hcl:"capture_output"`
	// CaptureCommandOutputOnError *bool              `hcl:"capture_outout_on_error"`
}

type File struct {
	Content  []byte  `hcl:"content"`
	Filename string  `hcl:"filename"`
	Keep     *bool   `hcl:"keep"`
	EnvName  *string `hcl:"envname"`
}

type PassFile struct {
	Source       string  `hcl:"source"`
	Name         string  `hcl:"name"`
	Filename     string  `hcl:"filename"`
	Base64Decode *bool   `hcl:"base64decode"`
	Keep         *bool   `hcl:"keep"`
	EnvName      *string `hcl:"envname"`
}

type PostExecConfig struct {
	Response *Response `hcl:"response,block"`
}

type Response struct {
	ResponseSuccess     *ResponseSuccess     `hcl:"success,block"`
	ResponseError       *ResponseError       `hcl:"error,block"`
	ResponseUnsatisfied *ResponseUnsatisfied `hcl:"unsatisfied_constraints,block"`
}

type ResponseError struct {
	StatusCode  *int               `hcl:"status_code"`
	ContentType *string            `hcl:"content_type"`
	Body        *string            `hcl:"body"`
	Headers     *map[string]string `hcl:"headers"`
}

type ResponseSuccess struct {
	StatusCode  *int               `hcl:"status_code"`
	ContentType *string            `hcl:"content_type"`
	Body        *string            `hcl:"body"`
	Headers     *map[string]string `hcl:"headers"`
}

type ResponseUnsatisfied struct {
	StatusCode  *int               `hcl:"status_code"`
	ContentType *string            `hcl:"content_type"`
	Body        *string            `hcl:"body"`
	Headers     *map[string]string `hcl:"headers"`
}

func (s Server) Dump() {
	fmt.Println("Server:")
	if s.IP != nil {
		fmt.Println("  IP: ", *s.IP)
	}
	if s.Port != nil {
		fmt.Println("  Port: ", *s.Port)
	}
	if s.Secure != nil {
		fmt.Println("  Secure: ", *s.Secure)
	}

	for _, h := range s.Hooks {
		fmt.Println("  Hook:")
		fmt.Println("    ID: ", h.ID)
		if h.Request != nil {
			fmt.Println("    Request:")
			if h.Request.IncomingPayloadContentType != nil {
				fmt.Println("      IncomingPayloadContentType:", *h.Request.IncomingPayloadContentType)
			}
			if h.Request.JSONStringParameters != nil {
				fmt.Println("      JSONStringParameters:", *h.Request.JSONStringParameters)
			}
		}

		if h.Constraints != nil {
			fmt.Println("    Constraints:", *h.Constraints)
		}

		fmt.Println("    Task:")
		fmt.Println("      ExecuteCommand:", h.Task.ExecuteCommand)
		if h.Task.CommandWorkingDirectory != nil {
			fmt.Println("      CommandWorkingDirectory:", *h.Task.CommandWorkingDirectory)
		}
		if h.Task.PassEnvironmentToCommand != nil {
			fmt.Println("      PassEnvironmentToCommand:", *h.Task.PassEnvironmentToCommand)
		}
		if h.Task.PassFile != nil {
			fmt.Println("      PassFile:", *h.Task.PassFile)
		}
		if h.Task.File != nil {
			fmt.Println("      File:", *h.Task.File)
		}

		if h.Response != nil {
			fmt.Println("    Response:")
			if h.Response.ResponseSuccess != nil {
				fmt.Println("      Success:")
				if h.Response.ResponseSuccess.StatusCode != nil {
					fmt.Println("        StatusCode:", *h.Response.ResponseSuccess.StatusCode)
				}
				if h.Response.ResponseSuccess.ContentType != nil {
					fmt.Println("        ContentType:", *h.Response.ResponseSuccess.ContentType)
				}
				if h.Response.ResponseSuccess.Body != nil {
					fmt.Println("        Body:", *h.Response.ResponseSuccess.Body)
				}
				if h.Response.ResponseSuccess.Headers != nil {
					fmt.Println("        Headers:", *h.Response.ResponseSuccess.Headers)
				}
			}
			if h.Response.ResponseError != nil {
				fmt.Println("      Error:")
				if h.Response.ResponseError.StatusCode != nil {
					fmt.Println("        StatusCode:", *h.Response.ResponseError.StatusCode)
				}
				if h.Response.ResponseError.ContentType != nil {
					fmt.Println("        ContentType:", *h.Response.ResponseError.ContentType)
				}
				if h.Response.ResponseError.Body != nil {
					fmt.Println("        Body:", *h.Response.ResponseError.Body)
				}
				if h.Response.ResponseError.Headers != nil {
					fmt.Println("        Headers:", *h.Response.ResponseError.Headers)
				}
			}
			if h.Response.ResponseUnsatisfied != nil {
				fmt.Println("      Unsatisfied:")
				if h.Response.ResponseUnsatisfied.StatusCode != nil {
					fmt.Println("        StatusCode:", *h.Response.ResponseUnsatisfied.StatusCode)
				}
				if h.Response.ResponseUnsatisfied.ContentType != nil {
					fmt.Println("        ContentType:", *h.Response.ResponseUnsatisfied.ContentType)
				}
				if h.Response.ResponseUnsatisfied.Body != nil {
					fmt.Println("        Body:", *h.Response.ResponseUnsatisfied.Body)
				}
				if h.Response.ResponseUnsatisfied.Headers != nil {
					fmt.Println("        Headers:", *h.Response.ResponseUnsatisfied.Headers)
				}
			}
		}
		fmt.Println("")
	}
}
