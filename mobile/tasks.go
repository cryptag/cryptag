package mobile

/*
type Counter struct {
	Value int
}

func (c *Counter) Inc() { c.Value++ }

func New() *Counter { return &Counter{5} }
*/

func NewTask(id int, description, assignee string) *Task {
	return &Task{
		ID:          id,
		Description: description,
		Assignee:    assignee,
	}
}

type Task struct {
	ID          int
	Description string
	Assignee    string
}
