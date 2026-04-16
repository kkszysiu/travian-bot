package parser

import "github.com/PuerkitoBio/goquery"

// IsNpcDialog checks if the NPC dialog is visible.
func IsNpcDialog(doc *goquery.Document) bool {
	return doc.Find("#npc").Length() > 0
}

// GetExchangeResourcesButtonSelector returns the CSS selector for the exchange resources button.
func GetExchangeResourcesButtonSelector() string {
	return "div.npcMerchant button.gold"
}

// GetDistributeButtonSelector returns the CSS selector for the distribute button in the NPC dialog.
func GetDistributeButtonSelector() string {
	return "div.exchangeResources #submitText button.textButtonV1.gold"
}

// GetRedeemButtonSelector returns the CSS selector for the redeem button.
func GetRedeemButtonSelector() string {
	return "#npc_market_button"
}

// GetNpcSum reads the total resource sum from the NPC dialog.
func GetNpcSum(doc *goquery.Document) int64 {
	sum := doc.Find("#sum")
	if sum.Length() == 0 {
		return -1
	}
	return ParseLong(sum.Text())
}

// GetNpcInputSelectors returns CSS selectors for the 4 NPC resource input fields (wood, clay, iron, crop).
func GetNpcInputSelectors() [4]string {
	return [4]string{
		"input[name='desired0']",
		"input[name='desired1']",
		"input[name='desired2']",
		"input[name='desired3']",
	}
}
