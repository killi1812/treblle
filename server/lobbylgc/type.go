package lobbylgc

import (
	"treblle/model"
	"treblle/util/ws"
)

// LobbyMessage is the structure for incoming messages from clients.
type LobbyMessage struct {
	Action lobbyAction `json:"action"`
	Slot   int         `json:"slot,omitempty"`
	User   model.User  `json:"user"`
}

type lobbyAction string

const _ACTION_JOIN_PLAYER lobbyAction = "join_player"
const _ACTION_JOIN_SPECTATOR lobbyAction = "join_spectator"
const _ACTION_LEAVE lobbyAction = "leave"

type LobbyState struct {
	Players    []*model.User `json:"players"`
	Spectators []*model.User `json:"spectators"`
}

type Lobby struct {
	// LobbyId is a representaion of isolated lobby instance
	//
	// value of discordSdk.InstanceId
	LobbyId string
	Hub     ws.Hub
	State   LobbyState
}

// NewLobby creates a new lobby and runs its event loop
func NewLobby(lobbyId string) *Lobby {
	var lobby = Lobby{
		LobbyId: lobbyId,
		Hub:     ws.NewHub(),
		State:   NewLobbyState(),
	}
	lobby.Hub.Handler = &lobby
	go lobby.Hub.Run()
	return &lobby
}
