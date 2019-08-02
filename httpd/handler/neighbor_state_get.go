package handler

import (
	"bgpush/bgp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NeighborStateGet(mh *bgp.MessageHandler) gin.HandlerFunc {
	fn := func(context *gin.Context) {
		neighborAddress := context.Param("neighborAddress")

		if state, ok := mh.States[neighborAddress]; ok {
			context.JSON(http.StatusOK, state)
		} else {
			context.String(http.StatusBadRequest, "State by specified neighborAddress not found")
		}
	}

	return gin.HandlerFunc(fn)
}
