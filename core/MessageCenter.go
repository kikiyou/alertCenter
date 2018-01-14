package core

import (
	"time"

	"github.com/astaxie/beego"

	"github.com/kikiyou/alertCenter/core/db"
	"github.com/kikiyou/alertCenter/core/notice"
	"github.com/kikiyou/alertCenter/core/service"
	"github.com/kikiyou/alertCenter/core/user"
	"github.com/kikiyou/alertCenter/models"

	"gopkg.in/mgo.v2/bson"

	mgo "gopkg.in/mgo.v2"
)

//HandleMessage 处理alertmanager回来的数据
// func HandleMessage(message *models.AlertReceive) {
// 	session := db.GetMongoSession()
// 	if session != nil {
// 		defer session.Close()
// 	}
// 	alertService := &service.AlertService{
// 		Session: session,
// 	}
// 	ok := SaveMessage(message, session)
// 	if !ok {
// 		beego.Error("save a message fail,message receiver:" + message.Receiver)
// 	}
// 	for _, alert := range message.Alerts {
// 		old := alertService.GetAlertByLabels(&alert)
// 		if old != nil {
// 			old.AlertCount = old.AlertCount + 1
// 			old = old.Merge(&alert)
// 			if !old.EndsAt.IsZero() {
// 				old.IsHandle = 2
// 				old.HandleDate = time.Now()
// 				old.HandleMessage = "报警已自动恢复"
// 			}
// 			alertService.Update(old)
// 		} else {
// 			alert.AlertCount = 1
// 			alert.IsHandle = 0
// 			alert.Mark = alert.Fingerprint().String()
// 			alert.Receiver = user.GetReceiverByTeam(message.Receiver)
// 			now := time.Now()
// 			// Ensure StartsAt is set.
// 			if alert.StartsAt.IsZero() {
// 				alert.StartsAt = now
// 			}
// 			if !alert.EndsAt.IsZero() {
// 				alert.IsHandle = 2
// 				alert.HandleDate = time.Now()
// 				alert.HandleMessage = "报警已自动恢复"
// 			}
// 			alertService.Save(&alert)
// 		}
// 	}

// }

//HandleAlerts 处理prometheus回来的数据
func HandleAlerts(alerts []*models.Alert) {
	session := db.GetMongoSession()
	if session != nil {
		defer session.Close()
	}
	alertService := &service.AlertService{
		Session: session,
	}
	for _, alert := range alerts {
		//start := time.Now()
		old, err := alertService.GetAlertByMark(alert.Labels.Fingerprint().String())
		beego.Debug(alert)
		if err != nil {
			if err.Error() == mgo.ErrNotFound.Error() {
				SaveAlert(alertService, alert)
			} else {
				continue
			}
		} else {
			//fmt.Println("get label:", time.Now().Sub(start))
			if old != nil && old.EndsAt.IsZero() {
				old.AlertCount = old.AlertCount + 1
				alert.UpdatedAt = time.Now()
				old = old.Merge(alert)
				//old已更新时间信息
				if !old.EndsAt.IsZero() {
					old.IsHandle = 2
					old.HandleDate = time.Now()
					old.HandleMessage = "报警已自动恢复"
					SaveHistory(alertService, old)
				}
				old.UpdatedAt = time.Now()
				Notice(old)
				alertService.Update(old)
			} else if old != nil && !old.EndsAt.IsZero() {
				//此报警曾出现过并已结束
				if alert.StartsAt.After(old.EndsAt) {
					//报警开始时间在原报警之后，我们认为这是新报警
					//old更新状态信息
					old = old.Reset(alert)
					if old.IsHandle == 2 {
						SaveHistory(alertService, old)
					}
					Notice(old)
					alertService.Update(old)
				} else if alert.StartsAt.Before(old.EndsAt) && alert.EndsAt.After(old.EndsAt) {
					// 新的结束时间
					history, err := alertService.FindHistory(old)
					if history != nil && err == nil {
						old.EndsAt = alert.EndsAt
						history.EndsAt = alert.EndsAt
						alertService.Update(old)
						alertService.UpdateHistory(history)
					}
				}
			}
		}

		//fmt.Println("alert cost:", time.Now().Sub(start))
	}
}

