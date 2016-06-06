package mobile

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elimisteve/cryptag/backend"
	"github.com/elimisteve/cryptag/mobile/cryptask"
	"github.com/elimisteve/cryptag/rowutil"
	"github.com/elimisteve/cryptag/types"
)

type Task cryptask.Task

func (t *Task) String() string {
	if t == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%#v", t)
}

var (
	bk        *backend.WebserverBackend
	taskCh    = make(chan types.Rows)
	tagPairCh = make(chan types.TagPairs)

	plainTagsGet = []string{
		"app:cryptask",
		"type:task",
	}
)

func init() {
	key := []byte{}
	ws, err := backend.NewWebserverBackend(
		key,
		"sandstorm-cryptask",
		"BASE_URL",
		"API_KEY",
	)
	if err != nil {
		log.Fatalf("Error from NewWebserverBackend: %v\n", err)
	}
	bk = ws

	go func() {
		pairs, _ := bk.AllTagPairs()
		log.Printf("AllTagPairs initially fetched %d pairs\n", len(pairs))
		tagPairCh <- pairs

		tick := time.Tick(30 * time.Second)
		for {
			select {
			case tagPairCh <- pairs:
				// Pass to polling goroutine, below
			case <-tick:
				// Fetch latest TagPairs
				latestPairs, err := bk.AllTagPairs()
				if err != nil {
					log.Printf("Error from AllTagPairs: %v\n", err)
					continue
				}
				log.Printf("Just fetched %d TagPairs\n", len(latestPairs))
				pairs = latestPairs
			}
		}
	}()

	go func() {
		taskRows, _ := backend.RowsFromPlainTags(bk, <-tagPairCh, plainTagsGet)
		log.Printf("RowsFromPlainTags initially fetched %d task rows\n",
			len(taskRows))
		taskCh <- taskRows

		tick := time.Tick(30 * time.Second)
		for {
			select {
			case taskCh <- taskRows:
				// Passed most-recently-fetched tasks to
				// NewTaskGetter()
			case <-tick:
				pairs := <-tagPairCh
				rows, err := backend.RowsFromPlainTags(bk, pairs, plainTagsGet)
				if err != nil {
					log.Printf("Error from RowsFromPlainTags: %v\n", err)
					continue
				}
				log.Printf("Just fetched %d *Rows\n", len(rows))
				taskRows = rows
			}
		}
	}()
}

type TaskGetter struct {
	tasks []*Task
}

func NewTaskGetter() *TaskGetter {
	rows := <-taskCh
	rows.Sort(rowutil.ByTagPrefix("created:", true))
	allTasks := tasksFromRows(rows)

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

func tasksFromRows(rows []*types.Row) []*Task {
	tasks := make([]*Task, 0, len(rows))
	for i, row := range rows {
		var t Task
		err := json.Unmarshal(row.Decrypted(), &t)
		if err != nil {
			log.Printf("Error unmarshaling task row: %v\n", err)
			continue
		}

		// Row tagged with "assignee:alice" and "assignee:bob" -> "alice, bob"
		names := rowutil.TagsWithPrefixStripped(row, "assignee:")
		assignees := strings.Join(names, ", ")

		t.Assignee = assignees

		// t.ID = rowutil.TagWithPrefixStripped(row, "id:")
		t.ID = fmt.Sprintf("%d", i+1)

		tasks = append(tasks, &t)
	}
	log.Printf("tasksFromRows: returning %d *Task objects\n", len(tasks))
	return tasks
}
