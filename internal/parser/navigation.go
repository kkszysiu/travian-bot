package parser

import "fmt"

// GetDorfButtonSelector returns the CSS selector for dorf1 or dorf2 navigation.
func GetDorfButtonSelector(dorf int) string {
	return fmt.Sprintf("#navigation a[accesskey='%d']", dorf)
}
