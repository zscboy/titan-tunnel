package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	// upgrader = websocket.Upgrader{} // use default options
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)
