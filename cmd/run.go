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
	"path"

	"github.com/mmcdole/gofeed"
	"github.com/olegsu/rss-sync/pkg/template"
	"github.com/olegsu/rss-sync/pkg/values"
	"github.com/open-integration/core"
	"github.com/open-integration/core/pkg/event"
	"github.com/open-integration/core/pkg/state"
	"github.com/open-integration/core/pkg/task"
	"github.com/open-integration/service-catalog/google-calendar/pkg/endpoints/getEvents"
	"github.com/open-integration/service-catalog/http/pkg/endpoints/call"
	"github.com/open-integration/service-catalog/jira/pkg/endpoints/list"
	"github.com/spf13/cobra"
)

var (
	runCmdOptions struct {
		files []string
	}
)

type (
	taskCandidate struct {
		target  Target
		binding Binding
		src     Source
	}

	createJiraTaskOptions struct {
		taskName string
		token    string
		endpoint string
		user     string
		jql      string
	}

	createGoogleCalendarTaskOptions struct {
		taskName                string
		ServiceAccount          getEvents.ServiceAccount
		CalendarID              string
		ICalUID                 *string
		MaxAttendees            *int64
		MaxResults              *int64
		OrderBy                 *string
		PrivateExtendedProperty *string
		Q                       *string
		SharedExtendedProperty  *string
		ShowDeleted             bool
		ShowHiddenInvitations   *bool
		SingleEvents            *bool
		TimeMax                 string
		TimeMin                 string
		TimeZone                *string
		UpdatedMin              *string
	}
)

const (
	rootContext = "Values"
)

