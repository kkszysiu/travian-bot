package parser

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// GetMaxTrainAmount parses the maximum trainable amount for a troop.
// Inside the matching troop node (div.troop:not(.empty) containing img.unit with
// class "uNN"), find the link in div.cta and parse its inner text.
func GetMaxTrainAmount(doc *goquery.Document, troopID int) int {
	var result int
	unitClass := fmt.Sprintf("u%d", troopID)
	doc.Find("div.troop:not(.empty)").Each(func(i int, s *goquery.Selection) {
		img := s.Find("img.unit")
		if img.Length() == 0 {
			return
		}
		if img.HasClass(unitClass) {
			a := s.Find("div.cta a")
			if a.Length() > 0 {
				result = ParseInt(a.Text())
			}
		}
	})
	return result
}

// GetTroopInputSelector returns the CSS selector for the training input box
// of a specific troop. It searches the document for the troop node containing
// the matching unit image, then returns a selector based on the input's name attribute.
func GetTroopInputSelector(doc *goquery.Document, troopID int) string {
	var selector string
	unitClass := fmt.Sprintf("u%d", troopID)
	doc.Find("div.troop:not(.empty)").Each(func(i int, s *goquery.Selection) {
		img := s.Find("img.unit")
		if img.Length() == 0 {
			return
		}
		if img.HasClass(unitClass) {
			input := s.Find("div.cta input.text")
			if input.Length() > 0 {
				name, exists := input.Attr("name")
				if exists {
					selector = fmt.Sprintf("input[name='%s']", name)
				}
			}
		}
	})
	return selector
}

// GetTrainButtonSelector returns the CSS selector for the train/submit button
// on a troop training page.
func GetTrainButtonSelector() string {
	return "#s1"
}
