package helps

import (
	"strconv"
	"strings"

	"github.com/MoeForge/moego/ojson"
	log "github.com/sirupsen/logrus"
)

const (
	// codexInlineImageKeepRecent is how many large historical data-URL images
	// remain after the current user turn. Codex Desktop resends every prior
	// screenshot inline and without previous_response_id.
	codexInlineImageKeepRecent = 2
	// codexInlineImageOmitMinBytes is the data-URL length above which an
	// earlier image is replaced. HTTPS URLs and small data URLs stay.
	codexInlineImageOmitMinBytes = 8192

	codexOmittedInlineImageText = "[earlier inline image omitted]"
	codexOmittedInlineImagePart = `{"type":"input_text","text":"[earlier inline image omitted]"}`
	codexOmittedInlineImageItem = `{"type":"message","role":"user","content":[{"type":"input_text","text":"[earlier inline image omitted]"}]}`
)

type codexInlineImageHit struct {
	path      string
	size      int
	protected bool
	topLevel  bool
}

// CompactCodexInlineImagesDoc replaces older large data: input_image parts
// with a short text marker. Images in the latest user message stay, and the
// two newest earlier images stay. Image-generation requests are left alone.
// Call this before FinalizePayloadDoc so user payload rules remain last.
func CompactCodexInlineImagesDoc(doc *ojson.Doc) int {
	if doc == nil || doc.Len() == 0 {
		return 0
	}
	if doc.Get("tool_choice.type").String() == "image_generation" {
		return 0
	}
	input := doc.Get("input")
	if !input.IsArray() {
		return 0
	}

	items := input.Array()
	lastUser := -1
	for index, item := range items {
		if item.Get("type").String() == "message" && strings.EqualFold(item.Get("role").String(), "user") {
			lastUser = index
		}
	}

	hits := make([]codexInlineImageHit, 0, 8)
	for index, item := range items {
		itemPath := "input." + strconv.Itoa(index)
		protected := lastUser >= 0 && index == lastUser
		if item.Get("type").String() == "input_image" {
			if size, ok := largeInlineImageSize(item); ok {
				if lastUser >= 0 && index > lastUser {
					protected = true
				}
				hits = append(hits, codexInlineImageHit{path: itemPath, size: size, protected: protected, topLevel: true})
			}
			continue
		}
		collectCodexInlineImageParts(item, itemPath, protected, &hits)
	}
	if len(hits) == 0 {
		return 0
	}

	omitUntil := 0
	historical := 0
	for _, hit := range hits {
		if !hit.protected {
			historical++
		}
	}
	if historical > codexInlineImageKeepRecent {
		omitUntil = historical - codexInlineImageKeepRecent
	}

	omitted := 0
	omittedBytes := 0
	seenHistorical := 0
	for _, hit := range hits {
		if hit.protected {
			continue
		}
		seenHistorical++
		if seenHistorical > omitUntil {
			continue
		}
		replacement := codexOmittedInlineImagePart
		if hit.topLevel {
			replacement = codexOmittedInlineImageItem
		}
		if errSet := doc.SetRaw(hit.path, replacement); errSet != nil {
			log.Errorf("codex: omit inline image at %s: %v", hit.path, errSet)
			continue
		}
		omitted++
		omittedBytes += hit.size
	}
	if omitted > 0 {
		log.Infof("codex: omitted %d earlier inline images (%d bytes)", omitted, omittedBytes)
	}
	return omitted
}

func collectCodexInlineImageParts(item ojson.Result, itemPath string, protected bool, hits *[]codexInlineImageHit) {
	for _, field := range []string{"content", "output"} {
		parts := item.Get(field)
		if !parts.IsArray() {
			continue
		}
		for index, part := range parts.Array() {
			if part.Get("type").String() != "input_image" {
				continue
			}
			size, ok := largeInlineImageSize(part)
			if !ok {
				continue
			}
			path := itemPath + "." + field + "." + strconv.Itoa(index)
			*hits = append(*hits, codexInlineImageHit{path: path, size: size, protected: protected})
		}
	}
}

func largeInlineImageSize(part ojson.Result) (int, bool) {
	field := part.Get("image_url")
	var url string
	switch field.Type {
	case ojson.String:
		url = field.String()
	case ojson.JSON:
		url = field.Get("url").String()
	default:
		return 0, false
	}
	if len(url) <= codexInlineImageOmitMinBytes || len(url) < 5 || !strings.EqualFold(url[:5], "data:") {
		return 0, false
	}
	return len(url), true
}
