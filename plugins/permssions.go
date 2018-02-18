package plugins

func init() {
	seabird.RegisterPlugin("permissions", newPermissionsPlugin)
}

type PermissionsPlugin struct {
	isupport *ISupportPlugin

	ctracker *ChannelTracker

	db *gorm.DB
}

type User struct {
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

func (p *PermissionPlugin) CheckPermission(m *irc.Message, perm string) (bool, error) {
	user := p.ctracker.LookupUser(m.Prefix.Nick)
	if user == nil {
		// do something dramatic
	}

}
