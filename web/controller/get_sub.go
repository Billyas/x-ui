package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"x-ui/util/common"
	"x-ui/web/service"
)

type GetSubController struct {
	subService   service.SubService
	getSubSevice service.GetSubService
}

func NewGetSubController(g *gin.RouterGroup) *GetSubController {
	a := &GetSubController{}
	a.initRouter(g)
	return a
}

func (a *GetSubController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/getsub")
	g.GET("/:token", a.getSubscribe)
}

func (a *GetSubController) getSubscribe(c *gin.Context) {
	token := c.Param("token")
	key := a.subService.GetAESKey().Url
	realId, err := common.AESDecryptECB(key, token)
	if err != nil {
		return
	}
	id, err := strconv.Atoi(realId)
	if err != nil {
		return
	}
	subscribe := a.subService.GetSubsByIdType(id).Url
	c.String(http.StatusOK, subscribe)
}
