package plugins

import (
	"github.com/belak/go-seabird"
	"github.com/belak/nut"
	"github.com/jinzhu/gorm"
	"seabird"
)

func init() {
	seabird.RegisterPlugin("sqldb", newSQLDBPlugin)
}

type dbConfig struct {
	Dialect    string
	Connection string
}

func newSQLDBPlugin(b *seabird.Bot) (*gorm.DB, error) {
	dbc := &dbConfig{}
	err := b.Config("sqldb", dbc)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dbc.Dialect, dbc.Connection)
	if err != nil {
		return nil, err
	}

	return db, nil
}