var runCmd = &cobra.Command{
	Use:  "run",
	Long: "Start to sync",
	Run: func(cmd *cobra.Command, args []string) {
		syncs := readSyncFiles(runCmdOptions.files)
		for name, cnf := range syncs {
			fmt.Printf("Startin to run sync from file %s\n", name)
			conditionRSSTaskFinished := &TaskFinished{}
			conditionJSONTaskFinished := &TaskFinished{}
			conditionJIRATaskFinished := &TaskFinished{}
			conditionGoogleCalendarTaskFinished := &TaskFinished{}
			services := []core.Service{
				{
					As:      "http",
					Name:    "http",
					Version: "0.0.1",
				},
				{
					Name:    "trello",
					Version: "0.10.0",
					As:      "trello",
				},
				{
					Name:    "jira",
					Version: "0.1.0",
					As:      "jira",
				},
				{
					Name:    "google-calendar",
					Version: "0.0.3",
					As:      "google-calendar",
				},
			}
			pipe := core.Pipeline{
				Metadata: core.PipelineMetadata{
					Name: "sync",
				},
				Spec: core.PipelineSpec{
					Services: services,
					Reactions: []core.EventReaction{
						{
							Condition: core.ConditionEngineStarted(),
							Reaction: func(ev event.Event, state state.State) []task.Task {
								tasks := []task.Task{}
								for _, binding := range cnf.Bindings {
									src, err := getSource(binding.Source, cnf.Sources)
									if err != nil {
										dieOnError(fmt.Errorf("Source \"%s\" not found", binding.Source))
									}
									name := buildTaskName(binding)

									if src.RSS != nil {
										username, password := "", ""
										if src.RSS.Auth != nil {
											username = src.RSS.Auth.Username
											password = src.RSS.Auth.Password
										}
										u, err := buildURL(src.RSS.URL, username, password)
										dieOnError(err)
										conditionRSSTaskFinished.AddTask(name)
										tasks = append(tasks, buildHTTPTask(name, u))
										continue
									}

									if src.JSON != nil {
										u, err := buildURL(src.JSON.URL, "", "")
										dieOnError(err)
										conditionJSONTaskFinished.AddTask(name)
										tasks = append(tasks, buildHTTPTask(name, u))
										continue
									}

									if src.JIRA != nil {
										u, err := buildURL(src.JIRA.Endpoint, "", "")
										dieOnError(err)
										conditionJIRATaskFinished.AddTask(name)
										tasks = append(tasks, createJiraTask(createJiraTaskOptions{
											endpoint: u,
											jql:      template.String(&src.JIRA.JQL, nil),
											taskName: name,
											token:    template.String(&src.JIRA.Token, nil),
											user:     template.String(&src.JIRA.User, nil),
										}))
										continue
									}

									if src.GoogleCalendar != nil {
										conditionGoogleCalendarTaskFinished.AddTask(name)
										f, err := ioutil.ReadFile(template.String(&src.GoogleCalendar.ServiceAccount, nil))
										dieOnError(err)
										sa := getEvents.ServiceAccount{}
										err = json.Unmarshal(f, &sa)
										dieOnError(err)
										tasks = append(tasks, createGoogleCalerndarTask(createGoogleCalendarTaskOptions{
											taskName:       name,
											ServiceAccount: sa,
											CalendarID:     template.String(&src.GoogleCalendar.CalendarID, nil),
											TimeMin:        template.String(&src.GoogleCalendar.TimeMin, nil),
											TimeMax:        template.String(&src.GoogleCalendar.TimeMax, nil),
											ShowDeleted:    true,
										}))
										continue
									}
								}
								return tasks
							},
						},
						{
							Condition: conditionRSSTaskFinished,
							Reaction:  reactToRSSCompletedTask(cnf),
						},
						{
							Condition: conditionJSONTaskFinished,
							Reaction:  reactToJSONCompletedTask(cnf),
						},
						{
							Condition: conditionJIRATaskFinished,
							Reaction:  reactToJIRACompletedTask(cnf),
						},
						{
							Condition: conditionGoogleCalendarTaskFinished,
							Reaction:  reactToGoogleCalendarCompletedTask(cnf),
						},
					},
				},
			}
			e := core.NewEngine(&core.EngineOptions{
				Pipeline: pipe,
			})
			core.HandleEngineError(e.Run())
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringArrayVarP(&runCmdOptions.files, "file", "f", nil, "Config file(s) that will be executed")
}

func buildHTTPTask(name string, url string) task.Task {
	return task.Task{
		Metadata: task.Metadata{
			Name: name,
		},
		Spec: task.Spec{
			Service:  "http",
			Endpoint: "call",
			Arguments: []task.Argument{
				{
					Key:   "URL",
					Value: url,
				},
				{
					Key:   "Verb",
					Value: "GET",
				},
			},
		},
	}
}

func reactToRSSCompletedTask(cnf Sync) func(ev event.Event, state state.State) []task.Task {
	return func(ev event.Event, state state.State) []task.Task {
		tasks := []task.Task{}
		res := &call.CallReturns{}
		err := json.Unmarshal([]byte(state.Tasks()[ev.Metadata.Task].Output), res)
		dieOnError(err)
		fp := gofeed.NewParser()
		feed, err := fp.ParseString(res.Body)
		dieOnError(err)
		taskCandidate := taskCandidate{}
		items := []gofeed.Item{}
		for _, item := range feed.Items {
			name := getBindingNameFromTaskName(ev.Metadata.Task)
			populateTaskCandidate(name, &taskCandidate, cnf)
			data := buildValues(taskCandidate)
			data.Add("item", gofeedItemToJSON(*item))
			if !filterSource(taskCandidate, data) {
				continue
			}
			items = append(items, *item)
		}
		feedValues := feedToJSON(*feed)
		for i, item := range items {
			root := buildValues(taskCandidate)
			root.Add("item", gofeedItemToJSON(item))
			root.Add("feed", feedValues)
			tasks = append(tasks, createTrelloTask(fmt.Sprintf("%d-created-card-%s", i, item.Title), taskCandidate, root))
		}
		return tasks
	}
}

func reactToJSONCompletedTask(cnf Sync) func(ev event.Event, state state.State) []task.Task {
	return func(ev event.Event, state state.State) []task.Task {
		tasks := []task.Task{}
		res := &call.CallReturns{}
		err := json.Unmarshal([]byte(state.Tasks()[ev.Metadata.Task].Output), res)
		dieOnError(err)

		taskCandidate := taskCandidate{}

		name := getBindingNameFromTaskName(ev.Metadata.Task)

		populateTaskCandidate(name, &taskCandidate, cnf)

		root := buildValues(taskCandidate)
		if taskCandidate.src.JSON.Type == "" || taskCandidate.src.JSON.Type == "object" {
			root.Add("content", toJSON([]byte(res.Body)))
			if !filterSource(taskCandidate, root) {
				return nil
			}
			tasks = append(tasks, createTrelloTask(fmt.Sprintf("%d-created-card-%s", 0, ""), taskCandidate, root))
		}

		if taskCandidate.src.JSON.Type == "array" {
			content := toArrayJSON([]byte(res.Body))
			for i, c := range content {
				root.Add("content", c)
				if !filterSource(taskCandidate, root) {
					continue
				}
				tasks = append(tasks, createTrelloTask(fmt.Sprintf("%d-created-card-%s", i, ""), taskCandidate, root))
			}
		}
		return tasks
	}
}

func reactToJIRACompletedTask(cnf Sync) func(ev event.Event, state state.State) []task.Task {
	return func(ev event.Event, state state.State) []task.Task {
		tasks := []task.Task{}
		res := &list.ListReturns{}
		err := json.Unmarshal([]byte(state.Tasks()[ev.Metadata.Task].Output), res)
		dieOnError(err)

		taskCandidate := taskCandidate{}

		name := getBindingNameFromTaskName(ev.Metadata.Task)

		populateTaskCandidate(name, &taskCandidate, cnf)

		root := buildValues(taskCandidate)
		for i, issue := range res.Issues {
			root.Add("issue", jiraIssueToJSON(issue))
			if !filterSource(taskCandidate, root) {
				return nil
			}
			tasks = append(tasks, createTrelloTask(fmt.Sprintf("%d-created-card-%s", i, name), taskCandidate, root))
		}
		return tasks
	}
}

func reactToGoogleCalendarCompletedTask(cnf Sync) func(ev event.Event, state state.State) []task.Task {
	return func(ev event.Event, state state.State) []task.Task {
		tasks := []task.Task{}
		res := &getEvents.GetEventsReturns{}
		err := json.Unmarshal([]byte(state.Tasks()[ev.Metadata.Task].Output), res)
		dieOnError(err)

		taskCandidate := taskCandidate{}

		name := getBindingNameFromTaskName(ev.Metadata.Task)

		populateTaskCandidate(name, &taskCandidate, cnf)

		root := buildValues(taskCandidate)
		for i, event := range res.Events {
			root.Add("event", googleCalendarEventToJSON(event))
			if !filterSource(taskCandidate, root) {
				return nil
			}
			tasks = append(tasks, createTrelloTask(fmt.Sprintf("%d-created-card-%s", i, name), taskCandidate, root))
		}
		return tasks
	}
}

func createTrelloTask(name string, taskCandidate taskCandidate, data interface{}) task.Task {
	return task.Task{
		Metadata: task.Metadata{
			Name: name,
		},
		Spec: task.Spec{
			Endpoint: "addcard",
			Service:  "trello",
			Arguments: []task.Argument{
				{
					Key:   "App",
					Value: template.String(&taskCandidate.target.Trello.Key, nil),
				},
				{
					Key:   "Token",
					Value: template.String(&taskCandidate.target.Trello.Token, nil),
				},
				{
					Key:   "Board",
					Value: template.String(&taskCandidate.target.Trello.BoardID, nil),
				},
				{
					Key:   "List",
					Value: template.String(&taskCandidate.target.Trello.ListID, nil),
				},
				{
					Key:   "Name",
					Value: template.String(taskCandidate.target.Trello.Card.Title, data),
				},
				{
					Key:   "Description",
					Value: template.String(taskCandidate.target.Trello.Card.Description, data),
				},
				{
					Key:   "Labels",
					Value: template.StringArray(taskCandidate.target.Trello.Card.Labels),
				},
			},
		},
	}
}

func createJiraTask(options createJiraTaskOptions) task.Task {
	return task.Task{
		Metadata: task.Metadata{
			Name: options.taskName,
		},
		Spec: task.Spec{
			Service:  "jira",
			Endpoint: "list",
			Arguments: []task.Argument{
				{
					Key:   "API_Token",
					Value: options.token,
				},
				{
					Key:   "Endpoint",
					Value: options.endpoint,
				},
				{
					Key:   "User",
					Value: options.user,
				},
				{
					Key:   "JQL",
					Value: options.jql,
				},
				{
					Key:   "QueryFields",
					Value: "*all",
				},
			},
		},
	}
}

func createGoogleCalerndarTask(options createGoogleCalendarTaskOptions) task.Task {
	arguments := []task.Argument{
		{
			Key:   "ServiceAccount",
			Value: options.ServiceAccount,
		},
		{
			Key:   "CalendarID",
			Value: options.CalendarID,
		},
		{
			Key:   "ShowDeleted",
			Value: options.ShowDeleted,
		},
	}

	if options.ICalUID != nil {
		arguments = append(arguments, task.Argument{
			Key:   "ICalUID",
			Value: *options.ICalUID,
		})
	}

	if options.MaxAttendees != nil {
		arguments = append(arguments, task.Argument{
			Key:   "MaxAttendees",
			Value: *options.MaxAttendees,
		})
	}
	if options.MaxResults != nil {
		arguments = append(arguments, task.Argument{
			Key:   "MaxResults",
			Value: *options.MaxResults,
		})
	}
	if options.OrderBy != nil {
		arguments = append(arguments, task.Argument{
			Key:   "OrderBy",
			Value: *options.OrderBy,
		})
	}
	if options.PrivateExtendedProperty != nil {
		arguments = append(arguments, task.Argument{
			Key:   "PrivateExtendedProperty",
			Value: *options.PrivateExtendedProperty,
		})
	}
	if options.Q != nil {
		arguments = append(arguments, task.Argument{
			Key:   "Q",
			Value: *options.Q,
		})
	}
	if options.SharedExtendedProperty != nil {
		arguments = append(arguments, task.Argument{
			Key:   "SharedExtendedProperty",
			Value: *options.SharedExtendedProperty,
		})
	}
	if options.ShowHiddenInvitations != nil {
		arguments = append(arguments, task.Argument{
			Key:   "ShowHiddenInvitations",
			Value: *options.ShowHiddenInvitations,
		})
	}
	if options.SingleEvents != nil {
		arguments = append(arguments, task.Argument{
			Key:   "SingleEvents",
			Value: *options.SingleEvents,
		})
	}
	if options.TimeMax != "" {
		arguments = append(arguments, task.Argument{
			Key:   "TimeMax",
			Value: options.TimeMax,
		})
	}
	if options.TimeMin != "" {
		arguments = append(arguments, task.Argument{
			Key:   "TimeMin",
			Value: options.TimeMin,
		})
	}
	if options.TimeZone != nil {
		arguments = append(arguments, task.Argument{
			Key:   "TimeZone",
			Value: *options.TimeZone,
		})
	}
	if options.UpdatedMin != nil {
		arguments = append(arguments, task.Argument{
			Key:   "UpdatedMin",
			Value: *options.UpdatedMin,
		})
	}

	return task.Task{
		Metadata: task.Metadata{
			Name: options.taskName,
		},
		Spec: task.Spec{
			Service:   "google-calendar",
			Endpoint:  "getEvents",
			Arguments: arguments,
		},
	}
}

func buildValues(taskCandidate taskCandidate) *values.Values {
	targetValues := targetToJSON(taskCandidate.target)
	bindingValues := bindingToJSON(taskCandidate.binding)
	srcValues := srcToJSON(taskCandidate.src)
	root := &values.Values{}
	root.Add("source", srcValues)
	root.Add("binding", bindingValues)
	root.Add("target", targetValues)
	return root
}

func filterSource(taskCandidate taskCandidate, data interface{}) bool {
	matched := true
	for _, f := range taskCandidate.src.Filter {
		if res := filter(data, f); !res {
			matched = false
		}
	}
	return matched
}

func populateTaskCandidate(bindingname string, tc *taskCandidate, cnf Sync) error {
	binding, err := getBinding(bindingname, cnf.Bindings)
	if err != nil {
		return err
	}
	tc.binding = binding

	source, err := getSource(tc.binding.Source, cnf.Sources)
	if err != nil {
		return err
	}
	tc.src = source

	target, err := getTarget(tc.binding.Target, cnf.Targets)
	if err != nil {
		return err
	}

	tc.target = target
	return nil
}

func readSyncFiles(files []string) map[string]Sync {
	result := map[string]Sync{}
	for _, f := range files {
		cnf, err := readFile(f)
		dieOnError(err)
		result[path.Base(f)] = cnf
	}

	if len(result) == 0 {
		dieOnError(fmt.Errorf("File not provided"))
	}

	return result
}
