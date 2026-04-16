package parser

import "github.com/PuerkitoBio/goquery"

// CountQueueBuilding counts the number of buildings in the construction queue.
func CountQueueBuilding(doc *goquery.Document) int {
	finishDiv := doc.Find("div.finishNow")
	if finishDiv.Length() == 0 {
		return 0
	}
	return finishDiv.Parent().Find("li").Length()
}

// GetCompleteButtonSelector returns the CSS selector for the "finish now" button.
func GetCompleteButtonSelector() string {
	return "div.finishNow button"
}

// GetConfirmButtonSelector returns the CSS selector for the confirm button in the finish dialog.
func GetConfirmButtonSelector() string {
	return "#finishNowDialog button"
}
