package config

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/apparentlymart/go-textseg/textseg"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type Context struct {
	EvalContext *hcl.EvalContext

	Payload map[string]interface{}
	Headers map[string]string
	Params  map[string]string

	Debug bool
}

func NewContext() *Context {
	c := &Context{
		EvalContext: &hcl.EvalContext{
			Variables: map[string]cty.Value{},
		},
	}

	c.EvalContext.Functions = map[string]function.Function{
		"join": stdlib.JoinFunc,

		"header":  c.HeaderFunc(),
		"payload": c.PayloadFunc(),
		"url":     c.ParamFunc(),
		// "request":  c.RequestFunc(),

		"all":          c.allFunc(),
		"and":          c.andFunc(),
		"any":          c.anyFunc(),
		"base64decode": c.base64decodeFunc(),
		"base64encode": c.base64encodeFunc(),
		"cidr":         c.cidrFunc(),
		"concat":       c.concatFunc(),
		"contains":     c.containsFunc(),
		"debug":        c.debugFunc(),
		"duration":     c.durationFunc(),
		"eq":           c.eqFunc(),
		"find":         c.findFunc(),
		"float":        c.floatFunc(),
		"format":       c.formatFunc(),
		"ge":           c.geFunc(),
		"getenv":       c.getenvFunc(),
		"gt":           c.gtFunc(),
		"le":           c.leFunc(),
		"len":          c.lenFunc(),
		"lower":        c.lowerFunc(),
		"lt":           c.ltFunc(),
		"match":        c.matchFunc(),
		"ne":           c.neFunc(),
		"not":          c.notFunc(),
		"or":           c.orFunc(),
		"readfile":     c.readfileFunc(),
		"sha1":         c.sha1Func(),
		"sha256":       c.sha256Func(),
		"sha512":       c.sha512Func(),
		"since":        c.sinceFunc(),
		"upper":        c.upperFunc(),
	}

	return c
}

func (c *Context) debugf(format string, v ...interface{}) {
	if c.Debug {
		log.Printf("DEBUG: "+format, v...)
	}
}

func (c *Context) PayloadFunc() function.Function {
	// TODO: rewrite to use stdlib.HasIndex() or something
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "key",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			k := args[0].AsString()
			kk := strings.ToLower(k)
			if v, ok := c.Payload[kk]; ok {
				s := fmt.Sprintf("%v", v)
				c.debugf("payload(%q) => [%T] %q", k, v, s)
				return cty.StringVal(s), nil
			}
			// TODO: should we return an error here?
			return cty.StringVal(""), nil // fmt.Errorf("failed to find payload value: %s", k)
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
			kk := strings.ToLower(k)
			if v, ok := c.Headers[kk]; ok {
				s := fmt.Sprintf("%v", v)
				c.debugf("header(%q) => %q", k, s)
				return cty.StringVal(v), nil
			}
			c.debugf("header(%q) => %q", k, "")
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
			kk := strings.ToLower(k)
			if v, ok := c.Params[kk]; ok {
				s := fmt.Sprintf("%v", v)
				c.debugf("param(%q) => %q", k, s)
				return cty.StringVal(v), nil
			}
			c.debugf("param(%q) => %q", k, "")
			return cty.StringVal(""), nil
		},
	})
}

func (c *Context) sha1Func() function.Function {
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

			c.debugf("sha1(%q, %q) => %q", data, secret, expectedMAC)
			return cty.StringVal(expectedMAC), err
		},
	})
}

func (c *Context) sha256Func() function.Function {
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

			c.debugf("sha256(%q, %q) => %q", data, secret, expectedMAC)
			return cty.StringVal(expectedMAC), err
		},
	})
}

func (c *Context) sha512Func() function.Function {
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

			c.debugf("sha512(%q, %q) => %q", data, secret, expectedMAC)
			return cty.StringVal(expectedMAC), err
		},
	})
}

func (c *Context) matchFunc() function.Function {
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

			ret := re.MatchString(s)
			c.debugf("match(%q, %q) => %v", pattern, s, ret)
			return cty.BoolVal(ret), nil
		},
	})
}

func (c *Context) concatFunc() function.Function {
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

			ret := s1 + s2
			c.debugf("concat(%q, %q) => %q", s1, s2, ret)
			return cty.StringVal(ret), nil
		},
	})
}

func (c *Context) containsFunc() function.Function {
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
			ret := strings.Contains(s, substr)

			c.debugf("contains(%q, %q) => %v", s, substr, ret)
			return cty.BoolVal(ret), nil
		},
	})
}

