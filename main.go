package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// http to websocket connection upgrader.
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// websocket handler.
func ws(w http.ResponseWriter, r *http.Request, h *Hub) {
	t := strings.Split(r.URL.Path, "/")
	room := t[len(t)-1]
	log.Println(room)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error in upgrading: ", err)
		return
	}
	// defer conn.Close()

	if _, ok := h.Clients[room]; !ok {
		h.Clients[room] = make(map[*Client]bool)
	}

	client := &Client{
		Conn: conn,
		Hub:  h,
		Send: make(chan *Message),
		Room: room,
	}

	client.Hub.Connect <- client

	go client.ReadPump()
	go client.WritePump()
}

func main() {
	// hub
	hub := &Hub{
		Clients:    make(map[string]map[*Client]bool),
		Connect:    make(chan *Client),
		Disconnect: make(chan *Client),
		Broadcast:  make(chan *Message),
		Mutex:      sync.RWMutex{},
	}

	// run hub
	go hub.Run()

	// parse template
	t, err := template.ParseFiles("./index.html")
	if err != nil {
		log.Println("Error while parsing template")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/general", http.StatusTemporaryRedirect)
		}
		if err := t.Execute(w, map[string]interface{}{
			"title": "Chat app",
		}); err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		ws(w, r, hub)
	})

	log.Fatalln(http.ListenAndServe(":8000", nil))
}
