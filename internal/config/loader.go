// +build ignore
package config

import "github.com/hashicorp/hcl/v2/hclparse"

type hcl2Loader struct {
	Parser *hclparse.Parser
}
