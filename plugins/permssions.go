package plugins

import (
	"github.com/jinzhu/gorm"
	"github.com/go-irc/irc"
	"github.com/belak/go-seabird"
)

func init() {
	seabird.RegisterPlugin("permissions", newPermissionsPlugin)
}

type PermissionsPlugin struct {
	isupport *ISupportPlugin

	ctracker *ChannelTracker

	db *gorm.DB
}

type PermUser struct {
	gorm.Model
	Type        string       // 'mask', 'account', and 'channel' supported.
	Identifier  string       // either a *!*@* mask, an account name, or a channel name.
	Permissions []Permission `gorm:"many2many:user_permissions;"`
	Roles       []Role       `gorm:"many2many:user_roles;"`
}

type Role struct {
	gorm.Model
	Name        string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type Permission struct {
	gorm.Model
	Name        string
	Description string
}

func newPermissionsPlugin(b *seabird.Bot, isupport *ISupportPlugin, ctracker *ChannelTracker, db *gorm.DB) *PermissionsPlugin {
	p := &PermissionsPlugin{
		isupport: isupport,
		ctracker: ctracker,
		db:       db,
	}

	return p
}

// Permitted should never error. Failure mode is to deny permission. When using to check
func (p *PermissionsPlugin) Permitted(m *irc.Message, perms []string) bool {
	user := p.ctracker.LookupUser(m.Prefix.Nick)
	if user == nil {
		// do something dramatic
	}


	return false
}



