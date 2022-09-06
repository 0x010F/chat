package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Hub  *Hub
	Send chan *Message
	Room string
}

type ClientMessage struct {
	Type    string  `json:"type"`
	Payload Message `json:"payload"`
}

const MESSAGE string = "message"

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Disconnect <- c
		c.Conn.Close()
	}()
	for {
		var message ClientMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			log.Println("Error reading message")
			log.Println(err)
			return
		}
		switch message.Type {
		case MESSAGE:
			var m = Message{
				Body: message.Payload.Body,
				Room: message.Payload.Room,
			}
			c.Hub.Broadcast <- &m
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		// c.Hub.Disconnect <- *c
		// close(c.Send)
		c.Conn.Close()
	}()
	for {
		message, ok := <-c.Send
		if !ok {
			log.Println("Channel is closed")
			return
		}
		var m = ClientMessage{
			Type: "message",
			Payload: Message{
				Body: message.Body,
				Room: message.Room,
			},
		}
		c.Conn.WriteJSON(m)
	}
}
