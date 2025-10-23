package app

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RequestLogger interface {
	LogRequest(req *http.Request) error
}

func Proxy(router *gin.RouterGroup) {
	const targetAPI = "https://www.thecocktaildb.com"
	target, err := url.Parse(targetAPI)
	if err != nil {
		zap.S().Fatalf("Failed to parse target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		const prefixToRemove = "/proxy"
		if after, ok := strings.CutPrefix(req.URL.Path, prefixToRemove); ok {
			req.URL.Path = after
		}
		originalDirector(req)
		req.Host = target.Host
	}
	var reqLogger RequestLogger
	Invoke(func(logger RequestLogger) {
		reqLogger = logger
	})

	proxyHandler := func(c *gin.Context) {
		if err := reqLogger.LogRequest(c.Request); err != nil {
			zap.S().Errorf("Failed to log request, error %v", err)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	router.Any("/*proxyPath", proxyHandler)
}
