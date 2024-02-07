package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"x-ui/database/model"
	"x-ui/web/service"
)

type SubController struct {
	subService   service.SubService
	getSubSevice service.GetSubService
}

func NewSubController(g *gin.RouterGroup) *SubController {
	a := &SubController{}
	a.initRouter(g)
	return a
}

func (a *SubController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/subs")

	g.POST("/list", a.getSubs)
	g.POST("/add", a.addSub)
	g.POST("/del/:id", a.delSub)
	g.POST("/update/:id", a.updateSub)
	g.GET("/getCfNode", a.getCfNode)
	g.GET("/getSubNode", a.getSubNode)
}

func (a *SubController) getSubs(c *gin.Context) {

	subs, err := a.subService.GetSubs()
	if err != nil {
		jsonMsg(c, "获取", err)
		return
	}
	jsonObj(c, subs, nil)
}

func (a *SubController) addSub(c *gin.Context) {
	sub := &model.Sub{}
	err := c.ShouldBind(sub)
	if err != nil {
		jsonMsg(c, "添加", err)
		return
	}
	err = a.subService.AddSub(sub)
	jsonMsg(c, "添加", err)
}

func (a *SubController) delSub(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "删除", err)
		return
	}
	err = a.subService.DelSub(id)
	jsonMsg(c, "删除", err)
}

func (a *SubController) updateSub(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, "修改", err)
		return
	}
	sub := &model.Sub{
		Id: id,
	}
	err = c.ShouldBind(sub)
	if err != nil {
		jsonMsg(c, "修改", err)
		return
	}
	err = a.subService.UpdateSub(sub)
	jsonMsg(c, "修改", err)
}

func (a *SubController) getCfNode(c *gin.Context) {
	node, err := a.getSubSevice.GetLatestCFNode()
	if err != nil {
		return
	}
	c.String(http.StatusOK, node)
}

func (a *SubController) getSubNode(c *gin.Context) {
	proxy, err := a.getSubSevice.GetLatestUrlSub()
	if err != nil {
		return
	}
	c.String(http.StatusOK, proxy)
}

func (a *SubController) getSubscribe(c *gin.Context) {
	subscribe := a.subService.GetSubsById(10).Url
	c.String(http.StatusOK, subscribe)
}
