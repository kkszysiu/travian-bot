package parser

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// BuildingDTO holds parsed building data from the village layout.
type BuildingDTO struct {
	Location            int
	Level               int
	Type                int // BuildingEnums value (gid)
	IsUnderConstruction bool
}

// QueueBuildingDTO holds parsed queue building data from the build queue.
type QueueBuildingDTO struct {
	Type     string // building type name string
	Level    int
	Duration time.Duration
	Location int
}

// GetFields parses resource field buildings from #resourceFieldContainer.
// Each child with class "level" represents a resource field slot.
func GetFields(doc *goquery.Document) []BuildingDTO {
	var buildings []BuildingDTO
	doc.Find("#resourceFieldContainer .level").Each(func(i int, s *goquery.Selection) {
		classes, exists := s.Attr("class")
		if !exists {
			return
		}

		location := -1
		buildingType := -1
		level := -1
		underConstruction := false

		for _, cls := range strings.Fields(classes) {
			if strings.HasPrefix(cls, "buildingSlot") {
				location = ParseInt(cls[len("buildingSlot"):])
			} else if strings.HasPrefix(cls, "gid") {
				buildingType = ParseInt(cls[len("gid"):])
			} else if strings.HasPrefix(cls, "level") {
				level = ParseInt(cls[len("level"):])
			} else if cls == "underConstruction" {
				underConstruction = true
			}
		}

		buildings = append(buildings, BuildingDTO{
			Location:            location,
			Level:               level,
			Type:                buildingType,
			IsUnderConstruction: underConstruction,
		})
	})
	return buildings
}

// GetInfrastructures parses infrastructure buildings from #villageContent.
// Each div.buildingSlot represents an infrastructure building slot.
func GetInfrastructures(doc *goquery.Document) []BuildingDTO {
	var buildings []BuildingDTO
	nodes := doc.Find("#villageContent div.buildingSlot")

	count := nodes.Length()
	// If 23 nodes found, the last one is a wall duplicate; skip it.
	limit := count
	if count == 23 {
		limit = count - 1
	}

	nodes.Each(func(i int, s *goquery.Selection) {
		if i >= limit {
			return
		}

		aid, _ := s.Attr("data-aid")
		location := ParseInt(aid)

		gid, _ := s.Attr("data-gid")
		buildingType := ParseInt(gid)

		level := 0
		underConstruction := false
		link := s.Find("a[data-level]")
		if link.Length() > 0 {
			lvlStr, _ := link.Attr("data-level")
			level = ParseInt(lvlStr)
			// Check for underConstruction class on the link or parent
			classes, _ := link.Attr("class")
			if strings.Contains(classes, "underConstruction") {
				underConstruction = true
			}
		}
		// Also check the slot itself for underConstruction
		if !underConstruction {
			classes, _ := s.Attr("class")
			if strings.Contains(classes, "underConstruction") {
				underConstruction = true
			}
		}

		// Special cases: location 26 = MainBuilding (15), location 39 = RallyPoint (16)
		if location == 26 {
			buildingType = 15 // MainBuilding
		} else if location == 39 {
			buildingType = 16 // RallyPoint
		}

		buildings = append(buildings, BuildingDTO{
			Location:            location,
			Level:               level,
			Type:                buildingType,
			IsUnderConstruction: underConstruction,
		})
	})
	return buildings
}

// GetQueueBuilding parses the building queue from the finish-now section.
// Returns a list of queued buildings with their type name, level, and duration.
func GetQueueBuilding(doc *goquery.Document) []QueueBuildingDTO {
	var queue []QueueBuildingDTO

	finishNow := doc.Find("div.finishNow")
	if finishNow.Length() == 0 {
		return queue
	}

	parent := finishNow.Parent()
	parent.Find("li").Each(func(i int, s *goquery.Selection) {
		// Type: first text child of div.name
		nameDiv := s.Find("div.name")
		typeName := ""
		if nameDiv.Length() > 0 {
			// Get first text node content (before any child elements)
			nameDiv.Contents().Each(func(j int, c *goquery.Selection) {
				if c.Get(0).Type == html.TextNode {
					text := strings.TrimSpace(c.Text())
					if text != "" && typeName == "" {
						typeName = text
					}
				}
			})
			// Fallback: if no text node found, try the full text trimmed
			if typeName == "" {
				typeName = strings.TrimSpace(nameDiv.Text())
			}
		}

		// Level: from span.lvl
		levelStr := strings.TrimSpace(s.Find("span.lvl").Text())
		level := ParseInt(levelStr)

		// Duration: from .timer value attribute (seconds)
		timerVal, _ := s.Find(".timer").Attr("value")
		seconds := ParseInt(timerVal)
		duration := time.Duration(0)
		if seconds > 0 {
			duration = time.Duration(seconds) * time.Second
		}

		queue = append(queue, QueueBuildingDTO{
			Type:     typeName,
			Level:    level,
			Duration: duration,
			Location: -1,
		})
	})
	return queue
}
