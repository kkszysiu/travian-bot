package parser

import "github.com/PuerkitoBio/goquery"

// GetWood parses the wood resource amount.
func GetWood(doc *goquery.Document) int64 {
	return getResource(doc, "#l1")
}

// GetClay parses the clay resource amount.
func GetClay(doc *goquery.Document) int64 {
	return getResource(doc, "#l2")
}

// GetIron parses the iron resource amount.
func GetIron(doc *goquery.Document) int64 {
	return getResource(doc, "#l3")
}

// GetCrop parses the crop resource amount.
func GetCrop(doc *goquery.Document) int64 {
	return getResource(doc, "#l4")
}

// GetFreeCrop parses the free crop value.
func GetFreeCrop(doc *goquery.Document) int64 {
	return getResource(doc, "#stockBarFreeCrop")
}

func getResource(doc *goquery.Document, selector string) int64 {
	sel := doc.Find(selector)
	if sel.Length() == 0 {
		return -1
	}
	return ParseLong(sel.Text())
}

// GetWarehouseCapacity parses the warehouse capacity.
func GetWarehouseCapacity(doc *goquery.Document) int64 {
	val := doc.Find("#stockBar .warehouse .capacity .value")
	if val.Length() == 0 {
		return -1
	}
	return ParseLong(val.Text())
}

// GetGranaryCapacity parses the granary capacity.
func GetGranaryCapacity(doc *goquery.Document) int64 {
	val := doc.Find("#stockBar .granary .capacity .value")
	if val.Length() == 0 {
		return -1
	}
	return ParseLong(val.Text())
}
