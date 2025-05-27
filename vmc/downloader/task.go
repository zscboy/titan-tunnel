package downloader

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

type Task struct {
	id   string
	url  string
	md5  string
	path string

	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	downloadSize int64 // 当前已下载大小
	totalSize    int64
	err          error
	success      bool
}

type TaskOptions struct {
	Id   string
	URL  string
	MD5  string
	Path string
}

// NewTask 初始化下载任务
func NewTask(opts *TaskOptions) *Task {
	return &Task{id: opts.Id, url: opts.URL, md5: opts.MD5, path: opts.Path}
}

// Start 启动下载
func (t *Task) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return errors.New("task already running")
	}

	// todo: check file if exist
	info, err := os.Stat(t.path)
	if err == nil {
		file, err := os.Open(t.path)
		if err != nil {
			t.err = err
			return fmt.Errorf("open file failed: %v", err)
		}
		defer file.Close()

		if err = t.verify(file); err == nil {
			t.totalSize = info.Size()
			t.downloadSize = info.Size()
			t.success = true
			return nil
		}

	}

	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.running = true

	go t.download()
	return nil
}

// Stop 停止下载
func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}
	t.running = false
}

// download 是下载的实际逻辑
func (t *Task) download() {
	defer t.setStopped()
	defer func() {
		if t.err != nil {
			logx.Errorf("download failed:%s", t.err.Error())
		}
	}()

	file, err := os.OpenFile(t.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		t.err = err
		return
	}
	defer file.Close()

	req, err := http.NewRequestWithContext(t.ctx, "GET", t.url, nil)
	if err != nil {
		t.err = err
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.err = err
		return
	}
	defer resp.Body.Close()

	if clen := resp.Header.Get("Content-Length"); clen != "" {
		if size, err := strconv.ParseInt(clen, 10, 64); err == nil {
			t.totalSize = size
		}
	}

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-t.ctx.Done():
			t.err = fmt.Errorf("ctx cancel task")
			return
		default:
			n, err := resp.Body.Read(buf)
			if n > 0 {
				if _, err := file.Write(buf[:n]); err != nil {
					t.err = err
					return
				}
				t.downloadSize += int64(n)
			}
			if err != nil {
				if err == io.EOF {
					file.Seek(0, io.SeekStart)
					if err = t.verify(file); err != nil {
						os.RemoveAll(t.path)
					} else {
						t.success = true
					}
				}
				t.err = err
				return
			}
		}
	}
}

func (t *Task) setStopped() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.running = false
	t.cancel = nil
}

func (t *Task) verify(file io.Reader) error {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	if hex.EncodeToString(hash.Sum(nil)) != t.md5 {
		return fmt.Errorf("md5 not match")
	}

	return nil
}

func (t *Task) IsRunning() bool {
	return t.running
}

func (t *Task) GetId() string {
	return t.id
}
