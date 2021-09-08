// from github.com/zclconf/go-cty/cty/function/stdlib

package config

import (
	"fmt"
	"regexp"
	"regexp/syntax"

	"github.com/zclconf/go-cty/cty"
)

// regexPatternResultType parses the given regular expression pattern and
// returns the structural type that would be returned to represent its
// capture groups.
//
// Returns an error if parsing fails or if the pattern uses a mixture of
// named and unnamed capture groups, which is not permitted.
func regexPatternResultType(pattern string) (cty.Type, error) {
	re, rawErr := regexp.Compile(pattern)
	switch err := rawErr.(type) {
	case *syntax.Error:
		return cty.NilType, fmt.Errorf("invalid regexp pattern: %s in %s", err.Code, err.Expr)
	case error:
		// Should never happen, since all regexp compile errors should
		// be resyntax.Error, but just in case...
		return cty.NilType, fmt.Errorf("error parsing pattern: %s", err)
	}

	allNames := re.SubexpNames()[1:]
	var names []string
	unnamed := 0
	for _, name := range allNames {
		if name == "" {
			unnamed++
		} else {
			if names == nil {
				names = make([]string, 0, len(allNames))
			}
			names = append(names, name)
		}
	}
	switch {
	case unnamed == 0 && len(names) == 0:
		// If there are no capture groups at all then we'll return just a
		// single string for the whole match.
		return cty.String, nil
	case unnamed > 0 && len(names) > 0:
		return cty.NilType, fmt.Errorf("invalid regexp pattern: cannot mix both named and unnamed capture groups")
	case unnamed > 0:
		// For unnamed captures, we return a tuple of them all in order.
		etys := make([]cty.Type, unnamed)
		for i := range etys {
			etys[i] = cty.String
		}
		return cty.Tuple(etys), nil
	default:
		// For named captures, we return an object using the capture names
		// as keys.
		atys := make(map[string]cty.Type, len(names))
		for _, name := range names {
			atys[name] = cty.String
		}
		return cty.Object(atys), nil
	}
}

func regexPatternResult(re *regexp.Regexp, str string, captureIdxs []int, retType cty.Type) cty.Value {
	switch {
	case retType == cty.String:
		start, end := captureIdxs[0], captureIdxs[1]
		return cty.StringVal(str[start:end])
	case retType.IsTupleType():
		captureIdxs = captureIdxs[2:] // index 0 is the whole pattern span, which we ignore by skipping one pair
		vals := make([]cty.Value, len(captureIdxs)/2)
		for i := range vals {
			start, end := captureIdxs[i*2], captureIdxs[i*2+1]
			if start < 0 || end < 0 {
				vals[i] = cty.NullVal(cty.String) // Did not match anything because containing group didn't match
				continue
			}
			vals[i] = cty.StringVal(str[start:end])
		}
		return cty.TupleVal(vals)
	case retType.IsObjectType():
		captureIdxs = captureIdxs[2:] // index 0 is the whole pattern span, which we ignore by skipping one pair
		vals := make(map[string]cty.Value, len(captureIdxs)/2)
		names := re.SubexpNames()[1:]
		for i, name := range names {
			start, end := captureIdxs[i*2], captureIdxs[i*2+1]
			if start < 0 || end < 0 {
				vals[name] = cty.NullVal(cty.String) // Did not match anything because containing group didn't match
				continue
			}
			vals[name] = cty.StringVal(str[start:end])
		}
		return cty.ObjectVal(vals)
	default:
		// Should never happen
		panic(fmt.Sprintf("invalid return type %#v", retType))
	}
}
