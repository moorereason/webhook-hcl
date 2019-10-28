package config

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var (
	funcMap map[string]function.Function
	// varMap map[string]cty.Value

	Payload map[string]string
	Headers map[string]string
	Params  map[string]string
)

func init() {
	funcMap = map[string]function.Function{
		"base64decode": base64decodeFunc(),
		"duration":     durationFunc(),
		"format":       stdlib.FormatFunc,
		"header":       headerFunc(),
		"match":        matchFunc(),
		"param":        paramFunc(),
		"payload":      payloadFunc(),
		"sha1":         sha1Func(),
		"sha256":       sha256Func(),
		"since":        sinceFunc(),
	}
}

func NewContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: funcMap,
	}
}

func payloadFunc() function.Function {
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
			if v, ok := Payload[k]; ok {
				return cty.StringVal(v), nil
			}
			return cty.StringVal(""), nil
		},
	})
}

func headerFunc() function.Function {
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
			if v, ok := Headers[k]; ok {
				return cty.StringVal(v), nil
			}
			return cty.StringVal(""), nil
		},
	})
}

func paramFunc() function.Function {
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
			if v, ok := Params[k]; ok {
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

			return cty.NumberIntVal(int64(time.Now().Sub(t))), err
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