//Notice 发送报警通知信息
func Notice(alert *models.Alert) {
	//全局通知开关关闭
	service := &service.GlobalConfigService{
		Session: db.GetMongoSession(),
	}
	if service.Session != nil {
		defer service.Session.Close()
	}
	noticeOn, err := service.GetConfig("noticeOn")
	if err != nil && err.Error() != mgo.ErrNotFound.Error() {
		beego.Debug("get noticeOn from database error." + err.Error())
		return
	}
	if noticeOn != nil && !noticeOn.Value.(bool) {
		beego.Debug("notice center closed.")
		return
	}
	if users, ok := CheckRules(alert); ok {
		alert.Receiver.UserNames = users
		mark := alert.Fingerprint().String()

		ch, err := notice.GetChannelByMark(mark)
		if err == nil && ch != nil {
			ch <- alert
		} else {
			err := notice.CreateChanByMark(alert.Fingerprint().String())
			if err != nil {
				beego.Error(err)
			}
			go notice.NoticControl(alert)
		}
	}
}

//CheckRules 检验是否为用户忽略的报警
func CheckRules(alert *models.Alert) ([]string, bool) {
	if alert != nil && alert.Receiver != nil {
		userNames := alert.Receiver.UserNames
		//没有接收对象，不需要发送
		if userNames == nil || len(userNames) < 1 {
			beego.Debug("no receiver")
			return nil, false
		}
		relation := user.Relation{}
		var users []string
		session := db.GetMongoSession()
		if session != nil {
			defer session.Close()
		}
		ruleService := &service.IgnoreRuleService{
			Session: session,
		}
		for _, userName := range userNames {
			user := relation.GetUserByName(userName)
			var ignore bool
			if user != nil {
				rules, err := ruleService.FindRuleByUser(user.Name)
				if err != nil && err.Error() != mgo.ErrNotFound.Error() {
					continue
				}
				if rules != nil && len(rules) > 0 {
					for _, rule := range rules {
						//判断是否已过期
						if rule.EndsAt.After(time.Now()) && rule.StartsAt.Before(time.Now()) {
							if alert.Labels.Contains(rule.Labels) {
								ignore = true
							}
						}
					}
				}
			} else {
				ignore = true
			}
			if !ignore {
				users = append(users, user.Name)
			}
		}
		if len(users) == 0 {
			//beego.Debug("no receiver after check rule")
			return nil, false
		}
		return users, true
	}
	beego.Debug("alert is nil")
	return nil, false
}

//SaveHistory 存快照纪录
func SaveHistory(alertService *service.AlertService, alert *models.Alert) {
	history := &models.AlertHistory{
		ID:       bson.NewObjectId(),
		Mark:     alert.Fingerprint().String(),
		AddTime:  time.Now(),
		StartsAt: alert.StartsAt,
		EndsAt:   alert.EndsAt,
		Duration: alert.EndsAt.Sub(alert.StartsAt),
		Value:    string(alert.Annotations.LabelSet["value"]),
		Message:  string(alert.Annotations.LabelSet["description"]),
	}
	alertService.Session.Insert("AlertHistory", history)
}

//SaveAlert 保存alert信息
func SaveAlert(alertService *service.AlertService, alert *models.Alert) {
	alert.AlertCount = 1
	alert.IsHandle = 0
	alert.Mark = alert.Fingerprint().String()
	alert.Receiver = user.GetReceiver(alert.Labels)
	now := time.Now()
	// Ensure StartsAt is set.
	if alert.StartsAt.IsZero() {
		alert.StartsAt = now
	}
	alert.UpdatedAt = now
	if !alert.EndsAt.IsZero() {
		alert.IsHandle = 2
		alert.HandleDate = time.Now()
		alert.HandleMessage = "报警已自动恢复"
		SaveHistory(alertService, alert)
	}
	Notice(alert)
	alertService.Save(alert)
}

//SaveMessage 储存alertmanager的消息
func SaveMessage(message *models.AlertReceive, session *db.MongoSession) bool {
	ok := session.Insert("AlertReceive", message)
	return ok
}
