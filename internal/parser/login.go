package parser

import "github.com/PuerkitoBio/goquery"

// IsIngamePage checks if the page is the in-game page (has #servertime element).
func IsIngamePage(doc *goquery.Document) bool {
	return doc.Find("#servertime").Length() > 0
}

// IsLoginPage checks if the page is the login page.
func IsLoginPage(doc *goquery.Document) bool {
	return doc.Find("#loginScene button.green").Length() > 0
}

// GetLoginButtonSelector returns the CSS selector for the login button.
func GetLoginButtonSelector() string {
	return "#loginScene button.green"
}

// GetUsernameInputSelector returns the CSS selector for the username input.
func GetUsernameInputSelector() string {
	return "input[name='name']"
}

// GetPasswordInputSelector returns the CSS selector for the password input.
func GetPasswordInputSelector() string {
	return "input[name='password']"
}
