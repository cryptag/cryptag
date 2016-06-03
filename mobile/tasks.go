package mobile

import (
	"fmt"
	"sync"
)

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

func (t *Task) IDStr() string {
	return fmt.Sprintf("%v", t.ID)
}

func (t *Task) String() string {
	if t == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%#v", t)
}

type TaskGetter struct {
	tasks []*Task
}

var allTasks = []*Task{
	{1, "Show demo CrypTask Android app to Sky!", "(Description 1)", "steve"},
	{2, "Show this app to Gabrielle!", "(Description 2)", "steve"},
	{3, "Tell AJ!", "(Description 3)", "steve"},
	{4, "Show this off to Jim!", "(Description 4)", "steve"},
	{5, "Shout out to Sam!", "(Description 5)", "steve"},
}

var taskLock sync.RWMutex

func NewTaskGetter() *TaskGetter {
	taskLock.Lock()
	defer taskLock.Unlock()

	tasks := make([]*Task, len(allTasks))
	copy(tasks, allTasks)
	return &TaskGetter{tasks: allTasks}
}

func (tg *TaskGetter) At(i int) *Task {
	return tg.tasks[i]
}

func (tg *TaskGetter) Length() int {
	return len(tg.tasks)
}
