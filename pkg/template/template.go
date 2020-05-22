package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/hairyhenderson/gomplate"
)

// String executes template passing data as variables
func String(tmpl *string, data interface{}) string {
	if tmpl == nil {
		fmt.Println("nil input")
		return ""
	}
	out := new(bytes.Buffer)
	funcs := gomplate.Funcs(nil)
	template.Must(template.New(*tmpl).Funcs(funcs).Parse(*tmpl)).Execute(out, data)
	return out.String()
}

// StringArray executes template on each on the items
func StringArray(tmpls []string) []string {
	res := []string{}
	for _, tmpl := range tmpls {
		res = append(res, String(&tmpl, nil))
	}

	return res
}
