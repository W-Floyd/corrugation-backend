package tag

import (
	"github.com/W-Floyd/corrugation/backend/permission/user"
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Title string
	Owner *user.User
}
