package main

type (
	values map[string]interface{}
)

func (v *values) Add(key string, content interface{}) {
	(*v)[key] = content
}
