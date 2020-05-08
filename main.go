package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mmcdole/gofeed"
	"github.com/open-integration/core"
	"github.com/open-integration/core/pkg/state"
	"github.com/open-integration/core/pkg/task"
	"github.com/open-integration/service-catalog/http/pkg/endpoints/call"
)

type (
	taskCandidate struct {
		target  Target
		binding Binding
		rss     RSS
		items   []gofeed.Item
	}
)

func main() {
	p := os.Getenv("FEED")
	if p == "" {
		dieOnError(fmt.Errorf("File not provided"))
	}
	cnf, err := readFile(p)
	dieOnError(err)
	condition := &TaskFinished{}

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
	}
	pipe := core.Pipeline{
		Metadata: core.PipelineMetadata{
			Name: "rss-sync",
		},
		Spec: core.PipelineSpec{
			Services: services,
			Reactions: []core.EventReaction{
				{
					Condition: core.ConditionEngineStarted(),
					Reaction: func(ev state.Event, state state.State) []task.Task {
						tasks := []task.Task{}
						for _, binding := range cnf.Bindings {
							for _, rss := range cnf.RSS {
								if rss.Name != binding.RSS {
									continue
								}
								u, err := buildURL(rss)
								if err != nil {
									return nil
								}
								name := buildTaskName(binding)
								condition.AddTask(name)
								tasks = append(tasks, task.Task{
									Metadata: task.Metadata{
										Name: name,
									},
									Spec: task.Spec{
										Service:  "http",
										Endpoint: "call",
										Arguments: []task.Argument{
											{
												Key:   "URL",
												Value: u,
											},
											{
												Key:   "Verb",
												Value: "GET",
											},
										},
									},
								})
							}

						}
						return tasks
					},
				},
				{
					Condition: condition,
					Reaction: func(ev state.Event, state state.State) []task.Task {
						tasks := []task.Task{}
						res := &call.CallReturns{}
						err := json.Unmarshal([]byte(state.Tasks()[ev.Metadata.Task].Output), res)
						if err != nil {
							fmt.Println(err.Error())
							return nil
						}
						fp := gofeed.NewParser()
						feed, err := fp.ParseString(res.Body)
						if err != nil {
							fmt.Println(err.Error())
							return nil
						}
						taskCandidate := taskCandidate{}
						for _, item := range feed.Items {
							name := getBindingNameFromTaskName(ev.Metadata.Task)
							var rssname string
							var targetname string
							for _, b := range cnf.Bindings {
								if b.Name != name {
									continue
								}
								rssname = b.RSS
								targetname = b.Target
								taskCandidate.binding = b
							}

							for _, t := range cnf.Targets {
								if t.Name != targetname {
									continue
								}
								taskCandidate.target = t
							}

							var rss RSS
							for _, r := range cnf.RSS {
								if r.Name != rssname {
									continue
								}
								rss = r
								taskCandidate.rss = r
							}
							matched := true
							for _, f := range rss.Filter {
								if res := filter(*item, f); !res {
									matched = false
								}
							}
							if !matched {
								continue
							}
							taskCandidate.items = append(taskCandidate.items, *item)
						}
						targetValues := targetToJSON(taskCandidate.target)
						bindingValues := bindingToJSON(taskCandidate.binding)
						rssValues := rssToJSON(taskCandidate.rss)
						feedValues := feedToJSON(*feed)
						for i, item := range taskCandidate.items {
							root := map[string]interface{}{}
							root["item"] = gofeedItemToJSON(item)
							root["rss"] = rssValues
							root["binding"] = bindingValues
							root["target"] = targetValues
							root["feed"] = feedValues
							tasks = append(tasks, task.Task{
								Metadata: task.Metadata{
									Name: fmt.Sprintf("%d-created-card-%s", i, item.Title),
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
											Value: templateString(taskCandidate.target.Trello.Card.Title, root),
										},
										{
											Key:   "Description",
											Value: templateString(taskCandidate.target.Trello.Card.Description, root),
										},
										{
											Key:   "Labels",
											Value: templateStringArray(taskCandidate.target.Trello.Card.Labels),
										},
									},
								},
							})
						}
						return tasks
					},
				},
			},
		},
	}
	e := core.NewEngine(&core.EngineOptions{
		Pipeline: pipe,
	})
	core.HandleEngineError(e.Run())
}
