package main

import (
	"log"
	"sync"
)

type Hub struct {
	Clients    map[string]map[*Client]bool
	Mutex      sync.RWMutex
	Connect    chan *Client
	Disconnect chan *Client
	Broadcast  chan *Message
}

func (h *Hub) add(client *Client) {
	log.Println("Adding new client")
	h.Clients[client.Room][client] = true
	log.Println("Joined room:", client.Room)
}

func (h *Hub) delete(client *Client) {
	log.Println("Deleting client: ", client)

	// connected clients
	log.Println("Currently Connected clients")
	for k, v := range h.Clients {
		log.Println(k, v)
	}

	// remove from clients
	if _, ok := h.Clients[client.Room]; ok {
		delete(h.Clients[client.Room], client)
		close(client.Send)
		log.Println("Left room:", client.Room)
	}

	log.Println("Currently Connected clients after deleting")
	for k, v := range h.Clients {
		log.Println(k, v)
	}
}

func (h *Hub) broadcast(message *Message) {
	for k := range h.Clients[message.Room] {
		log.Println("Sending message to client: ", k)
		k.Send <- message
	}
}

func (h *Hub) Run() {
	log.Println("Running")
	for {
		select {
		case client := <-h.Connect:
			log.Println("Connect")
			h.add(client)
			client.Send <- &Message{
				Body: "Connected to room #" + client.Room,
				Room: client.Room,
			}
		case message := <-h.Broadcast:
			log.Println(message)
			h.broadcast(message)
		case client := <-h.Disconnect:
			log.Println("To be deleted: ", client)
			h.delete(client)
		}
	}
}
