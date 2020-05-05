package main

type (
	Config struct {
		Targets  []Target  `json:"targets" yaml:"targets"`
		RSS      []RSS     `json:"rss" yaml:"rss"`
		Bindings []Binding `json:"bindings" yaml:"bindings"`
	}

	Target struct {
		Name   string `json:"name" yaml:"name"`
		Trello *struct {
			Token         string `json:"token" yaml:"token"`
			ApplicationID string `json:"application-id" yaml:"application-id"`
			BoardID       string `json:"board-id" yaml:"board-id"`
			ListID        string `json:"list-id" yaml:"list-id"`
			Card          *struct {
				Title       *string  `json:"title,omitempty" yaml:"title,omitempty"`
				Description *string  `json:"description,omitempty" yaml:"description,omitempty"`
				Labels      []string `json:"labels" yaml:"labels"`
			} `json:"card,omitempty" yaml:"card,omitempty"`
		} `json:"trello,omitempty" yaml:"trello,omitempty"`
	}

	RSS struct {
		Name string `json:"name" yaml:"name"`
		URL  string `json:"url" yaml:"url"`
		Auth *struct {
			Username string `json:"username" yaml:"username"`
			Password string `json:"password" yaml:"password"`
		} `json:"auth" yaml:"auth"`
		Filter map[string]string `json:"filter" yaml:"filter"`
	}

	Binding struct {
		Name   string `json:"name" yaml:"name"`
		RSS    string `json:"rss" yaml:"rss"`
		Target string `json:"target" yaml:"target"`
	}
)
