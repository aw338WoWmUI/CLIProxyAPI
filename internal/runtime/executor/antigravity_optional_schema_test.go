package executor

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestAntigravityToolOptionalAnnotationsPreserveHistory(t *testing.T) {
	const payload = `{"request":{"contents":[{"role":"model","parts":[{"functionCall":{"name":"lookup","args":{"optional":true,"task":"keep"}}}]}],"tools":[{"functionDeclarations":[{"name":"lookup","parametersJsonSchema":{"type":"object","properties":{"optional":{"type":"boolean"},"task":{"type":"string","optional":true}},"required":["optional"]}}]}]}}`
	for _, claude := range []bool{false, true} {
		got := sanitizeAntigravityRequestSchemasForModel(payload, claude, claude)
		schema := gjson.Get(got, "request.tools.0.functionDeclarations.0.parameters")
		if schema.Get("properties.task.optional").Exists() {
			t.Errorf("claude=%v: optional annotation reached upstream: %s", claude, schema.Raw)
		}
		if schema.Get("properties.optional.type").String() != "boolean" || schema.Get("required").Raw != `["optional"]` {
			t.Errorf("claude=%v: literal optional property changed: %s", claude, schema.Raw)
		}
		if gjson.Get(got, "request.contents").Raw != gjson.Get(payload, "request.contents").Raw {
			t.Errorf("claude=%v: replayed tool arguments changed", claude)
		}
	}
}
