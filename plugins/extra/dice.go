package extra

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/belak/go-seabird"
	"github.com/go-irc/irc"
)

func init() {
	seabird.RegisterPlugin("dice", newDicePlugin)
}

var diceRe = regexp.MustCompile(`(?:^|\b)(\d*)d(\d+)\b`)

func newDicePlugin(b *seabird.Bot, mm *seabird.MentionMux) {
	mm.Event(diceCallback)
}

func diceCallback(b *seabird.Bot, m *irc.Message) {
	var rolls []string
	totalCount := 0

	matches := diceRe.FindAllStringSubmatch(m.Trailing(), -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		// Grab the count, otherwise 1
		count, _ := strconv.Atoi(match[1])
		if count == 0 {
			count = 1
		}

		// Clamp count
		if count < 0 {
			b.MentionReply(m, "You cannot request a negative number of rolls")
			return
		}

		totalCount += count
		if totalCount > 100 {
			b.MentionReply(m, "You cannot request more than 100 dice")
			return
		}

		// How big is the die?
		size, _ := strconv.Atoi(match[2])

		if size > 100 {
			b.MentionReply(m, "You cannot request dice larger than 100")
			return
		}

		// Clamp size
		if size < 1 {
			b.MentionReply(m, "You cannot request dice smaller than 1")
			return
		}

		var dice []string
		for i := 0; i < count; i++ {
			dice = append(dice, fmt.Sprintf("%d", rand.Intn(size)+1))
		}

		rolls = append(rolls, fmt.Sprintf("%dd%d: %s", count, size, strings.Join(dice, ", ")))
	}

	if len(rolls) > 0 {
		b.MentionReply(m, "%s", strings.Join(rolls, " "))
	}
}
