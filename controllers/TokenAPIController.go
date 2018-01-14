package controllers

import (
	"github.com/kikiyou/alertCenter/core/db"
	"github.com/kikiyou/alertCenter/core/service"
	"github.com/kikiyou/alertCenter/util"

	"github.com/astaxie/beego"
)

type TokenAPIController struct {
	APIBaseController
}

func (e *TokenAPIController) GetAllToken() {

	user := e.Ctx.Input.Header("user")
	if user == "" {
		e.Data["json"] = util.GetErrorJson("please certification")
		e.ServeJSON()
	} else {
		service := &service.TokenService{
			Session: db.GetMongoSession(),
		}
		if service.Session != nil {
			defer service.Session.Close()
		}
		tokens := service.GetAllToken(user)
		e.Data["json"] = util.GetSuccessReJson(tokens)
		e.ServeJSON()
	}
}

func (e *TokenAPIController) DeleteToken() {
	user := e.Ctx.Input.Header("user")
	project := e.GetString(":project")
	beego.Debug("delete token,user:" + user)
	if user == "" {
		e.Data["json"] = util.GetErrorJson("please certification")
		e.ServeJSON()
	} else {
		service := &service.TokenService{
			Session: db.GetMongoSession(),
		}
		if service.Session != nil {
			defer service.Session.Close()
		}
		if ok := service.DeleteToken(project, user); ok {
			e.Data["json"] = util.GetSuccessJson("Delete project success")
		} else {
			e.Data["json"] = util.GetFailJson("Delete project faild")
		}
		e.ServeJSON()
	}
}
