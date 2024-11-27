package main

import (
	"fmt"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
)

func main() {
	server := socketio.NewServer(nil)

	// Redis adapter for socket.io
	_, err := server.Adapter(&socketio.RedisAdapterOptions{
		Addr:   "127.0.0.1:6379",
		Host:   "127.0.0.1",
		Port:   "6379",
		Prefix: "socket.io",
		DB:     1,
	})
	if err != nil {
		log.Fatal("error:", err)
	}

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		s.Join("chat_room") // Join room
		fmt.Println("connected:", s.ID(), s.Rooms())
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/", "broadcast-msg", func(s socketio.Conn, msg string) {
		fmt.Println("broadcast-msg:", msg)

		server.BroadcastToRoom("/", "chat_room", "broadcast_event", msg) // Broadcast to room: chat_room, event: broadcast_event
	})

	// Handle chat event from client
	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		fmt.Println("Message received:", msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		if e.Error() == "websocket: close 1001 (going away)" {
			fmt.Println("client disconnected:", s.ID())
		} else {
			fmt.Println("meet error:", e)
		}
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		// Connection closed, handle cleanup if necessary
		fmt.Println("closed", reason)
		s.Leave("chat_room")
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("../")))
	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
