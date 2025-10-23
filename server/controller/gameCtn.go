package controller

import (
	"errors"
	"net/http"
	"treblle/app"
	"treblle/lobbylgc"
	"treblle/util/ws"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const _MAX_LOBBY_CNT = 4

// GameCnt provides functions for WebSocket communication.
type GameCnt struct {
	logger  *zap.SugaredLogger
	lobbies map[string]*lobbylgc.Lobby // lobbies is a map of lobbies with id discordSdkId
}

// NewGameCnt creates a new controller for the game.
func NewGameCnt() app.Controller {
	var controller *GameCnt

	app.Invoke(func(logger *zap.SugaredLogger) {
		controller = &GameCnt{
			logger:  logger,
			lobbies: make(map[string]*lobbylgc.Lobby, _MAX_LOBBY_CNT),
		}
	})

	return controller
}

// RegisterEndpoints registers the WebSocket endpoint.
func (cnt *GameCnt) RegisterEndpoints(router *gin.RouterGroup) {
	router.GET("/ws/lobby/:userId/:clientId", cnt.serveLobbyWs)
}

// serveLobbyWs godoc
//
//	@Summary		web socket for lobby
//	@Description	Web socket for lobby providing data for players
//	@Tags			lobby
//	@Produce		json
//	@Failure		500
//	@Router			/ws/lobby/{userId}/{clientId} [get]
func (cnt *GameCnt) serveLobbyWs(c *gin.Context) {

	// TODO: implement checking for lobby size

	clientId := c.Param("clientId")
	if clientId == "" {
		cnt.logger.Errorf("clientId parameter is empty")
		c.AbortWithError(http.StatusBadRequest, errors.New("id parameter is required"))
		return
	}

	userId := c.Param("userId")
	if userId == "" {
		cnt.logger.Errorf("userId parameter is empty")
		c.AbortWithError(http.StatusBadRequest, errors.New("id parameter is required"))
		return
	}

	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		cnt.logger.Errorf("Failed to upgrade connection: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// crate new lobby
	if cnt.lobbies[clientId] == nil {
		cnt.lobbies[clientId] = lobbylgc.NewLobby(clientId)
	}

	// if lobby exists add new connection
	ws.NewClient(&cnt.lobbies[clientId].Hub, conn, userId, func() {
		zap.S().Debugf("Empty lobby with id %s is removed", clientId)
		delete(cnt.lobbies, clientId)
		zap.S().Infof("Lobby count is %d", len(cnt.lobbies))
	})
	zap.S().Infof("Lobby count is %d", len(cnt.lobbies))
}
