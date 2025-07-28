//go:build !android

package log

import "log"

func logInfo(tag, msg string) {
	log.Printf("[INFO] [%s] %s\n", tag, msg)
}
