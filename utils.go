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

func readFile(location string) (Config, error) {
	bytes, err := ioutil.ReadFile(location)
	if err != nil {
		return Config{}, err
	}
	cnf := Config{}
	if err := yaml.Unmarshal(bytes, &cnf); err != nil {
		return cnf, err
	}
	return cnf, nil
}

func dieOnError(err error) {
	if err != nil {
		fmt.Printf("[Error] %s", err.Error())
		os.Exit(1)
	}
}

func buildURL(rss RSS) (string, error) {
	u, err := url.Parse(rss.URL)
	if err != nil {
		return "", err
	}
	if rss.Auth != nil {
		u.User = url.UserPassword(templateString(&rss.Auth.Username, nil), templateString(&rss.Auth.Password, nil))
	}
	return u.String(), nil
}

func filter(item *gofeed.Item, filter string) bool {
	root := map[string]interface{}{}
	root["item"] = gofeedItemToJSON(item)
	out := templateString(&filter, root)
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
	return fmt.Sprintf("%s%s%s", binding.Name, seperator, binding.RSS)
}

func getBindingNameFromTaskName(name string) string {
	return strings.Split(name, seperator)[0]
}

func gofeedItemToJSON(item *gofeed.Item) map[string]interface{} {
	data := map[string]interface{}{}
	b, err := json.Marshal(item)
	if err != nil {
		return data
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return data
	}
	return data
}