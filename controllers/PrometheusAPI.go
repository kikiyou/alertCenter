package controllers

import (
	"encoding/json"
	"fmt"

	"github.com/kikiyou/alertCenter/core"
	"github.com/kikiyou/alertCenter/core/db"
	"github.com/kikiyou/alertCenter/core/service"
	"github.com/kikiyou/alertCenter/models"
	"github.com/kikiyou/alertCenter/util"

	"github.com/astaxie/beego"
)

type PrometheusAPI struct {
	beego.Controller
}

//ReceivePrometheus 单独验证prometheus
func (e *PrometheusAPI) ReceivePrometheus() {
	ip := e.Ctx.Input.IP()
	fmt.Println("send ip -> ", ip)
	configService := &service.GlobalConfigService{
		Session: db.GetMongoSession(),
	}
	if configService.Session != nil {
		defer configService.Session.Close()
	}
	if ok, _ := configService.CheckExist("TrustIP", ip); ok {
		data := e.Ctx.Input.RequestBody
		if data != nil && len(data) > 0 {
			var Alerts []*models.Alert
			err := json.Unmarshal(data, &Alerts)
			if err == nil {
				core.HandleAlerts(Alerts)
				beego.Debug(Alerts)
				e.Data["json"] = util.GetSuccessJson("receive alert success")
			} else {
				e.Data["json"] = util.GetErrorJson("receive a unknow data")
			}
		}
	} else {
		beego.Debug(ip + " is not trust ip")
		e.Data["json"] = util.GetFailJson("Have no right to access")
	}
	e.ServeJSON()

}
