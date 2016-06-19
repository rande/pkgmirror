package pkgmirror

import (
	"fmt"
	"net/http"
)

// code adapted from https://gist.github.com/ismasan/3fb75381cd2deb6bfa9c
type SseBroker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool

	onConnect func()
}

func (broker *SseBroker) Handler(rw http.ResponseWriter, req *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()

	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	if broker.onConnect != nil {
		go broker.onConnect()
	}

	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}
}

func (broker *SseBroker) OnConnect(fn func()) {
	broker.onConnect = fn
}

func (broker *SseBroker) Listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
		case event := <-broker.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan := range broker.clients {
				clientMessageChan <- event
			}
		}
	}
}

func NewSseBroker() *SseBroker {
	// Instantiate a broker
	return &SseBroker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}
}
