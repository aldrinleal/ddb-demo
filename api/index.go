package handler

import (
	"github.com/aldrinleal/ddb-demo/analytics"
	"github.com/gin-gonic/gin"
	"net/http"
)

var analyticsHandler *gin.Engine

func Handler(w http.ResponseWriter, r *http.Request) {
	if analyticsHandler == nil {
		newAnalyticsHandler, err := analytics.NewAnalyticsHandler("/api")

		if nil != err {
			w.WriteHeader(500)

			panic(err)
		}

		analyticsHandler = newAnalyticsHandler.GetEngine()
	}

	analyticsHandler.ServeHTTP(w, r)
}