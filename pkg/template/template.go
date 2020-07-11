package template

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/hairyhenderson/gomplate"
)

func StartDay(format string) string {
	t := time.Now()
	year, month, day := t.Date()
	res := time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Format(format)
	return res
}

func EndDay(format string) string {
	t := time.Now()
	year, month, day := t.Date()
	res := time.Date(year, month, day, 23, 59, 59, 0, t.Location()).Format(format)
	return res
}

// String executes template passing data as variables
func String(tmpl *string, data interface{}) string {
	if tmpl == nil {
		fmt.Println("nil input")
		return ""
	}
	out := new(bytes.Buffer)
	funcs := gomplate.Funcs(nil)
	funcs["StartDay"] = StartDay
	funcs["EndDay"] = EndDay
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
