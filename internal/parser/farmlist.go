package parser

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// FarmListInfo holds parsed farm list data from the rally point farm list page.
type FarmListInfo struct {
	ID   int
	Name string
}

// GetFarmLists parses all farm lists from the rally point farm list tab.
func GetFarmLists(doc *goquery.Document) []FarmListInfo {
	var farms []FarmListInfo
	doc.Find("#rallyPointFarmList div.farmListHeader").Each(func(i int, s *goquery.Selection) {
		dragDrop := s.Find("div.dragAndDrop")
		idStr, _ := dragDrop.Attr("data-list")
		id := ParseInt(idStr)

		name := strings.TrimSpace(s.Find("div.name").Text())

		if id > 0 {
			farms = append(farms, FarmListInfo{ID: id, Name: name})
		}
	})
	return farms
}

// GetStartFarmListButtonSelector returns the CSS selector for a specific farm
// list's start button. The button is inside the farmListHeader that contains
// a div.dragAndDrop with the matching data-list attribute.
func GetStartFarmListButtonSelector(farmListID int) string {
	return fmt.Sprintf(
		"#rallyPointFarmList div.farmListHeader:has(div.dragAndDrop[data-list='%d']) button.startFarmList",
		farmListID,
	)
}

// GetStartAllFarmListButtonSelector returns the CSS selector for the "Start All"
// button on the rally point farm list page.
func GetStartAllFarmListButtonSelector() string {
	return "#rallyPointFarmList button.startAllFarmLists"
}
