package downloader

import "fmt"

const (
	maxTask = 5
)

type Manager struct {
	tasks []*Task
}

func NewManager() *Manager {
	return &Manager{tasks: make([]*Task, 0)}
}

func (tm *Manager) AddTask(t *Task) error {
	if len(tm.tasks) >= maxTask {
		return fmt.Errorf("task out of %d", maxTask)
	}

	for _, task := range tm.tasks {
		if task.id == t.id {
			return fmt.Errorf("id %s already exist", t.id)
		}

		if task.url == t.url {
			return fmt.Errorf("url  %s already exist", t.md5)
		}

		if task.md5 == t.md5 {
			return fmt.Errorf("md5 %s already exist", t.md5)
		}

		if task.path == t.path {
			return fmt.Errorf("path  %s already exist", t.md5)
		}
	}

	tm.tasks = append(tm.tasks, t)
	return nil
}

func (tm *Manager) DeleteTask(t *Task) error {
	for i, task := range tm.tasks {
		if task.id == t.id {
			tm.tasks = append(tm.tasks[:i], tm.tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task with id %s not found", t.id)
}

func (tm *Manager) GetTask(id string) *Task {
	for _, task := range tm.tasks {
		if task.id == id {
			return task
		}
	}
	return nil
}
