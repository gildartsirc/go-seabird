package plugins

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/belak/irc"
	"github.com/belak/go-seabird/bot"
)

func init() {
	bot.RegisterPlugin("fcc", NewFccPlugin)
}

type fccPlugin struct {
	Key string
}

type fccLicense struct {
	Name       string `json:"licName"`
	Frn        string `json:"frn"`
	Callsign   string `json:"callsign"`
	Category   string `json:"categoryDesc"`
	Service    string `json:"serviceDesc"`
	Status     string `json:"statusDesc"`
	ExpireDate string `json:"expiredDate"`
	LicenseId  string `json:"licenseID"`
	LicenseUrl string `json:"licDetailURL"`
}

type fccLicenses struct {
	Page       string       `json:"page"`
	RowPerPage string       `json:"rowPerPage"`
	TotalRows  string       `json:"totalRows"`
	LastUpdate string       `json:"lastUpdate"`
	Licenses   []fccLicense `json:"License"`
}

type fccResponse struct {
	Status      string      `json:"status"`
	LicenseData fccLicenses `json:"Licenses"`
}

func NewFccPlugin(b *bot.Bot) (bot.Plugin, error) {
	p := &fccPlugin{}

	b.Config("fcc", p)

	b.CommandMux.Event("call", p.Search, &bot.HelpInfo{
		Usage:       "<callsign>",
		Description: "Finds information about given FCC callsign",
	})

	return p, nil
}

func (p *fccPlugin) Search(b *bot.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Callsign required")
			return
		}
		url := "http://data.fcc.gov/api/license-view/basicSearch/getLicenses?format=json&searchValue=" + url.QueryEscape(m.Trailing())
		resp, err := http.Get(url)
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}
		defer resp.Body.Close()

		fr := &fccResponse{}
		err = json.NewDecoder(resp.Body).Decode(fr)
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		if len(fr.LicenseData.Licenses) == 0 {
			b.MentionReply(m, "No licenses found")
			return
		}

		license := fr.LicenseData.Licenses[0]
		b.MentionReply(m, "%s (%s): %s, %s, expires %s", license.Callsign, license.Service, license.Name, license.Status, license.ExpireDate)
	}()
}
