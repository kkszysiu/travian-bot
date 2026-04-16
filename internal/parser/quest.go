package parser

import "github.com/PuerkitoBio/goquery"

// GetQuestMasterSelector returns the CSS selector for the questmaster button.
func GetQuestMasterSelector() string {
	return "#questmasterButton"
}

// IsQuestClaimable checks if there's a claimable quest (new quest speech bubble).
func IsQuestClaimable(doc *goquery.Document) bool {
	questmaster := doc.Find("#questmasterButton")
	if questmaster.Length() == 0 {
		return false
	}
	return questmaster.Find("div.newQuestSpeechBubble").Length() > 0
}

// GetQuestCollectButtonSelector returns the CSS selector for the quest collect button.
// It's inside div.taskOverview, a button.collect that is not .disabled.
func GetQuestCollectButtonSelector() string {
	return "div.taskOverview button.collect:not(.disabled)"
}

// HasQuestCollectButton checks if a collect button exists on the page.
func HasQuestCollectButton(doc *goquery.Document) bool {
	return doc.Find("div.taskOverview button.collect:not(.disabled)").Length() > 0
}

// IsQuestPage checks if the quest page is shown.
func IsQuestPage(doc *goquery.Document) bool {
	return doc.Find("div.tasks.tasksVillage").Length() > 0
}
