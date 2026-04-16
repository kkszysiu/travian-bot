package parser

import "github.com/PuerkitoBio/goquery"

// GetAvailableMerchants parses the number of available merchants from the marketplace page.
func GetAvailableMerchants(doc *goquery.Document) int {
	sel := doc.Find(".merchantsInformation .available .value")
	if sel.Length() == 0 {
		return 0
	}
	v := ParseInt(sel.Text())
	if v < 0 {
		return 0
	}
	return v
}

// GetMerchantCapacity parses the merchant carry capacity from the marketplace page.
func GetMerchantCapacity(doc *goquery.Document) int {
	sel := doc.Find(".merchantsInformation .capacity .value")
	if sel.Length() == 0 {
		return 0
	}
	v := ParseInt(sel.Text())
	if v < 0 {
		return 0
	}
	return v
}

// GetSendResourceInputSelectors returns CSS selectors for the 4 send resource input fields.
func GetSendResourceInputSelectors() [4]string {
	return [4]string{
		"input[name='lumber']",
		"input[name='clay']",
		"input[name='iron']",
		"input[name='crop']",
	}
}

// GetCoordXInputSelector returns the CSS selector for the X coordinate input.
func GetCoordXInputSelector() string {
	return "label.coordinateX input"
}

// GetCoordYInputSelector returns the CSS selector for the Y coordinate input.
func GetCoordYInputSelector() string {
	return "label.coordinateY input"
}

// GetSendButtonSelector returns the CSS selector for the send resources button.
func GetSendButtonSelector() string {
	return "button.send"
}
