package collection

import (
	"github.com/W-Floyd/corrugation/backend/entity"
	"github.com/W-Floyd/corrugation/backend/permission/group"
	"github.com/W-Floyd/corrugation/backend/permission/user"
	"gorm.io/gorm"
)

type Collection struct {
	gorm.Model
	Title    string
	Entities []*entity.Entity

	// Ownership (who can also access)
	Owners []*group.Group

	// Access
	AccessUsers  []*user.User
	AccessGroups []*group.Group
}
