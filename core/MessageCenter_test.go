package core

import (
	"github.com/kikiyou/alertCenter/core/db"
	"github.com/kikiyou/alertCenter/core/service"
	"fmt"
	"testing"
)

func Test_CheckRule(t *testing.T) {
	SESSION := db.GetMongoSession()
	service := &service.AlertService{
		Session: SESSION,
	}
	alerts := service.FindByUser("root", 1, 1)
	us, ok := CheckRules(alerts[0])
	fmt.Println("ok:", ok)
	fmt.Println("us:", us)
}