func (c *Context) durationFunc() function.Function {
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
			if err != nil {
				return cty.NumberIntVal(0), err
			}

			c.debugf("duration(%q) => %d", s, d)
			return cty.NumberIntVal(int64(d)), nil
		},
	})
}

func (c *Context) sinceFunc() function.Function {
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

			result := int64(time.Since(t))
			c.debugf("since(%q) => %v\n", s, result)

			return cty.NumberIntVal(result), err
		},
	})
}

func (c *Context) allFunc() function.Function {
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
					c.debugf("all(...) => false")
					return cty.False, nil
				}
			}

			c.debugf("all(...) => true")
			return cty.True, nil
		},
	})
}

func (c *Context) anyFunc() function.Function {
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
					c.debugf("any(...) => true")
					return cty.True, nil
				}
			}

			c.debugf("any(...) => false")
			return cty.False, nil
		},
	})
}

// TODO: String vs Bytes
func (c *Context) base64decodeFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s",
				Type: cty.String,
			},
		},
		// Type: function.StaticReturnType(stdlib.Bytes),
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			s := args[0].AsString()

			data, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return stdlib.BytesVal([]byte{}), err
			}

			ret := string(data)
			c.debugf("base64decode(%q) => %q", s, ret)
			return cty.StringVal(ret), nil
			// return stdlib.BytesVal(data), err
		},
	})
}

func (c *Context) base64encodeFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			s := args[0].AsString()
			data := base64.StdEncoding.EncodeToString([]byte(s))
			c.debugf("base64encode(%q) => %q", s, string(data))
			return cty.StringVal(data), nil
		},
	})
}

func (c *Context) cidrFunc() function.Function {
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
			a := args[0].AsString()
			b := args[1].AsString()

			_, cidr, err := net.ParseCIDR(a)
			if err != nil {
				return cty.BoolVal(false), err
			}

			ip := net.ParseIP(args[1].AsString())
			result := cidr.Contains(ip)
			c.debugf("cidr(%q, %q) => %v\n", a, b, result)

			return cty.BoolVal(result), nil
		},
	})
}

func (c *Context) debugFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "v",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// TODO: support other types besides string
			c.debugf("debug(%q)\n", args[0].AsString())

			return cty.BoolVal(true), nil
		},
	})
}

func (c *Context) getenvFunc() function.Function {
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
			ret := os.Getenv(v)
			c.debugf("getenv(%q) => %q", v, ret)
			return cty.StringVal(ret), nil
		},
	})
}

func (c *Context) readfileFunc() function.Function {
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

			c.debugf("readfile(%q) => %q", path, string(b))
			return cty.StringVal(string(b)), nil
		},
	})
}

func (c *Context) eqFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.DynamicPseudoType,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        true,
			},
			{
				Name:             "b",
				Type:             cty.DynamicPseudoType,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			// security: use constant time compare on strings
			if (args[0].Type() == cty.String || args[0].Type() == stdlib.Bytes) &&
				(args[1].Type() == cty.String || args[1].Type() == stdlib.Bytes) {

				a := args[0].AsString()
				b := args[1].AsString()
				result := subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
				c.debugf("eq(%q, %q) => %v\n", a, b, result)
				return cty.BoolVal(result), nil
			}

			log.Printf("HERE  eq(%s, %v)", args[0].Type().FriendlyName(), args[1].Type().FriendlyName())
			// TODO: Need tests
			ret = args[0].Equals(args[1])
			if c.Debug {
				// TODO: I don't like this...
				if a, err := convert.Convert(args[0], cty.String); err == nil {
					if b, err := convert.Convert(args[1], cty.String); err == nil {
						c.debugf("eq(%q, %q) => %v\n", a.AsString(), b.AsString(), ret.True())
					}
				}
			}
			return ret, nil
		},
	})
}

func (c *Context) lenFunc() function.Function {
	// from stdlib.StrlenFunc
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "str",
				Type:             cty.String,
				AllowDynamicType: true,
			},
		},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			in := args[0].AsString()
			l := 0

			inB := []byte(in)
			for i := 0; i < len(in); {
				d, _, _ := textseg.ScanGraphemeClusters(inB[i:], true)
				l++
				i += d
			}

			c.debugf("len(%q) => %d", in, l)
			return cty.NumberIntVal(int64(l)), nil
		},
	})
}

