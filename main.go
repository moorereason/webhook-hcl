package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/moorereason/webhook-hcl/internal/config"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf("Usage: %s\n FILE", os.Args[0])
		os.Exit(1)
	}

	t0 := time.Now()

	ct := time.Now()
	/////
	// Initialize Service Config
	/////

	conf, err := loadConfigFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("%% Initialize Service Config\n%% TIME", time.Since(ct))

	// fmt.Printf("1 config: %#v\n", conf)
	// conf[0].Dump()

	/////
	// Initialize hooks
	/////
	t1 := time.Now()

	ctx := config.NewContext()
	ctx.Debug = true

	ct = time.Now()
	var hb config.HooksConfig
	diags := gohcl.DecodeBody(conf.RawHooks, ctx.EvalContext, &hb)
	if diags.HasErrors() {
		log.Fatal(diags)
	}
	conf.Hooks = hb.Hooks
	fmt.Println("%% Initialize Hooks\n%% TIME", time.Since(ct))
	// fmt.Printf("2 hooksConfig: %#v\n", conf)
	// conf.Dump()

	// Setup mux handler for hook.ID

	/////
	// Receive request
	/////

	// create a mock request
	req := struct {
		Method   string `cty:"method"`
		Proto    string `cty:"proto"`
		Host     string `cty:"host"`
		RemoteIP string `cty:"remote_ip"`
		// TLS ...
	}{
		Method:   "POST",
		Proto:    "HTTP/1.0",
		Host:     "foo.br",
		RemoteIP: "1.2.3.254",
	}

	// convert request to cty value
	ctyReq, err := gocty.ToCtyValue(req, cty.Object(map[string]cty.Type{
		"method":    cty.String,
		"proto":     cty.String,
		"host":      cty.String,
		"remote_ip": cty.String,
	}))
	if err != nil {
		panic(err)
	}
	ctx.EvalContext.Variables["request"] = ctyReq

	ctx.EvalContext.Variables["payload"] = cty.StringVal(`{
		"a": "z",
		"ref":            "refs/heads/master",
		"head_commit": {
			"id": "abcdf"
		},
		"pusher": {
			"name":    "John Q Public",
			"email":   "jqp@foo.br"
		},
		"newVolume":      50,
		"previousVolume": 80
	}`)
	// "zippedBinary": cty.StringVal("iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAA2lpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuMC1jMDYwIDYxLjEzNDc3NywgMjAxMC8wMi8xMi0xNzozMjowMCAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wUmlnaHRzPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvcmlnaHRzLyIgeG1sbnM6eG1wTU09Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9tbS8iIHhtbG5zOnN0UmVmPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvc1R5cGUvUmVzb3VyY2VSZWYjIiB4bWxuczp4bXA9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC8iIHhtcFJpZ2h0czpNYXJrZWQ9IkZhbHNlIiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOjEzMTA4RDI0QzMxQjExRTBCMzYzRjY1QUQ1Njc4QzFBIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOjEzMTA4RDIzQzMxQjExRTBCMzYzRjY1QUQ1Njc4QzFBIiB4bXA6Q3JlYXRvclRvb2w9IkFkb2JlIFBob3Rvc2hvcCBDUzMgV2luZG93cyI+IDx4bXBNTTpEZXJpdmVkRnJvbSBzdFJlZjppbnN0YW5jZUlEPSJ1dWlkOkFDMUYyRTgzMzI0QURGMTFBQUI4QzUzOTBEODVCNUIzIiBzdFJlZjpkb2N1bWVudElEPSJ1dWlkOkM5RDM0OTY2NEEzQ0REMTFCMDhBQkJCQ0ZGMTcyMTU2Ii8+IDwvcmRmOkRlc2NyaXB0aW9uPiA8L3JkZjpSREY+IDwveDp4bXBtZXRhPiA8P3hwYWNrZXQgZW5kPSJyIj8+IBFgEwAAAmJJREFUeNqkk89rE1EQx2d/NNq0xcYYayPYJDWC9ODBsKIgAREjBmvEg2cvHnr05KHQ9iB49SL+/BMEfxBQKHgwCEbTNNIYaqgaoanFJi+rcXezye4689jYkIMIDnx47837zrx583YFx3Hgf0xA6/dJyAkkgUy4vgryAnmNWH9L4EVmotFoKplMHgoGg6PkrFarjXQ6/bFcLj/G5W1E+3NaX4KZeDx+dX5+7kg4HBlmrC6JoiDFYrGhROLM/mp1Y6JSqdCd3/SW0GUqEAjkl5ZyHTSHKBQKnO6a9khD2m5cr91IJBJ1VVWdiM/n6LruNJtNDs3JR3ukIW03SHTHi8iVsbG9I51OG1bW16HVasHQZopDc/JZVgdIQ1o3BmTkEnJXURS/KIpgGAYPkCQJPi0u8uzDKQN0XQPbtgE1MmrHs9nsfSqAEjxCNtHxZHLy4G4smUQgyzL4LzOegDGGp1ucVqsNqKVrpJCM7F4hg6iaZvhqtZrg8XjA4xnAU3XeKLqWaRImoIZeQXVjQO5pYp4xNVirsR1erxer2O4yfa227WCwhtWoJmn7m0h270NxmemFW4706zMm8GCgxBGEASCfhnukIW03iFdQnOPz0LNKp3362JqQzSw4u2LXBe+Bs3xD+/oc1NxN55RiC9fOme0LEQiRf2rBzaKEeJJ37ZWTVunBeGN2WmQjg/DeLTVP89nzAive2dMwlo9bpFVC2xWMZr+A720FVn88fAUb3wDMOjyN7YNc6TvUSHQ4AH6TOUdLL7em68UtWPsJqxgTpgeiLu1EBt1R+Me/mF7CQPTfAgwAGxY2vOTrR3oAAAAASUVORK5CYII="),

	// TODO: only support map[string]string values right now.
	ctx.Payload = map[string]interface{}{
		"a":              "z",
		"ref":            "refs/heads/master",
		"head_commit.id": "abcdf",
		"pusher.name":    "John Q Public",
		"pusher.email":   "jqp@foo.br",
		"previousvolume": 50,
		"newvolume":      80,
	}
	ctx.Headers = map[string]string{
		"x-signature":       "f417af3a21bd70379b5796d5f013915e7029f62c580fb0f500f59a35a6f04c89",
		"x-coral-signature": "sha1=b17e04cbb22afa8ffbff8796fc1894ed27badd9e,sha256=f417af3a21bd70379b5796d5f013915e7029f62c580fb0f500f59a35a6f04c89",
		"x-forwarded-for":   "1.2.3.4",
		"date":              "Fri, 20 Sep 2019 14:09:11 GMT",
		"authorization":     "HMAC f417af3a21bd70379b5796d5f013915e7029f62c580fb0f500f59a35a6f04c89",
	}
	ctx.Params = map[string]string{"param1": "foo"}

	/////
	// Evaluate constraints and task block
	/////

	// fmt.Printf(".  %T\n", conf.Hooks[0].PreExecConfig)
	// synb := conf.Hooks[0].PreExecConfig.(*hclsyntax.Body)
	// for k, v := range synb.Attributes {
	// 	fmt.Printf(".  .  %T\n", v.Expr)
	// 	fmt.Printf(".  .  %s = %v\n", k, v.Expr)
	//
	// 	tce := v.Expr.(*hclsyntax.TupleConsExpr)
	// 	for _, vv := range tce.ExprList() {
	// 		fmt.Printf(".  .  .  %T\n", vv)
	// 		fmt.Printf(".  .  .  %v\n", vv)
	//
	// 		fce := vv.(*hclsyntax.FunctionCallExpr)
	// 		for _, vvv := range fce.Args {
	// 			fmt.Printf(".  .  .  .  %T\n", vvv)
	// 			fmt.Printf(".  .  .  .  %v\n", vvv)
	// 			// XXX
	// 		}
	// 	}
	// }

	ct = time.Now()
	var pre config.PreExecConfig
	diags = gohcl.DecodeBody(conf.Hooks[0].PreExecConfig, ctx.EvalContext, &pre)
	if diags.HasErrors() {
		log.Fatal(diags)
	}
	conf.Hooks[0].Constraints = pre.Constraints
	conf.Hooks[0].Task = pre.Task
	fmt.Println("%% Evaluate Constraints\n%% TIME", time.Since(ct))
	// fmt.Printf("3 hookConfig: %#v\n", conf)
	conf.Dump()

	satisfied := true
	if conf.Hooks[0].Constraints != nil {
		for _, v := range *conf.Hooks[0].Constraints {
			if v == false {
				satisfied = false
			}
		}
	}

	/////
	// Execute task, if necessary
	/////

	ctx.EvalContext.Variables["result"] = cty.ObjectVal(map[string]cty.Value{
		"exit_code":      cty.NumberIntVal(11),
		"error":          cty.BoolVal(true),
		"pid":            cty.NumberIntVal(12345),
		"CombinedOutput": cty.StringVal(`{"error":12,"output":"connection refused"}`),
	})

	if !satisfied {
		fmt.Println("hook constraints not satisfied.")
	}

	/////
	// Send Response
	/////

	ct = time.Now()
	var post config.PostExecConfig
	// TODO: if hook constraints are unsatisfied, the $result will be nil, so we
	// don't want to render the success or error blocks.
	diags = gohcl.DecodeBody(pre.PostExecConfig, ctx.EvalContext, &post)
	if diags.HasErrors() {
		log.Fatal(diags)
	}
	conf.Hooks[0].Response = post.Response
	fmt.Println("%% Build Response\n%% TIME", time.Since(ct))
	// fmt.Printf("4 hookConfig: %#v\n", conf)
	// conf.Dump()

	fmt.Println("%% TOTAL TIME", time.Since(t0))
	fmt.Println("%% TOTAL TIME LESS LOAD CONFIG", time.Since(t1))
}

func loadConfigFile(path string) (config.Service, error) {
	_, err := os.Stat(path)
	if err != nil {
		return config.Service{}, err
	}

	p := hclparse.NewParser()

	f, diags := p.ParseHCLFile(path)
	if diags.HasErrors() {
		return config.Service{}, diags
	}

	ctx := config.NewContext()

	var svc config.Service
	diags = gohcl.DecodeBody(f.Body, ctx.EvalContext, &svc)
	if diags.HasErrors() {
		return config.Service{}, diags
	}

	return svc, nil
}
