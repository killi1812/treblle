package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"treblle/docs"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var signalNotificationCh = make(chan os.Signal, 1)

// Start will start the web server of the app
func Start() {
	// relay selected signals to channel
	// - os.Interrupt, ctrl-c
	// - syscall.SIGTERM, program termination
	signal.Notify(signalNotificationCh, os.Interrupt, syscall.SIGTERM)

	// create scheduler
	schedulerWg := sync.WaitGroup{}
	schedulerCtx := context.Background()
	schedulerCtx, schedulerCancel := context.WithCancel(schedulerCtx)
	zap.S().Debugf("Created scheduler context")

	schedulerWg.Add(1)
	go checkInterrupt(schedulerCtx, &schedulerWg, schedulerCancel)
	zap.S().Debugf("Started CheckInterrupt")

	schedulerWg.Add(1)
	go run(schedulerCtx, &schedulerWg)
	zap.S().Debugf("Started HTTP server")

	schedulerWg.Wait()

	zap.S().Debugf("Terminated program")
}

func checkInterrupt(ctx context.Context, wg *sync.WaitGroup, schedulerCancel context.CancelFunc) {
	defer wg.Done()

	for {
		select {

		case <-ctx.Done():
			zap.S().Debugf("Terminated CheckInterrupt")
			return

		case sig := <-signalNotificationCh:
			zap.S().Debugf("Received signal on notification channel, signal = %v", sig)
			schedulerCancel()
		}
	}
}

func run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// setup gin
	if Build == BuildProd {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// setup swagger
	if Build == BuildDev {
		docs.SwaggerInfo.Title = "Mahjong API"
		docs.SwaggerInfo.Description = "This is the API for the Mahjong game."
		docs.SwaggerInfo.Version = Version
		docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", Port)
		docs.SwaggerInfo.BasePath = "/api"
		docs.SwaggerInfo.Schemes = []string{"http"}

		// BUG: 500 error when accessing /swagger/index.html/doc.json
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// setup controllers
	basePath := router.Group("/api")
	for _, c := range controllers {
		c.RegisterEndpoints(basePath)
	}
	// cleanup
	controllers = nil

	addr := fmt.Sprintf(":%d", Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Panicf("Failes to start server err = %+v", err)
		}
	}()
	zap.S().Infof("Started HTTP listen, address = http://localhost%v", srv.Addr)

	// wait for context cancellation
	<-ctx.Done()

	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeoutCancel()
	err := srv.Shutdown(timeoutCtx)
	if err != nil {
		zap.S().Errorf("Cannot shut down HTTP server, err = %v", err)
	}
	zap.S().Info("HTTP server was shut down")
}
