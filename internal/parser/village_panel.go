package parser

import (
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// VillageInfo holds parsed village data from the sidebar.
type VillageInfo struct {
	ID            int
	Name          string
	X             int
	Y             int
	IsActive      bool
	IsUnderAttack bool
}

// villageSidebarIDs lists the known element IDs for the village sidebar box.
// Older Travian versions use lowercase "list", newer ones use camelCase "List".
var villageSidebarIDs = []string{
	"sidebarBoxVillageList",  // modern Travian (camelCase)
	"sidebarBoxVillagelist",  // older Travian (lowercase)
}

// findVillageSidebar finds the village sidebar box in the document, trying known ID variants.
func findVillageSidebar(doc *goquery.Document) *goquery.Selection {
	for _, id := range villageSidebarIDs {
		sel := doc.Find("#" + id)
		if sel.Length() > 0 {
			return sel
		}
	}
	return nil
}

// GetVillages parses all villages from the sidebar village panel.
func GetVillages(doc *goquery.Document) []VillageInfo {
	sidebar := findVillageSidebar(doc)
	if sidebar == nil {
		return nil
	}

	var villages []VillageInfo
	sidebar.Find("div.listEntry").Each(func(i int, s *goquery.Selection) {
		did, _ := s.Attr("data-did")
		id := ParseInt(did)
		if id <= 0 {
			return // skip entries without a valid data-did
		}

		name := s.Find("a span.name").Text()

		xText := s.Find("span.coordinateX").Text()
		yText := s.Find("span.coordinateY").Text()
		x := ParseInt(xText)
		y := ParseInt(yText)

		isActive := s.HasClass("active")
		isUnderAttack := s.HasClass("attack")

		villages = append(villages, VillageInfo{
			ID:            id,
			Name:          name,
			X:             x,
			Y:             y,
			IsActive:      isActive,
			IsUnderAttack: isUnderAttack,
		})
	})
	return villages
}

// GetCurrentVillageID returns the active village's data-did.
func GetCurrentVillageID(doc *goquery.Document) int {
	sidebar := findVillageSidebar(doc)
	if sidebar == nil {
		return 0
	}
	active := sidebar.Find("div.listEntry.active")
	if active.Length() == 0 {
		return 0
	}
	did, _ := active.Attr("data-did")
	return ParseInt(did)
}

// VillageSidebarSelectors returns CSS selectors for a specific village by data-did,
// trying both known sidebar IDs.
func VillageSidebarSelectors(villageID int) []string {
	did := strconv.Itoa(villageID)
	selectors := make([]string, len(villageSidebarIDs))
	for i, id := range villageSidebarIDs {
		selectors[i] = "#" + id + " div.listEntry[data-did='" + did + "']"
	}
	return selectors
}

// VillageSidebarActiveSelectors returns CSS selectors for an active village by data-did.
func VillageSidebarActiveSelectors(villageID int) []string {
	did := strconv.Itoa(villageID)
	selectors := make([]string, len(villageSidebarIDs))
	for i, id := range villageSidebarIDs {
		selectors[i] = "#" + id + " div.listEntry.active[data-did='" + did + "']"
	}
	return selectors
}
