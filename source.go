package main

type (
	Config struct {
		Targets  []Target  `json:"targets" yaml:"targets"`
		Sources  []Source  `json:"sources" yaml:"sources"`
		Bindings []Binding `json:"bindings" yaml:"bindings"`
	}

	Target struct {
		Name   string `json:"name" yaml:"name"`
		Trello *struct {
			Token   string `json:"token" yaml:"token"`
			Key     string `json:"key" yaml:"key"`
			BoardID string `json:"board-id" yaml:"board-id"`
			ListID  string `json:"list-id" yaml:"list-id"`
			Card    *struct {
				Title       *string  `json:"title,omitempty" yaml:"title,omitempty"`
				Description *string  `json:"description,omitempty" yaml:"description,omitempty"`
				Labels      []string `json:"labels" yaml:"labels"`
			} `json:"card,omitempty" yaml:"card,omitempty"`
		} `json:"trello,omitempty" yaml:"trello,omitempty"`
	}

	Source struct {
		Name string `json:"name" yaml:"name"`
		RSS  *struct {
			URL  string `json:"url" yaml:"url"`
			Auth *struct {
				Username string `json:"username" yaml:"username"`
				Password string `json:"password" yaml:"password"`
			} `json:"auth" yaml:"auth"`
		} `json:"rss,omitempty" yaml:"rss,omitempty"`
		Filter map[string]string `json:"filter" yaml:"filter"`
	}

	Binding struct {
		Name   string `json:"name" yaml:"name"`
		Source string `json:"source" yaml:"source"`
		Target string `json:"target" yaml:"target"`
	}
)
