package parser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"travian-bot/internal/domain/enum"
)

// HeroItemDTO holds the parsed hero item type and amount.
type HeroItemDTO struct {
	Type   enum.HeroItem
	Amount int
}

// IsInventoryPage checks if the current page is showing the hero inventory tab.
// It looks for #heroV2 a.tabItem.active.
func IsInventoryPage(doc *goquery.Document) bool {
	heroDiv := doc.Find("#heroV2")
	if heroDiv.Length() == 0 {
		return false
	}
	tabItem := heroDiv.Find("a.tabItem")
	if tabItem.Length() == 0 {
		return false
	}
	return tabItem.HasClass("active")
}

// IsInventoryLoaded checks that the inventory wrapper has finished loading
// (i.e. it exists and does not have the "loading" class).
func IsInventoryLoaded(doc *goquery.Document) bool {
	wrapper := doc.Find("div.inventoryPageWrapper")
	if wrapper.Length() == 0 {
		return false
	}
	return !wrapper.HasClass("loading")
}

// GetHeroAvatarSelector returns the CSS selector for the hero avatar button
// that opens the hero inventory.
func GetHeroAvatarSelector() string {
	return "#heroImageButton"
}

// GetHeroItems parses all hero inventory items from the page.
func GetHeroItems(doc *goquery.Document) []HeroItemDTO {
	var items []HeroItemDTO

	heroItemsDiv := doc.Find("div.heroItems")
	if heroItemsDiv.Length() == 0 {
		return items
	}

	heroItemsDiv.Find("div.heroItem").Each(func(_ int, itemSlot *goquery.Selection) {
		if itemSlot.HasClass("empty") {
			return
		}

		children := itemSlot.Children()
		if children.Length() < 2 {
			return
		}

		itemNode := children.Eq(1)
		classAttr, exists := itemNode.Attr("class")
		if !exists {
			return
		}
		classes := strings.Fields(classAttr)
		if len(classes) < 2 {
			return
		}

		itemValue := ParseInt(classes[1])
		if itemValue <= 0 {
			return
		}
		itemType := enum.HeroItem(itemValue)
		if itemType == enum.HeroItemNone {
			return
		}

		amount := 1

		// Consumable items may have an amount badge
		dataTier, _ := itemSlot.Attr("data-tier")
		if strings.Contains(dataTier, "consumable") && children.Length() >= 3 {
			amountNode := children.Eq(2)
			parsed := ParseInt(amountNode.Text())
			if parsed > 0 {
				amount = parsed
			}
		}

		items = append(items, HeroItemDTO{
			Type:   itemType,
			Amount: amount,
		})
	})

	return items
}

// GetItemSlotSelector returns a CSS selector string that can be used to find
// a specific hero item slot by its item type. The selector finds a non-empty
// heroItem div whose child element has the item type as its second CSS class.
// Since goquery CSS selectors cannot directly match the second class, callers
// should use GetItemSlot instead for programmatic access.
func GetItemSlotSelector(itemType int) string {
	// This is a best-effort selector; exact matching requires DOM inspection.
	return "div.heroItems div.heroItem:not(.empty)"
}

// GetItemSlot finds the specific hero item slot element for the given item type.
// Returns the CSS selector path if found, empty string if not.
func GetItemSlot(doc *goquery.Document, itemType enum.HeroItem) *goquery.Selection {
	heroItemsDiv := doc.Find("div.heroItems")
	if heroItemsDiv.Length() == 0 {
		return nil
	}

	var found *goquery.Selection
	heroItemsDiv.Find("div.heroItem").Each(func(_ int, itemSlot *goquery.Selection) {
		if found != nil {
			return
		}
		if itemSlot.HasClass("empty") {
			return
		}

		children := itemSlot.Children()
		if children.Length() < 2 {
			return
		}

		itemNode := children.Eq(1)
		classAttr, exists := itemNode.Attr("class")
		if !exists {
			return
		}
		classes := strings.Fields(classAttr)
		if len(classes) < 2 {
			return
		}

		itemValue := ParseInt(classes[1])
		if enum.HeroItem(itemValue) == itemType {
			found = itemSlot
		}
	})

	return found
}

// GetAmountInputSelector returns the CSS selector for the consumable item
// amount input in the use-item dialog.
func GetAmountInputSelector() string {
	return "#consumableHeroItem input"
}

// GetConfirmUseItemButtonSelector returns the CSS selector for the confirm
// button in the hero item use dialog. It is the second button in the
// #dialogContent .buttonsWrapper.
func GetConfirmUseItemButtonSelector() string {
	return "#dialogContent .buttonsWrapper button:nth-child(2)"
}
