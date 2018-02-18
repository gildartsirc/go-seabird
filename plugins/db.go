package extra

import (
	"github.com/belak/go-seabird"
	"github.com/belak/nut"
	"github.com/jinzhu/gorm"
	"seabird"
)

func init() {
	seabird.RegisterPlugin("db", newDBPlugin)
}

type dbConfig struct {
	Dialect    string
	Connection string
}

func newDBPlugin(b *seabird.Bot) (*gorm.DB, error) {
	dbc := &dbConfig{}
	err := b.Config("db", dbc)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dbc.Dialect, dbc.Connection)
	if err != nil {
		return nil, err
	}

	return db, nil
}
