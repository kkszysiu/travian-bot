package parser

import "github.com/PuerkitoBio/goquery"

// GetAdventureDuration returns the adventure duration in seconds from the hero adventure section.
// Finds #heroAdventure span.timer, reads its "value" attribute (seconds).
func GetAdventureDuration(doc *goquery.Document) int {
	timer := doc.Find("#heroAdventure span.timer")
	if timer.Length() == 0 {
		return 0
	}
	val, _ := timer.Attr("value")
	return ParseInt(val)
}

// IsAdventurePage checks if the current page shows the adventure list table.
func IsAdventurePage(doc *goquery.Document) bool {
	return doc.Find("table.adventureList").Length() > 0
}

// GetHeroAdventureButtonSelector returns the CSS selector for the hero adventure button.
func GetHeroAdventureButtonSelector() string {
	return "a.adventure.round"
}

// CanStartAdventure checks if the hero is home and adventures are available.
func CanStartAdventure(doc *goquery.Document) bool {
	heroStatus := doc.Find("div.heroStatus")
	if heroStatus.Length() == 0 {
		return false
	}
	if heroStatus.Find("i.heroHome").Length() == 0 {
		return false
	}
	adventureBtn := doc.Find("a.adventure.round")
	if adventureBtn.Length() == 0 {
		return false
	}
	return adventureBtn.Find("div.content").Length() > 0
}

// GetAdventureButtonSelector returns the CSS selector for the first adventure's start button.
// Finds #heroAdventure tbody tr:first-child button
func GetAdventureButtonSelector() string {
	return "#heroAdventure tbody tr:first-child button"
}

// GetContinueButtonSelector returns the selector for the continue button after starting an adventure.
func GetContinueButtonSelector() string {
	return "button.continue"
}
