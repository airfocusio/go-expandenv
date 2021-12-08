package expandenv

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ExpandEnv(input interface{}) (interface{}, error) {
	singleRegex := regexp.MustCompile(`^\$\{[^\}]+\}$`)
	detectRegex := regexp.MustCompile(`(?:^|[^\\])\$\{[^\}]+\}`)
	var recursion func(current interface{}) (interface{}, []error)
	recursion = func(current interface{}) (interface{}, []error) {
		if current, ok := current.(string); ok {
			p := singleRegex.FindStringSubmatch(current)
			if p != nil {
				expanded, err := expandEnvValue(current)
				if err != nil {
					return current, []error{err}
				}
				return expanded, nil
			}
			errs := []error{}
			expanded := detectRegex.ReplaceAllStringFunc(current, func(str string) string {
				index := strings.IndexAny(str, "$")
				prefix := str[:index]
				value := str[index:]

				expanded, err := expandEnvValue(value)
				if err != nil {
					errs = append(errs, err)
					return str
				}

				return fmt.Sprintf("%s%v", prefix, expanded)
			})
			return expanded, errs
		}
		if current, ok := current.([]interface{}); ok {
			current2 := make([]interface{}, len(current))
			errs := []error{}
			for i := range current {
				v, err := recursion(current[i])
				if err != nil {
					errs = append(errs, err...)
				}
				current2[i] = v
			}
			return current2, errs
		}
		if current, ok := current.(map[string]interface{}); ok {
			errs := []error{}
			current2 := map[string]interface{}{}
			for k, v := range current {
				v, err := recursion(v)
				if err != nil {
					errs = append(errs, err...)
				}
				current2[k] = v
			}
			return current2, errs
		}
		return current, []error{}
	}
	output, errs := recursion(input)
	if len(errs) > 0 {
		errMsgs := []string{}
		for _, err := range errs {
			errMsgs = append(errMsgs, err.Error())
		}
		return output, fmt.Errorf(strings.Join(errMsgs, ", "))
	}
	return output, nil
}

func expandEnvValue(str string) (interface{}, error) {
	regex := regexp.MustCompile(`^\$\{(?P<name>[^:]+)(?::(?P<format>number|boolean|string))?(?::-(?P<fallback>.*))?\}$`)
	p := regex.FindStringSubmatch(str)
	if p == nil {
		return nil, fmt.Errorf("could not parse %s", str)
	}
	name := p[regex.SubexpIndex("name")]
	format := p[regex.SubexpIndex("format")]
	fallback := p[regex.SubexpIndex("fallback")]
	value, ok := os.LookupEnv(name)
	if !ok {
		if fallback == "" {
			return nil, fmt.Errorf("environment variable %s is missing", name)
		} else {
			value = fallback
		}
	}

	switch format {
	case "":
		return value, nil
	case "string":
		return value, nil
	case "number":
		formatted, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid number", value)
		}
		return formatted, nil
	case "boolean":
		switch value {
		case "0":
			return false, nil
		case "1":
			return true, nil
		case "false":
			return false, nil
		case "true":
			return true, nil
		case "no":
			return false, nil
		case "yes":
			return true, nil
		default:
			return nil, fmt.Errorf("%s is not a valid boolean", value)
		}
	default:
		return nil, fmt.Errorf("format %s is not supported", format)
	}
}
