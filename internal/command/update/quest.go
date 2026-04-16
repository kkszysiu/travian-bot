package update

import (
	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// CheckQuestAvailable parses the current page to check if there are
// claimable quests. Returns true if quests can be claimed.
func CheckQuestAvailable(b *browser.Browser) bool {
	html, err := b.PageHTML()
	if err != nil {
		return false
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return false
	}
	return parser.IsQuestClaimable(doc)
}
