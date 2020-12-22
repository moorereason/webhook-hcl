package config

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var (
	funcMap map[string]function.Function
)

func init() {
	funcMap = map[string]function.Function{
		"base64decode": base64decodeFunc(),
		"concat":       concatFunc(),
		"contains":     containsFunc(),
		"debug":        debugFunc(),
		"duration":     durationFunc(),
		"format":       stdlib.FormatFunc,
		"match":        matchFunc(),
		"sha1":         sha1Func(),
		"sha256":       sha256Func(),
		"since":        sinceFunc(),
	}
}

type Context struct {
	EvalContext *hcl.EvalContext

	Payload map[string]string
	Headers map[string]string
	Params  map[string]string
}

func NewContext() *Context {
	c := &Context{
		EvalContext: &hcl.EvalContext{
			Variables: map[string]cty.Value{},
			Functions: funcMap,
		},
	}

	c.EvalContext.Functions["header"] = c.HeaderFunc()
	c.EvalContext.Functions["payload"] = c.PayloadFunc()
	c.EvalContext.Functions["param"] = c.ParamFunc()

	return c
}

func (c *Context) PayloadFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "key",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			k := args[0].AsString()
			if v, ok := c.Payload[k]; ok {
				return cty.StringVal(v), nil
			}
			return cty.StringVal(""), nil
		},
	})
}

func (c *Context) HeaderFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "key",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			k := args[0].AsString()
			if v, ok := c.Headers[k]; ok {
				return cty.StringVal(v), nil
			}
			return cty.StringVal(""), nil
		},
	})
}

func (c *Context) ParamFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "key",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			k := args[0].AsString()
			if v, ok := c.Params[k]; ok {
				return cty.StringVal(v), nil
			}
			return cty.StringVal(""), nil
		},
	})
}

func sha1Func() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "data",
				Type: cty.String,
			},
			{
				Name: "secret",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			data := args[0].AsString()
			secret := args[1].AsString()

			mac := hmac.New(sha1.New, []byte(secret))
			_, err := mac.Write([]byte(data))
			if err != nil {
				return cty.StringVal(""), err
			}

			expectedMAC := hex.EncodeToString(mac.Sum(nil))

			// if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			//      return expectedMAC, &SignatureError{signature}
			// }
			return cty.StringVal(expectedMAC), err
		},
	})
}

func sha256Func() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "data",
				Type: cty.String,
			},
			{
				Name: "secret",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			data := args[0].AsString()
			secret := args[1].AsString()

			mac := hmac.New(sha256.New, []byte(secret))
			_, err := mac.Write([]byte(data))
			if err != nil {
				return cty.StringVal(""), err
			}

			expectedMAC := hex.EncodeToString(mac.Sum(nil))

			// if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			//      return expectedMAC, &SignatureError{signature}
			// }
			return cty.StringVal(expectedMAC), err
		},
	})
}

func matchFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "pattern",
				Type: cty.String,
			},
			{
				Name: "string",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			pattern := args[0].AsString()
			s := args[1].AsString()

			re, err := regexp.Compile(pattern)
			if err != nil {
				return cty.BoolVal(false), function.NewArgErrorf(0, "error parsing regexp pattern: %s", err)
			}

			return cty.BoolVal(re.MatchString(s)), nil
		},
	})
}

func concatFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s1",
				Type: cty.String,
			},
			{
				Name: "s2",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			s1 := args[0].AsString()
			s2 := args[1].AsString()

			return cty.StringVal(s1 + s2), nil
		},
	})
}

func containsFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s",
				Type: cty.String,
			},
			{
				Name: "substr",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			s := args[0].AsString()
			substr := args[1].AsString()

			return cty.BoolVal(strings.Contains(s, substr)), nil
		},
	})
}

func durationFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "duration",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			s := args[0].AsString()
			d, err := time.ParseDuration(s)
			return cty.NumberIntVal(int64(d)), err
		},
	})
}

func sinceFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "timestamp",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			var t time.Time
			var err error

			s := args[0].AsString()
			if s != "" {
				t, err = time.Parse(time.RFC1123, s)
				if err != nil {
					return cty.NumberIntVal(0), err
				}
			}

			return cty.NumberIntVal(int64(time.Since(t))), err
		},
	})
}

func base64decodeFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(stdlib.Bytes),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			s := args[0].AsString()

			data, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return stdlib.BytesVal([]byte{}), err
			}

			return stdlib.BytesVal(data), err
		},
	})
}

func debugFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "v",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			v := args[0].AsString()

			log.Print(v)

			return cty.BoolVal(true), nil
		},
	})
}