package repository

import "github.com/iraunit/get-link-backend/util"

func NewUsersMap() *map[string]util.User {
	usersMap := make(map[string]util.User)
	return &usersMap
}
