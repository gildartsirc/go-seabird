package extra

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/belak/go-seabird"
	"github.com/belak/nut"
	"github.com/go-irc/irc"
)

func init() {
	seabird.RegisterPlugin("karma", newKarmaPlugin)
}

type karmaPlugin struct {
	db *nut.DB
}

// KarmaTarget represents an item with a karma count
type KarmaTarget struct {
	Name  string
	Score int
}

var regex = regexp.MustCompile(`([\w]{2,}|".+?")(\+\++|--+)(?:\s|$)`)

func newKarmaPlugin(b *seabird.Bot, m *seabird.BasicMux, cm *seabird.CommandMux, db *nut.DB) error {
	p := &karmaPlugin{db: db}

	err := p.db.EnsureBucket("karma")
	if err != nil {
		return err
	}

	cm.Event("karma", p.karmaCallback, &seabird.HelpInfo{
		Usage:       "<nick>",
		Description: "Displays karma for given user",
	})

	/*
		cm.Event("topkarma", p.topKarmaCallback, &seabird.HelpInfo{
			Description: "Reports the user with the most karma",
		})

		cm.Event("bottomkarma", p.bottomKarmaCallback, &seabird.HelpInfo{
			Description: "Reports the user with the least karma",
		})
	*/

	m.Event("PRIVMSG", p.callback)

	return nil
}

func (p *karmaPlugin) cleanedName(name string) string {
	return strings.TrimFunc(strings.ToLower(name), unicode.IsSpace)
}

// GetKarmaFor returns the karma for the given name.
func (p *karmaPlugin) GetKarmaFor(name string) int {
	out := &KarmaTarget{Name: p.cleanedName(name)}

	_ = p.db.View(func(tx *nut.Tx) error {
		bucket := tx.Bucket("karma")
		return bucket.Get(out.Name, out)
	})

	return out.Score
}

// UpdateKarma will update the karma for a given name and return the new karma value.
func (p *karmaPlugin) UpdateKarma(name string, diff int) int {
	out := &KarmaTarget{Name: p.cleanedName(name)}

	_ = p.db.Update(func(tx *nut.Tx) error {
		bucket := tx.Bucket("karma")
		bucket.Get(out.Name, out)
		out.Score += diff
		return bucket.Put(out.Name, out)
	})

	return out.Score
}

func (p *karmaPlugin) karmaCallback(b *seabird.Bot, m *irc.Message) {
	term := strings.TrimSpace(m.Trailing())

	// If we don't provide a term, search for the current nick
	if term == "" {
		term = m.Prefix.Name
	}

	b.MentionReply(m, "%s's karma is %d", term, p.GetKarmaFor(term))
}

/*
func (p *karmaPlugin) karmaCheck(b *seabird.Bot, m *irc.Message, msg string, sort string) {
	res := &KarmaTarget{}
	p.db.Order("score " + sort).First(res)
	b.MentionReply(m, "%s has the %s karma with %d", res.Name, msg, res.Score)
}
func (p *karmaPlugin) topKarmaCallback(b *seabird.Bot, m *irc.Message) {
	p.karmaCheck(b, m, "top", "DESC")
}

func (p *karmaPlugin) bottomKarmaCallback(b *seabird.Bot, m *irc.Message) {
	p.karmaCheck(b, m, "bottom", "ASC")
}
*/

func (p *karmaPlugin) callback(b *seabird.Bot, m *irc.Message) {
	if len(m.Params) < 2 || !b.FromChannel(m) {
		return
	}

	var buzzkillTriggered bool
	var changes = make(map[string]int)

	matches := regex.FindAllStringSubmatch(m.Trailing(), -1)
	for _, v := range matches {
		// If it starts with a ", we know it also ends with a quote so we
		// can chop them off.
		if strings.HasPrefix(v[1], "\"") {
			v[1] = v[1][1 : len(v[1])-1]
		}

		diff := len(v[2]) - 1
		cleanedName := p.cleanedName(v[1])
		cleanedNick := p.cleanedName(m.Prefix.Name)

		// If it's negative, or positive and someone is trying to change
		// their own karma we need to reverse the sign.
		if v[2][0] == '-' || cleanedName == cleanedNick {
			diff *= -1
		}

		changes[v[1]] = changes[v[1]] + diff
	}

	for name, diff := range changes {
		if diff > 5 {
			buzzkillTriggered = true
			diff = 5
		}

		b.Reply(m, "%s's karma is now %d", name, p.UpdateKarma(name, diff))
	}

	if buzzkillTriggered {
		b.Reply(m, "Buzzkill Mode (tm) enforced a maximum karma change of 5")
	}
}
