package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// RallyPointTroopSlot represents one troop input on the rally point send form.
type RallyPointTroopSlot struct {
	InputName string // e.g., "troop[t1]", "troop[t2]", ... "troop[t11]"
	Available int    // troops available to send
}

// GetRallyPointTroopSlots parses all troop input fields from the Rally Point "send troops" tab.
// In Travian, troop inputs are named troop[t1] through troop[t11].
// Available counts are in <a> tags next to each input (disabled troops have <span class="none">).
func GetRallyPointTroopSlots(doc *goquery.Document) []RallyPointTroopSlot {
	var slots []RallyPointTroopSlot

	for i := 1; i <= 11; i++ {
		name := fmt.Sprintf("troop[t%d]", i)
		selector := fmt.Sprintf("input[name='troop[t%d]']", i)
		inputEl := doc.Find(selector)
		if inputEl.Length() == 0 {
			continue
		}

		// Skip disabled inputs (troops not available in this village)
		if _, disabled := inputEl.Attr("disabled"); disabled {
			continue
		}

		// The available count is in an <a> tag within the same <td> parent.
		// The <a> has an onclick that sets the input value to the max count.
		available := 0
		inputEl.Parent().Find("a").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				v := ParseInt(text)
				if v > 0 {
					available = v
				}
			}
		})

		if available > 0 {
			slots = append(slots, RallyPointTroopSlot{
				InputName: name,
				Available: available,
			})
		}
	}

	return slots
}

// GetRallyPointTroopInputSelector returns the CSS selector for a specific troop input.
func GetRallyPointTroopInputSelector(name string) string {
	return fmt.Sprintf("input[name='%s']", name)
}

// GetRallyPointCoordXSelector returns the CSS selector for the X coordinate input on the rally point.
func GetRallyPointCoordXSelector() string {
	return "#xCoordInput"
}

// GetRallyPointCoordYSelector returns the CSS selector for the Y coordinate input on the rally point.
func GetRallyPointCoordYSelector() string {
	return "#yCoordInput"
}

// GetReinforcementRadioSelector returns the CSS selector for the "Reinforcement" radio button.
// In Travian, movement type radios use name='eventType' with value=5 for reinforcements.
func GetReinforcementRadioSelector() string {
	return "input[name='eventType'][value='5']"
}

// GetRallyPointSendButtonSelector returns the CSS selector for the send button on the rally point.
func GetRallyPointSendButtonSelector() string {
	return "button#ok"
}

// GetRallyPointConfirmButtonSelector returns the CSS selector for the confirm button
// on the rally point confirmation page (button id="confirmSendTroops").
func GetRallyPointConfirmButtonSelector() string {
	return "button#confirmSendTroops"
}

// GetRecallTroopsSelector returns the CSS selector for the "return" / recall link
// on the rally point overview tab (tt=1).
// The recall link is an <a class="arrow"> inside a <div class="sback">.
// Text is "powrót" (return) for reinforcements sent to other villages.
func GetRecallTroopsSelector() string {
	return "div.sback a.arrow"
}
