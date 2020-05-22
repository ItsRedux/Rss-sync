package cmd

// Copyright © 2020 oleg2807@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/olegsu/rss-sync/pkg/template"
	"github.com/open-integration/service-catalog/jira/pkg/endpoints/list"
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
	u, err := url.Parse(template.String(&URL, nil))
	if err != nil {
		return "", err
	}
	if username != "" && password != "" {
		u.User = url.UserPassword(template.String(&username, nil), template.String(&password, nil))
	}
	return u.String(), nil
}

func filter(data interface{}, filter string) bool {
	out := template.String(&filter, data)
	return out == "true"
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

func jiraIssueToJSON(issue list.Issue) map[string]interface{} {
	b, err := json.Marshal(issue)
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
