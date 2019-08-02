package main

import (
	"bgpush/bgp"
	"bgpush/httpd/handler"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	help, apiListen, path := runParameters()

	if help {
		printHelp()
	}

	configuration := configuration(path)
	mh := bgp.New(configuration)

	mh.Run()

	router := gin.Default()

	pprof.Register(router)

	router.POST("/rib/update", handler.RibUpdatePost(mh))
	router.GET("/rib/out", handler.RibOutGet(mh))
	router.GET("/rib/in", handler.RibInGet(mh))
	router.GET("/rib/global", handler.RibGlobalGet(mh))
	router.GET("/rib/count", handler.RibCountGet(mh))
	router.GET("/neighbor/state/:neighborAddress", handler.NeighborStateGet(mh))

	if err := router.Run(apiListen); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
