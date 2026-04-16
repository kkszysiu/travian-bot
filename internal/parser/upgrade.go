package parser

import (
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// findContractNode locates the contract element for a given building type.
// It first tries #contract_buildingNN, then falls back to #contract.
func findContractNode(doc *goquery.Document, buildingType int) *goquery.Selection {
	sel := doc.Find(fmt.Sprintf("#contract_building%d", buildingType))
	if sel.Length() > 0 {
		return sel
	}
	sel = doc.Find("#contract")
	if sel.Length() > 0 {
		return sel
	}
	return nil
}

// GetRequiredResource parses the resource costs from the contract section.
// Returns a slice of 5 int64 values: [wood, clay, iron, crop, freeCrop].
func GetRequiredResource(doc *goquery.Document, buildingType int) []int64 {
	contract := findContractNode(doc, buildingType)
	if contract == nil {
		return nil
	}

	resources := contract.Find("div.resourceWrapper div.resource")
	if resources.Length() < 5 {
		return nil
	}

	result := make([]int64, 5)
	resources.Each(func(i int, s *goquery.Selection) {
		if i >= 5 {
			return
		}
		result[i] = ParseLong(s.Text())
	})
	return result
}

// GetTimeWhenEnoughResource parses the time remaining until enough resources
// are available from the contract's error message timer.
func GetTimeWhenEnoughResource(doc *goquery.Document, buildingType int) time.Duration {
	contract := findContractNode(doc, buildingType)
	if contract == nil {
		return 0
	}

	timer := contract.Find("div.errorMessage span.timer")
	if timer.Length() == 0 {
		return 0
	}

	val, exists := timer.Attr("value")
	if !exists {
		return 0
	}

	seconds := ParseInt(val)
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

// GetConstructButtonSelector returns the CSS selector for the construct/upgrade button.
// For resource fields (buildingType 1-4) it uses the upgrade button; otherwise
// it uses the contract-specific build button.
func GetConstructButtonSelector(buildingType int) string {
	if buildingType >= 1 && buildingType <= 4 {
		return GetUpgradeButtonSelector()
	}
	return fmt.Sprintf("#contract_building%d button.new", buildingType)
}

// GetSpecialUpgradeButtonSelector returns the CSS selector for the special
// (video feature) upgrade button.
func GetSpecialUpgradeButtonSelector() string {
	return "div.upgradeButtonsContainer button.videoFeatureButton.green"
}

// GetUpgradeButtonSelector returns the CSS selector for the standard upgrade button.
func GetUpgradeButtonSelector() string {
	return "div.upgradeButtonsContainer button.build"
}
