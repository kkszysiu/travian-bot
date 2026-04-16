package parser

import "github.com/PuerkitoBio/goquery"

// GetGold parses the gold amount from the sidebar.
func GetGold(doc *goquery.Document) int {
	sel := doc.Find("div.ajaxReplaceableGoldAmount")
	if sel.Length() == 0 {
		return -1
	}
	return ParseInt(sel.Text())
}

// GetSilver parses the silver amount from the sidebar.
func GetSilver(doc *goquery.Document) int {
	sel := doc.Find("div.ajaxReplaceableSilverAmount")
	if sel.Length() == 0 {
		return -1
	}
	return ParseInt(sel.Text())
}

// HasPlusAccount checks whether the account has an active Plus subscription.
// It looks for the edit link in the sidebar linklist: "green" means active, "gold" means inactive.
func HasPlusAccount(doc *goquery.Document) bool {
	sel := doc.Find("#sidebarBoxLinklist a.edit.round")
	if sel.Length() == 0 {
		return false
	}
	return sel.HasClass("green")
}
