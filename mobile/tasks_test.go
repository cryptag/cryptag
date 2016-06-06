// Steven Phillips / elimisteve
// 2016.05.31

package mobile

import "testing"

func TestTaskGetter(t *testing.T) {
	tasks := NewTaskGetter()
	for i := 0; i < tasks.Length(); i++ {
		task := tasks.At(i)
		if task == nil {
			t.Errorf("Error: *Task %d is nil!", i)
			continue
		}
		t.Logf("Task %d: %s\n", i, task)
	}
}
