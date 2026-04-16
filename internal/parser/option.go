package parser

import "github.com/PuerkitoBio/goquery"

// IsContextualHelpEnabled checks if contextual help is shown.
func IsContextualHelpEnabled(doc *goquery.Document) bool {
	return doc.Find("#contextualHelp").Length() > 0
}

// GetOptionButtonSelector returns the CSS selector for the options button.
func GetOptionButtonSelector() string {
	return "#outOfGame a.options"
}

// GetHideContextualHelpSelector returns the CSS selector for the hide contextual help checkbox.
func GetHideContextualHelpSelector() string {
	return "#hideContextualHelp"
}

// GetSubmitButtonSelector returns the CSS selector for the submit button.
func GetSubmitButtonSelector() string {
	return ".submitButtonContainer button"
}
