package plugins

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/jmoiron/sqlx"

	"github.com/belak/go-seabird/seabird"
	"github.com/belak/irc"
)

func init() {
	seabird.RegisterPlugin("karma", newKarmaPlugin)
}

type karmaUser struct {
	Name  string
	Score int
}

type karmaPlugin struct {
	db *sqlx.DB
}

var regex = regexp.MustCompile(`([^\s]+)(\+\+|--)(?:\s|$)`)

func newKarmaPlugin(b *seabird.Bot, m *seabird.BasicMux, cm *seabird.CommandMux, db *sqlx.DB) {
	p := &karmaPlugin{db: db}

	cm.Event("karma", p.karmaCallback, &seabird.HelpInfo{
		Usage:       "<nick>",
		Description: "Gives karma for given user",
	})

	cm.Event("topkarma", p.topKarmaCallback, &seabird.HelpInfo{
		Description: "Reports the user with the most karma",
	})

	cm.Event("bottomkarma", p.bottomKarmaCallback, &seabird.HelpInfo{
		Description: "Reports the user with the least karma",
	})

	m.Event("PRIVMSG", p.callback)
}

func (p *karmaPlugin) cleanedName(name string) string {
	return strings.TrimFunc(strings.ToLower(name), unicode.IsSpace)
}

// GetKarmaFor returns the karma for the given name.
func (p *karmaPlugin) GetKarmaFor(name string) int {
	var score int
	err := p.db.Get(&score, "SELECT score FROM karma WHERE name=$1", p.cleanedName(name))
	if err != nil {
		return 0
	}

	return score
}

// UpdateKarma will update the karma for a given name and return the new karma value.
func (p *karmaPlugin) UpdateKarma(name string, diff int) int {
	_, err := p.db.Exec("INSERT INTO karma (name, score) VALUES ($1, $2)", p.cleanedName(name), diff)
	// If it was a nil error, we got the insert
	if err == nil {
		return diff
	}

	// Grab a transaction, just in case
	tx, err := p.db.Beginx()
	defer tx.Commit()

	if err != nil {
		fmt.Println("TX:", err)
	}

	// If there was an error, we try an update.
	_, err = tx.Exec("UPDATE karma SET score=score+$1 WHERE name=$2", diff, p.cleanedName(name))
	if err != nil {
		fmt.Println("UPDATE:", err)
	}

	var score int
	err = tx.Get(&score, "SELECT score FROM karma WHERE name=$1", p.cleanedName(name))
	if err != nil {
		fmt.Println("SELECT:", err)
	}

	return score
}

func (p *karmaPlugin) karmaCallback(b *seabird.Bot, m *irc.Message) {
	term := strings.TrimSpace(m.Trailing())

	// If we don't provide a term, search for the current nick
	if term == "" {
		term = m.Prefix.Name
	}

	b.MentionReply(m, "%s's karma is %d", term, p.GetKarmaFor(term))
}

func (p *karmaPlugin) karmaCheck(b *seabird.Bot, m *irc.Message, msg string, sort string) {
	user := &karmaUser{}
	err := p.db.Get(user, fmt.Sprintf("SELECT name, score FROM karma ORDER BY score %s LIMIT 1", sort))
	if err != nil {
		b.MentionReply(m, "Error fetching scores")
		return
	}

	b.MentionReply(m, "%s has the %s karma with %d", user.Name, msg, user.Score)
}
func (p *karmaPlugin) topKarmaCallback(b *seabird.Bot, m *irc.Message) {
	p.karmaCheck(b, m, "top", "DESC")
}

func (p *karmaPlugin) bottomKarmaCallback(b *seabird.Bot, m *irc.Message) {
	p.karmaCheck(b, m, "bottom", "ASC")
}

func (p *karmaPlugin) callback(b *seabird.Bot, m *irc.Message) {
	if len(m.Params) < 2 || !m.FromChannel() {
		return
	}

	matches := regex.FindAllStringSubmatch(m.Trailing(), -1)
	if len(matches) > 0 {
		for _, v := range matches {
			if len(v) < 3 {
				continue
			}

			var diff int
			if v[2] == "++" {
				diff = 1
			} else {
				diff = -1
			}

			name := strings.ToLower(v[1])
			if name == m.Prefix.Name {
				// penalize self-karma
				diff = -1
			}

			b.Reply(m, "%s's karma is now %d", v[1], p.UpdateKarma(name, diff))
		}
	}
}
