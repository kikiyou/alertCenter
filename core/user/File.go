package user

import (
	"github.com/kikiyou/alertCenter/models"

	"github.com/astaxie/beego"
)

// var err error

// func init() {
// 	// FileServer = beego.AppConfig.String("LADPServer")
// 	// ldapPort, err = beego.AppConfig.Int("LDAPPort")
// 	// ldapDN = beego.AppConfig.String("LDAPDN")
// 	// ldapPass = beego.AppConfig.String("LDAPPass")
// }

type FileServer struct {
}

func (e *FileServer) SearchTeams() (teams []*models.Team, err error) {
	if err != nil {
		beego.Error(err)
		return nil, err
	}

	team := &models.Team{
		ID:   "1000",
		Name: "烽视威",
	}
	teams = append(teams, team)
	// }
	return teams, nil
}

func (e *FileServer) SearchUsers() (users []*models.User, err error) {
	if err != nil {
		beego.Error(err)
		return nil, err
	}

	user := &models.User{
		ID:     "1000",
		Name:   "yh",
		TeamID: "1000",
		Phone:  "15994806909",
		Mail:   "monkey@fonsview.com",
	}
	users = append(users, user)
	return users, nil
}

func (e *FileServer) GetUserByTeam(id string) ([]*models.User, error) {

	var users []*models.User

	user := &models.User{
		ID:     "1000",
		Name:   "yh",
		TeamID: "1000",
		Phone:  "15994806909",
		Mail:   "monkey@fonsview.com",
	}
	users = append(users, user)

	return users, nil
}
