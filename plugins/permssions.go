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
	Type       string // 'mask', 'account', and 'channel' supported.
	Identifier string // either a *!*@* mask, an account name, or a channel name.
}

type Role struct {
	gorm.Model
	Name string
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

}
