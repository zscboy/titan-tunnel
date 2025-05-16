package downloader

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStructToMap(t *testing.T) {
	opts := TaskOptions{
		Id:   uuid.NewString(),
		URL:  "http://192.168.0.132:5001/NiuLinkOS-v1.1.7-2411141913.iso",
		MD5:  "a5b380ecd94622de13b5e8261e5afc15",
		Path: "D:/codes/titan-vm/vmc/downloader/test-download-data/NiuLinkOS-v1.1.7-2411141913.iso",
	}
	task := NewTask(&opts)
	task.Start()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			if task.totalSize != 0 {
				t.Logf("download %d/%d", task.offset, task.totalSize)

				if task.totalSize == task.offset {
					cancel()
					break
				}
			}

			if task.err != nil {
				cancel()
				break
			}

			t.Logf("totalsize:%d, download %d", task.totalSize, task.offset)
			time.Sleep(1 * time.Second)

		}
	}()

	<-ctx.Done()

	if task.err != nil {
		t.Logf("download error:%s", task.err.Error())
		return
	}

	if task.offset == task.totalSize {
		t.Logf("download complete")
	}
}
