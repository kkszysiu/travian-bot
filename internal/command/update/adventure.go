package update

import (
	"travian-bot/internal/browser"
	"travian-bot/internal/parser"
)

// CheckAdventureAvailable parses the current page to check if the hero
// can start an adventure. Returns true if adventures are available.
func CheckAdventureAvailable(b *browser.Browser) bool {
	html, err := b.PageHTML()
	if err != nil {
		return false
	}
	doc, err := parser.DocFromHTML(html)
	if err != nil {
		return false
	}
	return parser.CanStartAdventure(doc)
}
