package handler

import (
	"bgpush/bgp"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	api "github.com/osrg/gobgp/api"
	"net/http"
)

func RibCountGet(mh *bgp.MessageHandler) gin.HandlerFunc {
	fn := func(context *gin.Context) {
		paths := mh.GetPaths(api.TableType_GLOBAL)

		var announcements = make([]Announce, 0)

		for _, path := range paths {
			prefix := path.Prefix
			var asPath string
			var nextHop string
			var age int64

			if len(path.Paths) == 1 {
				age = path.Paths[0].Age.Seconds
			}

			for _, attr := range path.Paths[0].Pattrs {

				if attr.TypeUrl == "type.googleapis.com/gobgpapi.NextHopAttribute" {
					nhAttr := api.NextHopAttribute{}
					err := proto.Unmarshal(attr.Value, &nhAttr)
					if err == nil {
						nextHop = nhAttr.NextHop
					}
				} else if attr.TypeUrl == "type.googleapis.com/gobgpapi.AsPathAttribute" {
					apAttr := api.AsPathAttribute{}
					err := proto.Unmarshal(attr.Value, &apAttr)
					if err == nil {
						asPath = bgp.NumbersToString(apAttr.Segments[0].Numbers, " ")
					}
				}

			}

			announcements = append(announcements, Announce{
				Prefix:  prefix,
				AsPath:  asPath,
				NextHop: nextHop,
				Age:     age,
			})
		}

		context.JSON(http.StatusOK, len(announcements))
	}

	return gin.HandlerFunc(fn)
}
