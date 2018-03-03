package plugins

import (
	"github.com/belak/go-seabird"
	"github.com/belak/go-seabird/plugins"
	"github.com/go-irc/irc"
	"github.com/jinzhu/gorm"
)

func init() {
	seabird.RegisterPlugin("permissions", newPermissionsPlugin)
}

type PermissionsPlugin struct {
	isupport *plugins.ISupportPlugin
	ctracker *plugins.ChannelTracker
	db       *gorm.DB

	UserCache []PermUser
}

type Identifier struct {
	gorm.Model
	Type       string // 'mask', 'account', and 'channel' supported.
	Pattern    string // either a *!*@* mask, an account name, or a channel name.
	PermUserID int
	PermUser   PermUser
}

type PermUser struct {
	gorm.Model
	Name             string
	PermissionGrants []PermissionGrant `gorm:"polymorphic:User;"`
	Roles            []Role            `gorm:"many2many:user_roles;"`
}

type Role struct {
	gorm.Model
	Name            string
	PermissionGrant []PermissionGrant `gorm:"polymorphic:User;"`
}

type PermissionGrant struct {
	gorm.Model
	UserID     int    `gorm:"primary_key"`
	UserType   string `gorm:"primary_key"`
	Channel    string `gorm:"primary_key"`
	Permission []Permission
}

type Permission struct {
	gorm.Model
	Domain      string `gorm:"primary_key"`
	Name        string `gorm:"primary_key"`
	Description string
}

func newPermissionsPlugin(b *seabird.Bot, cm *seabird.CommandMux, isupport *plugins.ISupportPlugin, ctracker *plugins.ChannelTracker, db *gorm.DB) *PermissionsPlugin {
	p := &PermissionsPlugin{
		isupport: isupport,
		ctracker: ctracker,
		db:       db,
	}

	cm.Event("user", p.userCallback, &seabird.HelpInfo{
		Usage:       "<action> <nick> <params>",
		Description: "Adds user to permission list. Defaults to using account for identification.",
	})

	cm.Event("user", p.roleCallback, &seabird.HelpInfo{
		Usage:       "<action> <roll> <params>",
		Description: "Adds user to permission list. Defaults to using account for identification.",
	})

	cm.Event("user", p.permCallback, &seabird.HelpInfo{
		Usage:       "<action> <permission> <params>",
		Description: "Adds user to permission list. Defaults to using account for identification.",
	})

	return p
}

func (p *PermissionsPlugin) RegisterPerm(domain string, name string, desc string) *Permission {
	perm := &Permission{}

	p.db.FirstOrCreate(&perm, Permission{ Domain: domain, Name: name })

	perm.Description = desc

	p.db.Save(&perm)

	return perm
}

// Permitted should never error. Failure mode is to deny permission. When using to check
func (p *PermissionsPlugin) Permitted(m *irc.Message, perms []string) bool {
	user := p.ctracker.LookupUser(m.Prefix.Name)
	if user == nil {
		// do something dramatic
	}

	return false
}

//
// no public functions below here
//

func (p *PermissionsPlugin) userCallback(b *seabird.Bot, m *irc.Message) {

}

func (p *PermissionsPlugin) roleCallback(b *seabird.Bot, m *irc.Message) {

}

func (p *PermissionsPlugin) permCallback(b *seabird.Bot, m*irc.Message) {

}