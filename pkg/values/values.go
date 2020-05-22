package values

type (
	Values map[string]interface{}
)

func (v *Values) Add(key string, content interface{}) {
	(*v)[key] = content
}
