package server

import (
	"fmt"
	"net"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"net/http"
)

var (
	// upgrader = websocket.Upgrader{} // use default options
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	tunMgr *TunnelManager
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	var uuid = r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("need uuid!")
		return
	}

	tunMgr.acceptWebsocket(c, uuid)
}

func libvirtHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	var uuid = r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("need uuid!")
		return
	}

	var address = r.URL.Query().Get("address")
	if len(address) > 0 {
		_, _, err := net.SplitHostPort(address)
		if err != nil {
			log.Printf("can not parse address %s\n", address)
			return
		}
	}

	tunMgr.onLibvirtClient(c, uuid, address)
}

func webVncHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	var uuid = r.URL.Query().Get("uuid")
	if uuid == "" {
		log.Println("need uuid!")
		return
	}

	var address = r.URL.Query().Get("address")
	if address == "" {
		log.Println("need uuid!")
		return
	}

	tunMgr.onWebNnc(c, uuid, address)
}

// indexHandler responds to requests with our greeting.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Hello, Stupid!")
}

// CreateHTTPServer start http server
func CreateHTTPServer(listenAddr string, wsPath string) {
	tunMgr = newTunnelManager()
	go tunMgr.keepalive()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc(wsPath, wsHandler)
	//proxy/{uuid}/{address}
	http.HandleFunc("/libvirt", libvirtHandler)
	http.HandleFunc("/vnc", webVncHandler)

	log.Printf("server listen at:%s, path:%s", listenAddr, wsPath)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
