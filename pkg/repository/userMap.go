package repository

import (
	"github.com/iraunit/get-link-backend/util/bean"
)

func NewUsersMap() *map[string]bean.User {
	usersMap := make(map[string]bean.User)
	return &usersMap
}
