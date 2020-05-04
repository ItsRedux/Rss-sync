package main

type (
	Config struct {
		Targets  []Target  `yaml:"targets"`
		RSS      []RSS     `yaml:"rss"`
		Bindings []Binding `yaml:"bindings"`
	}

	Target struct {
		Name   string `yaml:"name"`
		Trello *struct {
			Token         string `yaml:"token"`
			ApplicationID string `yaml:"application-id"`
			BoardID       string `yaml:"board-id"`
			ListID        string `yaml:"list-id"`
			Card          *struct {
				Title       *string  `yaml:"title,omitempty"`
				Description *string  `yaml:"description,omitempty"`
				Labels      []string `yaml:"labels"`
			} `yaml:"card,omitempty"`
		} `yaml:"trello,omitempty"`
	}

	RSS struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
		Auth *struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"auth"`
		Filter map[string]string `yaml:"filter"`
	}

	Binding struct {
		Name   string `yaml:"name"`
		RSS    string `yaml:"rss"`
		Target string `yaml:"target"`
	}
)
