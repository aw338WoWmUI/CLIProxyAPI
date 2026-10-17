package util

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestCleanJSONSchemaDropsOptionalAnnotations(t *testing.T) {
	const input = `{"type":"object","properties":{"optional":{"type":"boolean","optional":false},"task":{"type":"string","optional":true},"records":{"type":"array","optional":true,"items":{"type":"object","properties":{"count":{"type":"number","optional":true},"optional":{"type":"string"}},"required":["optional"]}}},"required":["optional"]}`
	for name, clean := range map[string]func(string) string{
		"antigravity":         CleanJSONSchemaForAntigravity,
		"antigravityClaude":   func(s string) string { return CleanJSONSchemaForAntigravityClaudeTool(s, false) },
		"antigravityResponse": CleanJSONSchemaForAntigravityResponse,
		"gemini":              CleanJSONSchemaForGemini,
		"geminiJSONSchema":    CleanJSONSchemaForGeminiJSONSchema,
	} {
		t.Run(name, func(t *testing.T) {
			got := gjson.Parse(clean(input))
			for _, path := range []string{"properties.optional.optional", "properties.task.optional", "properties.records.optional", "properties.records.items.properties.count.optional"} {
				if got.Get(path).Exists() {
					t.Errorf("optional annotation survived at %s: %s", path, got.Raw)
				}
			}
			if got.Get("properties.optional.type").String() != "boolean" || got.Get("properties.records.items.properties.optional.type").String() != "string" {
				t.Fatalf("property named optional was altered: %s", got.Raw)
			}
			for _, path := range []string{"required", "properties.records.items.required"} {
				if got.Get(path).Raw != `["optional"]` {
					t.Errorf("required properties changed at %s: %s", path, got.Raw)
				}
			}
		})
	}
}
