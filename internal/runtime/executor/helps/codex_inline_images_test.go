package helps

import (
	"strconv"
	"strings"
	"testing"

	"github.com/MoeForge/moego/ojson"
)

func largeTestDataURL(tag string) string {
	return "data:image/jpeg;base64," + tag + strings.Repeat("A", codexInlineImageOmitMinBytes)
}

func testImagePart(url string) string {
	doc := ojson.NewDoc([]byte(`{"type":"input_image","image_url":""}`))
	if errSet := doc.Set("image_url", url); errSet != nil {
		panic(errSet)
	}
	return string(doc.Bytes())
}

func testUserMessage(parts ...string) string {
	return `{"type":"message","role":"user","content":[` + strings.Join(parts, ",") + `]}`
}

func TestCompactCodexInlineImagesKeepsCurrentTurnAndTwoEarlier(t *testing.T) {
	small := `{"type":"input_image","image_url":"data:image/png;base64,AAAA"}`
	https := `{"type":"input_image","image_url":"https://example.com/keep.png"}`
	text := `{"type":"input_text","text":"look at this"}`
	objectURL := testImagePart("")
	objectDoc := ojson.NewDoc([]byte(objectURL))
	if errSet := objectDoc.SetRaw("image_url", `{"url":"`+largeTestDataURL("OBJ")+`"}`); errSet != nil {
		t.Fatal(errSet)
	}

	raw := []byte(`{"model":"gpt-6","input":[` +
		testUserMessage(text, testImagePart(largeTestDataURL("OLD0")), small, https) + `,` +
		testUserMessage(string(objectDoc.Bytes())) + `,` +
		testImagePart(largeTestDataURL("OLD2")) + `,` +
		testUserMessage(testImagePart(largeTestDataURL("OLD3"))) + `,` +
		testUserMessage(testImagePart(largeTestDataURL("NEW1")), testImagePart(largeTestDataURL("NEW2"))) +
		`]}`)
	doc := ojson.NewDoc(raw)
	if got := CompactCodexInlineImagesDoc(doc); got != 2 {
		t.Fatalf("omitted %d, want 2", got)
	}

	if doc.Get("input.0.content.0.text").String() != "look at this" {
		t.Fatalf("sibling text changed: %s", doc.Get("input.0.content.0").Raw)
	}
	if doc.Get("input.0.content.1.type").String() != "input_text" || doc.Get("input.0.content.1.text").String() != codexOmittedInlineImageText {
		t.Fatalf("oldest image was not replaced: %s", doc.Get("input.0.content.1").Raw)
	}
	if doc.Get("input.0.content.2.image_url").String() != "data:image/png;base64,AAAA" {
		t.Fatalf("small data URL changed")
	}
	if doc.Get("input.0.content.3.image_url").String() != "https://example.com/keep.png" {
		t.Fatalf("https URL changed")
	}
	if doc.Get("input.1.content.0.type").String() != "input_text" {
		t.Fatalf("object-form image was not replaced: %s", doc.Get("input.1.content.0.type").String())
	}
	if !strings.Contains(doc.Get("input.2.image_url").String(), "OLD2") {
		t.Fatalf("third-oldest top-level image should be one of the two kept historical images")
	}
	if !strings.Contains(doc.Get("input.3.content.0.image_url").String(), "OLD3") {
		t.Fatalf("second kept historical image missing")
	}
	if !strings.Contains(doc.Get("input.4.content.0.image_url").String(), "NEW1") || !strings.Contains(doc.Get("input.4.content.1.image_url").String(), "NEW2") {
		t.Fatalf("current user turn images were omitted")
	}
}

func TestCompactCodexInlineImagesSkipsImageGeneration(t *testing.T) {
	raw := []byte(`{"tool_choice":{"type":"image_generation"},"input":[` +
		testUserMessage(testImagePart(largeTestDataURL("A")), testImagePart(largeTestDataURL("B")), testImagePart(largeTestDataURL("C")), testImagePart(largeTestDataURL("D"))) +
		`]}`)
	doc := ojson.NewDoc(append([]byte(nil), raw...))
	if got := CompactCodexInlineImagesDoc(doc); got != 0 {
		t.Fatalf("image generation omitted %d images", got)
	}
	if !strings.Contains(doc.Get("input.0.content.0.image_url").String(), "A") || !strings.Contains(doc.Get("input.0.content.3.image_url").String(), "D") {
		t.Fatalf("image generation payload changed")
	}
}

func TestCompactCodexInlineImagesLeavesCurrentTurnIntact(t *testing.T) {
	raw := []byte(`{"input":[` + testUserMessage(
		testImagePart(largeTestDataURL("A")),
		testImagePart(largeTestDataURL("B")),
		testImagePart(largeTestDataURL("C")),
	) + `]}`)
	doc := ojson.NewDoc(raw)
	if got := CompactCodexInlineImagesDoc(doc); got != 0 {
		t.Fatalf("current turn omitted %d images", got)
	}
	for index, tag := range []string{"A", "B", "C"} {
		if !strings.Contains(doc.Get("input.0.content."+strconv.Itoa(index)+".image_url").String(), tag) {
			t.Fatalf("image %s missing", tag)
		}
	}
}

func TestCompactCodexInlineImagesLeavesSmallPayloadUntouched(t *testing.T) {
	raw := []byte(`{"model":"gpt-5","input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"hi"},{"type":"input_image","image_url":"https://example.com/a.png"},{"type":"input_image","image_url":"data:image/png;base64,AAAA"}]}]}`)
	doc := ojson.NewDoc(append([]byte(nil), raw...))
	if got := CompactCodexInlineImagesDoc(doc); got != 0 {
		t.Fatalf("omitted %d", got)
	}
	if string(doc.Bytes()) != string(raw) {
		t.Fatalf("payload changed:\n%s", doc.Bytes())
	}
}

func TestCompactCodexInlineImagesNilAndEmpty(t *testing.T) {
	if got := CompactCodexInlineImagesDoc(nil); got != 0 {
		t.Fatalf("nil omitted %d", got)
	}
	doc := ojson.NewDoc([]byte(`{"model":"gpt-5"}`))
	if got := CompactCodexInlineImagesDoc(doc); got != 0 {
		t.Fatalf("empty input omitted %d", got)
	}
}

func TestCompactCodexInlineImagesKeepsRoleOnlyCurrentTurn(t *testing.T) {
	raw := []byte(`{"input":[{"role":"user","content":[` + strings.Join([]string{
		testImagePart(largeTestDataURL("FIRST")),
		testImagePart(largeTestDataURL("SECOND")),
		testImagePart(largeTestDataURL("THIRD")),
	}, ",") + `]}]}`)
	doc := ojson.NewDoc(append([]byte(nil), raw...))
	if got := CompactCodexInlineImagesDoc(doc); got != 0 || string(doc.Bytes()) != string(raw) {
		t.Fatalf("role-only current message changed; omitted=%d", got)
	}
}
