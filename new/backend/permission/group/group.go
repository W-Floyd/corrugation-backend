package group

import (
	"github.com/W-Floyd/corrugation/backend/permission/user"
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	Title    string
	Owners   []*user.User
	Users    []*user.User
	Readonly bool
}
