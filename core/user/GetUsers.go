package user

import "fmt"

func GetUserBySource(source string) (UserInterface, error) {
	switch source {
	case "ldap":
		return &LDAPServer{}, nil
	case "gitlab":
		return &GitlabServer{}, nil
	case "file":
		return &FileServer{}, nil
	default:
		return nil, fmt.Errorf(fmt.Sprintf("Can not get the user server by source %s\n", source))
	}
}
