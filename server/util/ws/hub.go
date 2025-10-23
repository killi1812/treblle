package ws

import (
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NOTE: This is starting to be a bit confuzing
// Try to contain a hub and clients in the game state
// benefit of having each player id in the game state under the playerId key
// no mo wierd interfaces and dependencies

type MsgHandler interface {
	HandleMsg([]byte)
	// Update should update the chan with current status
	Update(*Client)
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	hubId      uuid.UUID
	register   chan *Client       // register used for creating new clients
	unregister chan *Client       // unregister unregisters the clinet
	Clients    map[string]*Client //  clients is a map of registered clients unser userId keys
	Broadcast  chan []byte        // broadcast is a message for all clients
	Mutex      sync.Mutex         // mutex for locking game logic
	Handler    MsgHandler         // handler handles incoming messages
}

func (hub *Hub) isEmpty() bool {
	return len(hub.Clients) == 0

}

// NewHub creates and returns a new Hub.
func NewHub() Hub {
	hubId := uuid.New()
	zap.S().Infof("Creating new hub with id %s", hubId)
	return Hub{
		hubId:      hubId,
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop.
func (hub *Hub) Run() {
	zap.S().Infof("Running event loop for hub: %s", hub.hubId)
	for {
		select {
		case client := <-hub.register:
			zap.S().Infof("Registering new client to hub: %s", hub.hubId)

			hub.Clients[client.UserId] = client
			// Send the current lobby state to the newly connected client
			zap.S().Debugf("Sending update to new client: %s", client.UserId)
			hub.Handler.Update(client)

		case client := <-hub.unregister:
			zap.S().Infof("Unregistering client %s from hub: %s", client.UserId, hub.hubId)
			if _, ok := hub.Clients[client.UserId]; ok {
				delete(hub.Clients, client.UserId)
				close(client.Send)
			}

		case message := <-hub.Broadcast:
			zap.S().Debugf("Sending messages to clients (%d) from hub: %s", len(hub.Clients), hub.hubId)
			for key, client := range hub.Clients {
				zap.S().Debugf("Sending message to client: %s", client.UserId)
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(hub.Clients, key)
				}
			}
		}
	}
}
