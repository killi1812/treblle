package lobbylgc

import (
	"slices"
	"treblle/model"

	"go.uber.org/zap"
)

func (state *LobbyState) handleJoinPlayer(msg LobbyMessage) {
	// TODO: find a better way of handling findUserFunc
	findUserFunc := func(s *model.User) bool {
		return s != nil && s.ID == msg.User.ID
	}

	zap.S().Debugf("Players (id: %s, username: %s) is trying to join", msg.User.ID, msg.User.Username)

	if msg.Slot < 0 || msg.Slot > 3 {
		zap.S().Errorf("Trying to join bad slot: %d", msg.Slot)
		return
	}

	if user := state.Players[msg.Slot]; user != nil {
		zap.S().Errorf("Trying to join occupied slot: %d, by player id: %s, username: %s", msg.Slot, user.ID, user.Username)
		return
	}

	skipSpecCheck := false
	// Remove player in another slot
	{
		index := slices.IndexFunc(state.Players, findUserFunc)
		if index != -1 {
			state.Players[index] = nil
			zap.S().Debugf("Player (id: %s, username: %s) changed position from %d to %d", msg.User.ID, msg.User.Username, index, msg.Slot)
			skipSpecCheck = true
		}
	}
	// Remove player from spectators
	if !skipSpecCheck {
		sizeOld := len(state.Spectators)
		state.Spectators = slices.DeleteFunc(state.Spectators, findUserFunc)
		size := len(state.Spectators)
		if sizeOld != size {
			zap.S().Debugf("Player (id: %s, username: %s) changed position from spectators to %d", msg.User.ID, msg.User.Username, msg.Slot)
		}
	}

	state.Players[msg.Slot] = &msg.User
	zap.S().Debugf("Player (id :%s, username: %s) joined successfully", msg.User.ID, msg.User.Username)
}

func (state *LobbyState) handleJoinSpectator(msg LobbyMessage) {
	// TODO: find a better way of handling findUserFunc
	findUserFunc := func(s *model.User) bool {
		return s != nil && s.ID == msg.User.ID
	}

	skipSpecCheck := false
	// Remove player in another slot
	{
		index := slices.IndexFunc(state.Players, findUserFunc)
		if index != -1 {
			state.Players[index] = nil
			zap.S().Debugf("Player (id: %s, username: %s) changed position from %d to spectators", msg.User.ID, msg.User.Username, index)
		}
	}
	// Remove player from spectators
	if !skipSpecCheck {
		index := slices.IndexFunc(state.Spectators, findUserFunc)
		if index != -1 {
			zap.S().Debugf("Player (id: %s, username: %s) already in spectators", msg.User.ID, msg.User.Username, index)
			return
		}
	}

	zap.S().Debugf("Players (id :%s, username: %s) joined spectators", msg.User.ID, msg.User.Username)
	state.Spectators = append(state.Spectators, &msg.User)
}

func (state *LobbyState) handleLeave(msg LobbyMessage) {
	// TODO: find a better way of handling findUserFunc
	findUserFunc := func(s *model.User) bool {
		return s != nil && s.ID == msg.User.ID
	}

	// Remove player in another slot
	{
		index := slices.IndexFunc(state.Players, findUserFunc)
		if index != -1 {
			state.Players[index] = nil
			zap.S().Debugf("Players (id :%s, username: %s) left the game", msg.User.ID, msg.User.Username)
			return
		}
	}
	// Remove player from spectators
	{
		sizeOld := len(state.Spectators)
		state.Spectators = slices.DeleteFunc(state.Spectators, findUserFunc)
		size := len(state.Spectators)
		if sizeOld != size {
			zap.S().Debugf("Players (id :%s, username: %s) left spectators", msg.User.ID, msg.User.Username)
		}
	}

}