func (c *Context) findFunc() function.Function {
	// from stdlib.RegexFunc
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
		Type: func(args []cty.Value) (cty.Type, error) {
			if !args[0].IsKnown() {
				// We can't predict our type without seeing our pattern
				return cty.DynamicPseudoType, nil
			}

			retTy, err := regexPatternResultType(args[0].AsString())
			if err != nil {
				err = function.NewArgError(0, err)
			}
			return retTy, err
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			if retType == cty.DynamicPseudoType {
				return cty.DynamicVal, nil
			}

			re, err := regexp.Compile(args[0].AsString())
			if err != nil {
				// Should never happen, since we checked this in the Type function above.
				return cty.NilVal, function.NewArgErrorf(0, "error parsing pattern: %s", err)
			}
			str := args[1].AsString()

			captureIdxs := re.FindStringSubmatchIndex(str)
			if captureIdxs == nil {
				return cty.NilVal, fmt.Errorf("pattern did not match any part of the given string")
			}

			// TODO: c.debugf
			return regexPatternResult(re, str, captureIdxs, retType), nil
		},
	})
}

func (c *Context) andFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.Bool,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
			{
				Name:             "b",
				Type:             cty.Bool,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ret := args[0].And(args[1])
			c.debugf("and(...) => %t", ret.True())
			return ret, nil
		},
	})
}

func (c *Context) orFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.Bool,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
			{
				Name:             "b",
				Type:             cty.Bool,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ret := args[0].Or(args[1])
			c.debugf("or(...) => %t", ret.True())
			return ret, nil
		},
	})
}

func (c *Context) notFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "val",
				Type:             cty.Bool,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ret := args[0].Not()
			c.debugf("not(...) => %t", ret.True())
			return ret, nil
		},
	})
}

func (c *Context) upperFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "str",
				Type:             cty.String,
				AllowDynamicType: true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			in := args[0].AsString()
			out := strings.ToUpper(in)
			c.debugf("upper(%q) => %s", in, out)
			return cty.StringVal(out), nil
		},
	})
}

func (c *Context) lowerFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "str",
				Type:             cty.String,
				AllowDynamicType: true,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			in := args[0].AsString()
			out := strings.ToLower(in)
			c.debugf("lower(%q) => %s", in, out)
			return cty.StringVal(out), nil
		},
	})
}

func (c *Context) formatFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "format",
				Type: cty.String,
			},
		},
		VarParam: &function.Parameter{
			Name:      "args",
			Type:      cty.DynamicPseudoType,
			AllowNull: true,
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ret, err := stdlib.FormatFunc.Call(args)
			if err != nil {
				return ret, err
			}

			// TODO: how to make this work with any data types?
			vargs := make([]string, len(args[1:]))
			for i, v := range args[1:] {
				vargs[i] = `"` + v.AsString() + `"`
			}
			c.debugf("format(%q, %s) => %q", args[0].AsString(), strings.Join(vargs, ", "), ret.AsString())
			return ret, err
		},
	})
}

func (c *Context) geFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
			{
				Name:             "b",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			ret = args[0].GreaterThanOrEqualTo(args[1])
			c.debugf("ge(%v, %v) => %t", args[0].AsBigFloat(), args[1].AsBigFloat(), ret.True())
			return ret, nil
		},
	})
}

func (c *Context) gtFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
			{
				Name:             "b",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			ret = args[0].GreaterThan(args[1])
			c.debugf("gt(%v, %v) => %t", args[0].AsBigFloat(), args[1].AsBigFloat(), ret.True())
			return ret, nil
		},
	})
}

func (c *Context) leFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
			{
				Name:             "b",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			ret = args[0].LessThanOrEqualTo(args[1])
			c.debugf("le(%v, %v) => %t", args[0].AsBigFloat(), args[1].AsBigFloat(), ret.True())
			return ret, nil
		},
	})
}

func (c *Context) ltFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
			{
				Name:             "b",
				Type:             cty.Number,
				AllowDynamicType: true,
				AllowMarked:      true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			ret = args[0].LessThan(args[1])
			c.debugf("lt(%v, %v) => %t", args[0].AsBigFloat(), args[1].AsBigFloat(), ret.True())
			return ret, nil
		},
	})
}

func (c *Context) neFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name:             "a",
				Type:             cty.DynamicPseudoType,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        true,
			},
			{
				Name:             "b",
				Type:             cty.DynamicPseudoType,
				AllowUnknown:     true,
				AllowDynamicType: true,
				AllowNull:        true,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
			ret = args[0].Equals(args[1]).Not()
			// TODO: print params?
			c.debugf("ne(...) => %t", ret.True())
			return ret, nil
		},
	})
}

// TODO: do we need this?
func (c *Context) floatFunc() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "s",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			ret, err := convert.Convert(args[0], cty.Number)
			if err != nil {
				return ret, err
			}

			// TODO: print params?
			c.debugf("float(...) => %d", ret.AsBigFloat())
			return ret, nil
		},
	})
}
