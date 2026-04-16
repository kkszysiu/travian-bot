package parser

import (
	"github.com/PuerkitoBio/goquery"
)

// CountTab returns the number of tab items in the building content navigation.
func CountTab(doc *goquery.Document) int {
	return doc.Find("div.contentNavi.subNavi a.tabItem").Length()
}

// GetTabHref returns the href of the Nth tab (0-based index), or empty string if not found.
func GetTabHref(doc *goquery.Document, index int) string {
	tabs := doc.Find("div.contentNavi.subNavi a.tabItem")
	if index < 0 || index >= tabs.Length() {
		return ""
	}
	href, _ := tabs.Eq(index).Attr("href")
	return href
}

// IsTabActive checks if the tab at the given index (0-based) has the "active" class.
func IsTabActive(doc *goquery.Document, index int) bool {
	tabs := doc.Find("div.contentNavi.subNavi a.tabItem")
	if index < 0 || index >= tabs.Length() {
		return false
	}
	return tabs.Eq(index).HasClass("active")
}
