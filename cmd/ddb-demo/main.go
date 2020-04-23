package main

import (
	"github.com/aldrinleal/ddb-demo/analytics"
	"github.com/gin-contrib/static"
)

func main() {
	handler, err := analytics.NewAnalyticsHandler("/api")

	if nil != err {
		panic(err)
	}

	engine := handler.GetEngine()

	engine.Use(static.Serve("/", static.LocalFile("public", true)))

	engine.Run(":8000")
}
