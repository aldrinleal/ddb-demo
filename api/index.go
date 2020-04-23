package handler

import (
	"github.com/aldrinleal/ddb-demo/analytics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var analyticsHandler *gin.Engine

func RootHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("req: %s", r.URL.String())

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
