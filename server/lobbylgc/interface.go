package lobbylgc

import (
	"encoding/json"
	"treblle/model"
	"treblle/util/ws"

	"go.uber.org/zap"
)

const _PLAYER_COUNT = 4

func NewLobbyState() LobbyState {
	return LobbyState{
		Players:    make([]*model.User, _PLAYER_COUNT),
		Spectators: make([]*model.User, 0),
	}
}

// HandleMessage implements ws.MessageProcessor.
func (lobby *Lobby) HandleMsg(data []byte) {
	// TODO: check when to lock the mutex before or after message decoding
	lobby.Hub.Mutex.Lock()

	if data != nil {
		var msg LobbyMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			zap.S().Errorf("error unmarshalling message: %v", err)
			return
		}
		zap.S().Debugf("Message: %+v", msg)

		switch msg.Action {
		case _ACTION_JOIN_PLAYER:
			lobby.State.handleJoinPlayer(msg)
		case _ACTION_JOIN_SPECTATOR:
			lobby.State.handleJoinSpectator(msg)
		case _ACTION_LEAVE:
			lobby.State.handleLeave(msg)
		default:
			zap.S().Errorf("Unknown lobbyAction %s", msg.Action)
		}
	}

	lobby.Hub.Mutex.Unlock()

	// Broadcast the updated state
	updatedState, err := json.Marshal(lobby.State)
	if err != nil {
		zap.S().Errorf("Faled to marshal state, err = %w", err)
	}

	zap.S().Debugf("Lobby state: %+v", lobby.State)
	lobby.Hub.Broadcast <- updatedState
}

// Update implements ws.MessageProcessor.
func (lobby *Lobby) Update(client *ws.Client) {
	// Broadcast the updated state
	updatedState, err := json.Marshal(lobby.State)
	if err != nil {
		zap.S().Errorf("Faled to marshal state, err = %w", err)
	}

	zap.S().Debugf("Lobby state: %+v", lobby.State)
	zap.S().Debugf("Updating client userId:", client.UserId)
	client.Send <- updatedState
}
