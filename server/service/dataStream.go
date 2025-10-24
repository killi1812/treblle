package service

import (
	"encoding/json"
	"sync"
	"time"
	"treblle/dto"
	"treblle/util/ws"

	"go.uber.org/zap"
)

// ChartMessage is the structure for incoming messages from clients.
type ChartMessage struct {
	Action           chartDataAction `json:"action"`
	TimeIntervalInMs int             `json:"time_interval_in_ms"`
}

type chartDataAction string

const _ACTION_REFRESH chartDataAction = "refresh"
const _ACTION_UPDATE_INTERVAL chartDataAction = "update_interval"

type Lobby struct {
	Hub                ws.Hub
	LastUpdate         *time.Time
	requestCrudService IRequestCrudService
	task               *PeriodicTask
}

// NewLobby creates a new lobby and runs its event loop
func NewLobby() *Lobby {
	now := time.Now()
	var lobby = Lobby{
		Hub:                ws.NewHub(),
		requestCrudService: NewRequestCrudService(),
		LastUpdate:         &now,
	}

	lobby.Hub.Handler = &lobby
	actionFunc := func() {
		var state dto.RequestStatistics
		now := time.Now()
		data, err := lobby.requestCrudService.GetStatistics(lobby.LastUpdate, &now)
		if err != nil {
			zap.S().Errorf("Error retriving statistics, error = %v", err)
			return
		}
		state.FromModel(data)
		lobby.LastUpdate = &now

		updatedState, err := json.Marshal(state)
		if err != nil {
			zap.S().Errorf("Faled to marshal state, state %+v ,err = %w", state, err)
		}
		zap.S().Debugf("Lobby state: %+v", state)
		lobby.Hub.Broadcast <- updatedState
	}
	task := NewPeriodicTask(actionFunc, 10*time.Second)
	go lobby.Hub.Run(task.Stop)
	task.Start()
	lobby.task = task
	return &lobby
}

// HandleMessage implements ws.MessageProcessor.
func (lobby *Lobby) HandleMsg(data []byte) {
	if data == nil {
		return
	}

	var msg ChartMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		zap.S().Errorf("error unmarshalling message: %v", err)
		return
	}
	zap.S().Debugf("Message: %+v", msg)

	var state dto.RequestStatistics
	switch msg.Action {
	case _ACTION_REFRESH:
		now := time.Now()
		data, err := lobby.requestCrudService.GetStatistics(lobby.LastUpdate, &now)
		if err != nil {
			zap.S().Errorf("Error retriving statistics, error = %v", err)
			return
		}
		state.FromModel(data)
		lobby.LastUpdate = &now
	case _ACTION_UPDATE_INTERVAL:
		lobby.task.UpdateInterval(time.Duration(msg.TimeIntervalInMs))
	default:
		zap.S().Errorf("Unknown action %s", msg.Action)
	}

	// Broadcast the updated state
	updatedState, err := json.Marshal(state)
	if err != nil {
		zap.S().Errorf("Faled to marshal state, err = %w", err)
	}

	zap.S().Debugf("Lobby state: %+v", state)
	lobby.Hub.Broadcast <- updatedState
}

// Update implements ws.MessageProcessor.
func (lobby *Lobby) Update(client *ws.Client) {
	// Broadcast the updated state
	var state dto.RequestStatistics
	updatedState, err := json.Marshal(state)
	if err != nil {
		zap.S().Errorf("Faled to marshal state, err = %w", err)
	}

	zap.S().Debugf("Lobby state: %+v", state)
	client.Send <- updatedState
}

type PeriodicTask struct {
	action          func()        // The function to execute periodically.
	initialInterval time.Duration // The starting interval.

	intervalChan chan time.Duration // Channel to receive interval updates.
	stopChan     chan struct{}      // Channel to signal stopping.
	wg           sync.WaitGroup     // WaitGroup to track the running goroutine.

}

// NewPeriodicTask creates a new task runner.
func NewPeriodicTask(action func(), initialInterval time.Duration) *PeriodicTask {
	// Basic validation
	if initialInterval <= 0 {
		initialInterval = 1 * time.Second // Default interval if invalid
		zap.S().Warnf("Warning: Initial interval was invalid, defaulting to %s", initialInterval)
	}
	return &PeriodicTask{
		action:          action,
		initialInterval: initialInterval,
		intervalChan:    make(chan time.Duration), // Channel for interval updates
		stopChan:        make(chan struct{}),      // Channel for stopping
	}
}

// Start begins the periodic execution in a separate goroutine.
func (pt *PeriodicTask) Start() {
	zap.S().Infof("Starting periodic task with interval %s", pt.initialInterval)
	pt.wg.Add(1) // Increment WaitGroup counter

	go func() {
		defer pt.wg.Done() // Decrement counter when goroutine exits

		currentInterval := pt.initialInterval
		ticker := time.NewTicker(currentInterval)
		defer ticker.Stop() // Ensure ticker resources are released

		for {
			select {
			case <-ticker.C:
				// Interval elapsed, run the action
				zap.S().Debugf("Ticker fired (interval: %s). Running action.", currentInterval)
				pt.action()

			case newInterval := <-pt.intervalChan:
				// Received an interval update
				if newInterval > 0 && newInterval != currentInterval {
					ticker.Stop() // Stop the old ticker
					currentInterval = newInterval
					ticker = time.NewTicker(currentInterval) // Start a new ticker
					zap.S().Debugf("Updated task interval to %s\n", currentInterval)
				} else if newInterval <= 0 {
					zap.S().Warnf("Warning: Received invalid interval update (%s), ignoring.\n", newInterval)
				}

			case <-pt.stopChan:
				// Received stop signal
				zap.S().Infoln("Stopping periodic task.")
				return // Exit the goroutine
			}
		}
	}()
}

// UpdateInterval sends a new interval duration to the running task.
// The task will apply the new interval on its next cycle after receiving.
func (pt *PeriodicTask) UpdateInterval(newInterval time.Duration) {
	select {
	case pt.intervalChan <- newInterval:
		zap.S().Debugf("Sent interval update request: %s\n", newInterval)
	default:
		// Optional: Handle case where update channel might be blocked
		zap.S().Warnln("Warning: Interval update channel busy, update might be skipped.")
	}
}

// Stop signals the periodic task to stop execution and waits for it to finish.
func (pt *PeriodicTask) Stop() {
	zap.S().Debugln("Requesting task stop...")
	close(pt.stopChan) // Close the channel to signal stop
	pt.wg.Wait()       // Wait for the goroutine to finish
	zap.S().Debugln("Task stopped.")
}
