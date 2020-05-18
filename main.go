package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/open-integration/core"
	"github.com/open-integration/core/pkg/state"
	"github.com/open-integration/core/pkg/task"
	"github.com/open-integration/service-catalog/http/pkg/endpoints/call"
	"github.com/open-integration/service-catalog/jira/pkg/endpoints/list"
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
)

func main() {
	syncs := readSyncFiles()
	for _, cnf := range syncs {
		conditionRSSTaskFinished := &TaskFinished{}
		conditionJSONTaskFinished := &TaskFinished{}
		conditionJIRATaskFinished := &TaskFinished{}
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
						Reaction: func(ev state.Event, state state.State) []task.Task {
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
										jql:      templateString(&src.JIRA.JQL, nil),
										taskName: name,
										token:    templateString(&src.JIRA.Token, nil),
										user:     templateString(&src.JIRA.User, nil),
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
				},
			},
		}
		e := core.NewEngine(&core.EngineOptions{
			Pipeline: pipe,
		})
		core.HandleEngineError(e.Run())
	}

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

func reactToRSSCompletedTask(cnf Sync) func(ev state.Event, state state.State) []task.Task {
	return func(ev state.Event, state state.State) []task.Task {
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

func reactToJSONCompletedTask(cnf Sync) func(ev state.Event, state state.State) []task.Task {
	return func(ev state.Event, state state.State) []task.Task {
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

func reactToJIRACompletedTask(cnf Sync) func(ev state.Event, state state.State) []task.Task {
	return func(ev state.Event, state state.State) []task.Task {
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
					Value: templateString(&taskCandidate.target.Trello.Key, nil),
				},
				{
					Key:   "Token",
					Value: templateString(&taskCandidate.target.Trello.Token, nil),
				},
				{
					Key:   "Board",
					Value: templateString(&taskCandidate.target.Trello.BoardID, nil),
				},
				{
					Key:   "List",
					Value: templateString(&taskCandidate.target.Trello.ListID, nil),
				},
				{
					Key:   "Name",
					Value: templateString(taskCandidate.target.Trello.Card.Title, data),
				},
				{
					Key:   "Description",
					Value: templateString(taskCandidate.target.Trello.Card.Description, data),
				},
				{
					Key:   "Labels",
					Value: templateStringArray(taskCandidate.target.Trello.Card.Labels),
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

func buildValues(taskCandidate taskCandidate) *values {
	targetValues := targetToJSON(taskCandidate.target)
	bindingValues := bindingToJSON(taskCandidate.binding)
	srcValues := srcToJSON(taskCandidate.src)
	root := &values{}
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

func readSyncFiles() []Sync {
	result := []Sync{}
	feed := os.Getenv("FEED")
	if feed != "" {
		cnf, err := readFile(feed)
		dieOnError(err)
		result = append(result, cnf)
	}

	config := os.Getenv("SYNCCONFIG")
	if config != "" {
		files := strings.Split(config, ";")
		for _, f := range files {
			cnf, err := readFile(f)
			dieOnError(err)
			result = append(result, cnf)
		}
	}

	if len(result) == 0 {
		dieOnError(fmt.Errorf("File not provided"))
	}

	return result
}
