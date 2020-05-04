package main

import (
	"github.com/open-integration/core/pkg/state"
)

type (
	TaskFinished struct {
		followTasks []string
	}
)

func (c *TaskFinished) Met(ev state.Event, s state.State) bool {
	met := false
	for _, t := range c.followTasks {
		if t == ev.Metadata.Task {
			met = true
		}
	}
	if !met {
		return false
	}
	return ev.Metadata.Name == state.EventTaskFinished
}

func (c *TaskFinished) AddTask(name string) {
	c.followTasks = append(c.followTasks, name)
}
