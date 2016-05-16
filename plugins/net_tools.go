package plugins

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/belak/irc"
	"github.com/belak/go-seabird/bot"
)

func init() {
	bot.RegisterPlugin("nettools", NewNetToolsPlugin)
}

type netToolsPlugin struct {
	Key string
}

func NewNetToolsPlugin(b *bot.Bot) (bot.Plugin, error) {
	p := &netToolsPlugin{}

	b.Config("net_tools", p)

	b.CommandMux.Event("rdns", p.RDNS, &bot.HelpInfo{
		Usage:       "<ip>",
		Description: "Does a reverse DNS lookup on the given IP",
	})
	b.CommandMux.Event("dig", p.Dig, &bot.HelpInfo{
		Usage:       "<domain>",
		Description: "Retrieves IP records for given domain",
	})
	b.CommandMux.Event("ping", p.Ping, &bot.HelpInfo{
		Usage:       "<host>",
		Description: "Pings given host once",
	})
	b.CommandMux.Event("traceroute", p.Traceroute, &bot.HelpInfo{
		Usage:       "<host>",
		Description: "Runs traceroute on given host and returns pastebin URL for results",
	})
	b.CommandMux.Event("whois", p.Whois, &bot.HelpInfo{
		Usage:       "<domain>",
		Description: "Runs whois on given domain and returns pastebin URL for results",
	})
	b.CommandMux.Event("dnscheck", p.DNSCheck, &bot.HelpInfo{
		Usage:       "<domain>",
		Description: "Returns DNSCheck URL for domain",
	})

	return p, nil
}

func (p *netToolsPlugin) RDNS(b *bot.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Argument required")
			return
		}
		names, err := net.LookupAddr(m.Trailing())
		if err != nil {
			b.MentionReply(m, err.Error())
			return
		}

		if len(names) == 0 {
			b.MentionReply(m, "No results found")
			return
		}

		b.MentionReply(m, names[0])

		if len(names) > 1 {
			for _, name := range names[1:] {
				b.Writef("NOTICE %s :%s", m.Prefix.Name, name)
			}
		}
	}()
}

func (p *netToolsPlugin) Dig(b *bot.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Domain required")
			return
		}

		addrs, err := net.LookupHost(m.Trailing())
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		if len(addrs) == 0 {
			b.MentionReply(m, "No results found")
			return
		}

		b.MentionReply(m, addrs[0])

		if len(addrs) > 1 {
			for _, addr := range addrs[1:] {
				b.Writef("NOTICE %s :%s", m.Prefix.Name, addr)
			}
		}
	}()
}

func (p *netToolsPlugin) Ping(b *bot.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Host required")
			return
		}

		out, err := exec.Command("ping", "-c1", m.Trailing()).Output()
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		arr := strings.Split(string(out), "\n")
		if len(arr) < 2 {
			b.MentionReply(m, "Error retrieving ping results")
			return
		}

		b.MentionReply(m, arr[1])
	}()
}

func (p *netToolsPlugin) Traceroute(b *bot.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Host required")
			return
		}

		out, err := exec.Command("traceroute", m.Trailing()).Output()
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		resp, err := http.PostForm("http://pastebin.com/api/api_post.php", url.Values{
			"api_dev_key":    {p.Key},
			"api_option":     {"paste"},
			"api_paste_code": {string(out)},
		})
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		b.MentionReply(m, "%s", body)
	}()
}

func (p *netToolsPlugin) Whois(b *bot.Bot, m *irc.Message) {
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Domain required")
			return
		}

		out, err := exec.Command("whois", m.Trailing()).Output()
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		resp, err := http.PostForm("http://pastebin.com/api/api_post.php", url.Values{
			"api_dev_key":    {p.Key},
			"api_option":     {"paste"},
			"api_paste_code": {string(out)},
		})
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			b.MentionReply(m, "%s", err)
			return
		}

		b.MentionReply(m, "%s", body)
	}()
}

func (p *netToolsPlugin) DNSCheck(b *bot.Bot, m *irc.Message) {
	// Just for Kaleb
	go func() {
		if m.Trailing() == "" {
			b.MentionReply(m, "Domain required")
			return
		}

		b.MentionReply(m, "https://www.whatsmydns.net/#A/"+m.Trailing())
	}()
}
