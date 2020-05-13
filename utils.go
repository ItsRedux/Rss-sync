package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/hairyhenderson/gomplate"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v2"
)

const (
	seperator = ":::"
)

func readFile(location string) (Sync, error) {
	bytes, err := ioutil.ReadFile(location)
	if err != nil {
		return Sync{}, err
	}
	cnf := Sync{}
	if err := yaml.Unmarshal(bytes, &cnf); err != nil {
		return cnf, err
	}
	return cnf, nil
}

func dieOnError(err error) {
	if err != nil {
		fmt.Printf("[Error] %s\n", err.Error())
		os.Exit(1)
	}
}

func buildURL(URL string, username string, password string) (string, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", err
	}
	if username != "" && password != "" {
		u.User = url.UserPassword(templateString(&username, nil), templateString(&password, nil))
	}
	return u.String(), nil
}

func filter(data interface{}, filter string) bool {
	out := templateString(&filter, data)
	return out == "true"
}

func templateString(tmpl *string, data interface{}) string {
	if tmpl == nil {
		fmt.Println("nil input")
		return ""
	}
	out := new(bytes.Buffer)
	funcs := gomplate.Funcs(nil)
	template.Must(template.New(*tmpl).Funcs(funcs).Parse(*tmpl)).Execute(out, data)
	return out.String()
}

func templateStringArray(tmpls []string) []string {
	res := []string{}
	for _, tmpl := range tmpls {
		res = append(res, templateString(&tmpl, nil))
	}

	return res
}

func buildTaskName(binding Binding) string {
	return fmt.Sprintf("%s%s%s", binding.Name, seperator, binding.Source)
}

func getBindingNameFromTaskName(name string) string {
	return strings.Split(name, seperator)[0]
}

func gofeedItemToJSON(item gofeed.Item) map[string]interface{} {
	b, err := json.Marshal(item)
	if err != nil {
		return nil
	}
	return toJSON(b)
}

func srcToJSON(src Source) map[string]interface{} {
	b, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	return toJSON(b)
}

func bindingToJSON(binding Binding) map[string]interface{} {
	b, err := json.Marshal(binding)
	if err != nil {
		return nil
	}
	return toJSON(b)
}

func targetToJSON(target Target) map[string]interface{} {
	b, err := json.Marshal(target)
	if err != nil {
		return nil
	}
	return toJSON(b)
}

func feedToJSON(feed gofeed.Feed) map[string]interface{} {
	feed.Items = []*gofeed.Item{}
	b, err := json.Marshal(feed)
	if err != nil {
		return nil
	}
	return toJSON(b)
}

func toJSON(input []byte) map[string]interface{} {
	data := map[string]interface{}{}
	if err := json.Unmarshal(input, &data); err != nil {
		return data
	}
	return data
}
func toArrayJSON(input []byte) []map[string]interface{} {
	data := []map[string]interface{}{}
	if err := json.Unmarshal(input, &data); err != nil {
		return data
	}
	return data
}
