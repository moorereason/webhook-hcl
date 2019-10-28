package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type C struct {
	Servers []Server `hcl:"server,block"`
}

type Server struct {
	IP       *string  `hcl:"ip"`
	Port     *int     `hcl:"port"`
	Secure   *bool    `hcl:"secure"`
	RawHooks hcl.Body `hcl:",remain"` // See https://hcl.readthedocs.io/en/latest/go_decoding_gohcl.html#partial-decoding

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
	Constraints *[]string
	Task        Task
	Response    *Response
}

type Request struct {
	IncomingPayloadContentType *string   `hcl:"content_type"`
	JSONStringParameters       *[]string `hcl:"json_parameters"`
}

type PreExecConfig struct {
	Constraints    *[]string `hcl:"constraints"`
	Task           Task      `hcl:"task,block"`
	PostExecConfig hcl.Body  `hcl:",remain"`
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
	SuccessHttpResponseCode             *int               `hcl:"success_response_code"`
	TriggerRuleMismatchHttpResponseCode *int               `hcl:"failed_constraints_response_code"`
	ContentType                         *string            `hcl:"content_type"`
	Body                                *string            `hcl:"body"`
	Headers                             *map[string]string `hcl:"headers"`
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
			if h.Response.SuccessHttpResponseCode != nil {
				fmt.Println("      SuccessHttpResponseCode:", *h.Response.SuccessHttpResponseCode)
			}
			if h.Response.TriggerRuleMismatchHttpResponseCode != nil {
				fmt.Println("      TriggerRuleMismatchHttpResponseCode:", *h.Response.TriggerRuleMismatchHttpResponseCode)
			}
			if h.Response.ContentType != nil {
				fmt.Println("      ContentType:", *h.Response.ContentType)
			}
			if h.Response.Body != nil {
				fmt.Println("      Body:", *h.Response.Body)
			}
			if h.Response.Headers != nil {
				fmt.Println("      Headers:", *h.Response.Headers)
			}
		}
		fmt.Println("")
	}
}
