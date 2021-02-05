package config

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var funcMap map[string]function.Function

func init() {
	funcMap = map[string]function.Function{
		// stdlib functions
		"and":    stdlib.AndFunc,
		"eq":     stdlib.EqualFunc,
		"format": stdlib.FormatFunc,
		"ge":     stdlib.GreaterThanOrEqualToFunc,
		"gt":     stdlib.GreaterThanFunc,
		"join":   stdlib.JoinFunc,
		"le":     stdlib.LessThanOrEqualToFunc,
		"lower":  stdlib.LowerFunc,
		"lt":     stdlib.LessThanFunc,
		"ne":     stdlib.NotEqualFunc,
		"not":    stdlib.NotFunc,
		"or":     stdlib.OrFunc,
		"upper":  stdlib.UpperFunc,
		"find":   stdlib.RegexFunc,
		"len":    stdlib.LengthFunc,

		"all":          allFunc(),
		"any":          anyFunc(),
		"base64decode": base64decodeFunc(),
		"base64encode": base64encodeFunc(),
		"cidr":         cidrFunc(),
		"concat":       concatFunc(),
		"contains":     containsFunc(),
		"debug":        debugFunc(),
		"duration":     durationFunc(),
		"getenv":       getenvFunc(),
		"match":        matchFunc(),
		"readfile":     readfileFunc(),
		"sha1":         sha1Func(),
		"sha256":       sha256Func(),
		"sha512":       sha512Func(),
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
	// c.EvalContext.Functions["request"] = c.RequestFunc()

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
			return cty.StringVal(""), fmt.Errorf("failed to find payload value: %s", k)
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
			k := strings.ToLower(args[0].AsString())
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

			// TODO: disabled for testing
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

			// TODO: disabled for testing
			// if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			//      return expectedMAC, &SignatureError{signature}
			// }
			return cty.StringVal(expectedMAC), err
		},
	})
}

func sha512Func() function.Function {
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

			mac := hmac.New(sha512.New, []byte(secret))
			_, err := mac.Write([]byte(data))
			if err != nil {
				return cty.StringVal(""), err
			}

			expectedMAC := hex.EncodeToString(mac.Sum(nil))

			// TODO: disabled for testing
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

func allFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		VarParam: &function.Parameter{
			Name:             "conditions",
			Type:             cty.Bool,
			AllowDynamicType: true,
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if len(args) == 0 {
				return cty.NilVal, fmt.Errorf("must pass at least one condition")
			}

			for _, v := range args {
				if v.False() {
					return cty.False, nil
				}
			}

			return cty.True, nil
		},
	})
}

func anyFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{},
		VarParam: &function.Parameter{
			Name:             "conditions",
			Type:             cty.Bool,
			AllowDynamicType: true,
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if len(args) == 0 {
				return cty.NilVal, fmt.Errorf("must pass at least one condition")
			}

			for _, v := range args {
				if v.True() {
					return cty.True, nil
				}
			}

			return cty.False, nil
		},
	})
}

// TODO: String vs Bytes
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

func base64encodeFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			b := args[0].AsString()
			data := base64.StdEncoding.EncodeToString([]byte(b))
			return cty.StringVal(data), nil
		},
	})
}

func cidrFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "cidr",
				Type: cty.String,
			},
			{
				Name: "ip",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			_, cidr, err := net.ParseCIDR(args[0].AsString())
			if err != nil {
				return cty.BoolVal(false), err
			}

			ip := net.ParseIP(args[1].AsString())

			return cty.BoolVal(cidr.Contains(ip)), nil
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

func getenvFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "var",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			v := args[0].AsString()
			return cty.StringVal(os.Getenv(v)), nil
		},
	})
}

func readfileFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()

			b, err := os.ReadFile(path)
			if err != nil {
				return cty.BoolVal(false), err
			}
			return cty.StringVal(string(b)), nil
		},
	})
}
