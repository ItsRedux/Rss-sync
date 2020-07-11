package cmd

import "errors"

var (
	errNotFound = errors.New("Not found")
)

type (
	Sync struct {
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
		JSON *struct {
			URL  string `json:"url" yaml:"url"`
			Type string `json:"type" yaml:"type"`
		} `json:"json,omitempty" yaml:"json,omitempty"`
		JIRA *struct {
			User     string `json:"user" yaml:"user"`
			Token    string `json:"token" yaml:"token"`
			Endpoint string `json:"endpoint" yaml:"endpoint"`
			JQL      string `json:"jql" yaml:"jql"`
		} `json:"jira,omitempty" yaml:"jira,omitempty"`
		GoogleCalendar *struct {
			ServiceAccount string `json:"service-account" yaml:"service-account"`
			CalendarID     string `json:"calendar-id" yaml:"calendar-id"`
			TimeMin        string `json:"time-min" yaml:"time-min"`
			TimeMax        string `json:"time-max" yaml:"time-max"`
		} `json:"google-calendar" yaml:"google-calendar"`
		Filter map[string]string `json:"filter" yaml:"filter"`
	}

	Binding struct {
		Name   string `json:"name" yaml:"name"`
		Source string `json:"source" yaml:"source"`
		Target string `json:"target" yaml:"target"`
	}
)

func getBinding(name string, list []Binding) (Binding, error) {
	for _, b := range list {
		if b.Name == name {
			return b, nil
		}
	}

	return Binding{}, errNotFound
}

func getSource(name string, list []Source) (Source, error) {
	for _, b := range list {
		if b.Name == name {
			return b, nil
		}
	}

	return Source{}, errNotFound
}

func getTarget(name string, list []Target) (Target, error) {
	for _, b := range list {
		if b.Name == name {
			return b, nil
		}
	}
	return Target{}, errNotFound
}
