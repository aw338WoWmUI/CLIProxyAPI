package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/MoeForge/moego/ojson"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor/helps"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
)

func TestCacheHelperDocCompactsInlineImagesBeforePayloadRules(t *testing.T) {
	cfg := &config.Config{
		Payload: config.PayloadConfig{
			Override: []config.PayloadRule{
				{
					Models: []config.PayloadModelRule{{Name: "*", Protocol: "codex"}},
					Params: map[string]any{
						"custom_override":        "barrier_won",
						"input.0.content.0.text": "rule_won",
					},
				},
			},
		},
	}
	exec := NewCodexExecutor(cfg)
	image := func(tag string) string {
		return testExecutorImagePart("data:image/jpeg;base64," + tag + strings.Repeat("B", helpsInlineImageMin()))
	}
	raw := []byte(`{"model":"gpt-6","input":[` +
		`{"type":"message","role":"user","content":[` + image("OLD") + `]},` +
		`{"type":"message","role":"user","content":[` + image("MID") + `]},` +
		`{"type":"message","role":"user","content":[` + image("RECENT") + `]},` +
		`{"type":"message","role":"user","content":[` + image("NOW") + `]}` +
		`]}`)
	req := cliproxyexecutor.Request{Model: "gpt-6", Payload: raw}
	opts := cliproxyexecutor.Options{SourceFormat: sdktranslator.FormatOpenAIResponse}
	ctx := helps.WithPayloadFinalizer(context.Background(), helps.NewPayloadFinalizer(cfg, exec.Identifier(), "gpt-6", "codex", "", raw, req, opts))

	doc := ojson.NewDoc(append([]byte(nil), raw...))
	_, out, errCache := exec.cacheHelperDoc(ctx, opts.SourceFormat, "https://chatgpt.com/backend-api/codex/responses", req, doc, nil, opts.Headers)
	if errCache != nil {
		t.Fatalf("cacheHelperDoc: %v", errCache)
	}
	if got := ojson.GetBytesNoCopy(out, "custom_override").String(); got != "barrier_won" {
		t.Fatalf("custom_override = %q", got)
	}
	if got := ojson.GetBytesNoCopy(out, "input.0.content.0.text").String(); got != "rule_won" {
		t.Fatalf("payload rule lost to compaction, text=%q type=%s", got, ojson.GetBytesNoCopy(out, "input.0.content.0.type").String())
	}
	if !strings.Contains(ojson.GetBytesNoCopy(out, "input.1.content.0.image_url").String(), "MID") {
		t.Fatalf("recent historical image was omitted")
	}
	if !strings.Contains(ojson.GetBytesNoCopy(out, "input.3.content.0.image_url").String(), "NOW") {
		t.Fatalf("current image was omitted")
	}
}

func testExecutorImagePart(url string) string {
	doc := ojson.NewDoc([]byte(`{"type":"input_image","image_url":""}`))
	if errSet := doc.Set("image_url", url); errSet != nil {
		panic(errSet)
	}
	return string(doc.Bytes())
}

func helpsInlineImageMin() int {
	return 8192
}
