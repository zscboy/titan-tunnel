package downloader

import (
	"fmt"
	"titan-vm/pb"
)

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
			return fmt.Errorf("similar task alrady exist, id %s", t.id)
		}

		if task.url == t.url {
			return fmt.Errorf("similar task alrady exist, rul %s", t.url)
		}

		if task.md5 == t.md5 {
			return fmt.Errorf("similar task alrady exist, md5 %s", t.md5)
		}

		if task.path == t.path {
			return fmt.Errorf("similar task alrady exist, path %s", t.md5)
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

func (tm *Manager) TasksToProto() []*pb.DownloadTask {
	pbDownloadTasks := make([]*pb.DownloadTask, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		pbDownloadTask := pb.DownloadTask{
			Id:           task.id,
			Url:          task.url,
			Md5:          task.md5,
			Path:         task.path,
			TotalSize:    task.totalSize,
			DownloadSize: task.downloadSize,
			Running:      task.running,
			Success:      task.success,
		}

		if task.err != nil {
			pbDownloadTask.ErrMsg = task.err.Error()
		}
		pbDownloadTasks = append(pbDownloadTasks, &pbDownloadTask)
	}
	return pbDownloadTasks
}

func (tm *Manager) TaskToProto(task *Task) *pb.DownloadTask {
	pbDownloadTask := &pb.DownloadTask{
		Id:           task.id,
		Url:          task.url,
		Md5:          task.md5,
		Path:         task.path,
		TotalSize:    task.totalSize,
		DownloadSize: task.downloadSize,
		Running:      task.running,
		Success:      task.success,
	}

	return pbDownloadTask
}
