package mobile

/*
type Counter struct {
	Value int
}

func (c *Counter) Inc() { c.Value++ }

func New() *Counter { return &Counter{5} }
*/

func NewTask(id int, title, description, assignee string) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Assignee:    assignee,
	}
}

type Task struct {
	ID          int
	Title       string
	Description string
	Assignee    string
}
