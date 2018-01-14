package controllers

import (
	"github.com/kikiyou/alertCenter/controllers/session"
	"github.com/kikiyou/alertCenter/core/user"
	"github.com/kikiyou/alertCenter/models"

	"github.com/astaxie/beego"
)

type UserController struct {
	BaseController
}

func (e *UserController) UserHome() {
	userName := e.GetString(":userName")
	relaction := user.Relation{}
	user := relaction.GetUserByName(userName)
	sess, err := session.GetSession(e.Ctx)
	var self interface{}
	if err != nil {
		beego.Error("get session error:", err)
	} else {
		self = sess.Get(session.SESSION_USER)
	}
	if user == nil {
		e.Abort("404")
	} else {
		e.Data["userInfo"] = user
		if self != nil && user.Name == self.(*models.User).Name {
			e.Data["self"] = true
		}
		e.Data["isAdmin"] = user.IsAdmin
		e.TplName = "userHome.html"
	}
}
