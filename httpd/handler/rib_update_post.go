package handler

import (
	"bgpush/bgp"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RibUpdatePost(mh *bgp.MessageHandler) gin.HandlerFunc {
	fn := func(context *gin.Context) {
		var messages []bgp.UpdateMessage
		if err := context.BindJSON(&messages); err != nil {
			context.String(http.StatusBadRequest, "Can't parse body")
		} else {

			for _, msg := range messages {
				mh.In <- msg
			}

			context.String(http.StatusOK, "")
		}
	}

	return gin.HandlerFunc(fn)
}
