package app

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"treblle/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const _REQUEST_ID_KEY = "RequestIdKey"

type RequestLogger interface {
	LogRequest(req *http.Request) (*model.Request, error)
	LogResponse(id uint, resp *http.Response) (*model.Request, error)
}

func Proxy(router *gin.RouterGroup) {
	// TODO: move to env var
	target, err := url.Parse(ProxyUrl)
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
		req, err := reqLogger.LogRequest(c.Request)
		if err != nil {
			zap.S().Errorf("Failed to log request, error %v", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx := context.WithValue(c.Request.Context(), _REQUEST_ID_KEY, req.ID)
		c.Request = c.Request.WithContext(ctx)
		// ----------------------------------------
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		strData := resp.Request.Context().Value(_REQUEST_ID_KEY)
		if strData != nil {
			requestID := strData.(uint)
			if _, err := reqLogger.LogResponse(requestID, resp); err != nil {
				zap.S().Errorf("Failed to log response, error %v", err)
				return err
			}
		}

		return nil
	}

	router.Any("/*proxyPath", proxyHandler)
}
