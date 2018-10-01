package main

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Message struct {
	mtype string
	msg   []byte
}

/* Reads and writes messages from client */
type Client struct {
	conn 	*websocket.Conn
	out  	chan Message
	player 	int
	mu 		sync.Mutex
}

func (c *Client) WritePump() {
	waitTime := 30 * time.Second
	ticker := time.NewTicker(waitTime)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadLoop() {
	defer close(c.out)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// log.Println("read:", err)
			break
		}
		msg := Message{"ex", message}
		c.out <- msg
	}
}

func (c *Client) KeepAlive(r *Room) {
	defer close(c.out)
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			// log.Println("read:", err)
			break
		} else {

		}
	}
}

func (c *Client) WriteMessage(msg []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
	}
}

func (c *Client) WriteJSON(i interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.WriteJSON(i)
	if err != nil {
		log.Println("write:", err)
	}
}

/* Constructor */
func NewClient(conn *websocket.Conn) *Client {
	client := new(Client)
	client.conn = conn
	client.out = make(chan Message)
	return client
}
