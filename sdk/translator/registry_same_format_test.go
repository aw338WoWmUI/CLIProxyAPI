package translator

import (
	"context"
	"testing"
)

// A same-format response has nothing to translate. Asking plugin translators
// anyway costs one full request-body round trip per stream chunk, which
// throttles long-context streams.
func TestTranslateStream_SameFormatSkipsPluginTranslator(t *testing.T) {
	r := NewRegistry()
	hooks := &fakePluginHooks{responseTranslateBody: []byte(`data: {"plugin":true}`), responseTranslateOK: true}
	r.SetPluginHooks(hooks)

	same := Format("openai-response")
	got := r.TranslateStream(context.Background(), same, same, "model", nil, nil, []byte(`data: {"raw":true}`), nil)
	if len(got) != 1 || string(got[0]) != `data: {"raw":true}` {
		t.Fatalf("same-format chunk was not forwarded unchanged, got %q", got)
	}
	if hasCall(hooks.calls, "translate-response") {
		t.Fatal("plugin response translator was called for a same-format stream chunk")
	}
	if !hasCall(hooks.calls, "normalize-response-before") || !hasCall(hooks.calls, "normalize-response-after") {
		t.Fatal("same-format stream chunk must still pass through the response normalizers")
	}
}

func TestTranslateNonStream_SameFormatSkipsPluginTranslator(t *testing.T) {
	r := NewRegistry()
	hooks := &fakePluginHooks{responseTranslateBody: []byte(`{"plugin":true}`), responseTranslateOK: true}
	r.SetPluginHooks(hooks)

	same := Format("openai-response")
	got := r.TranslateNonStream(context.Background(), same, same, "model", nil, nil, []byte(`{"raw":true}`), nil)
	if string(got) != `{"raw":true}` {
		t.Fatalf("same-format body was not forwarded unchanged, got %q", got)
	}
	if hasCall(hooks.calls, "translate-response") {
		t.Fatal("plugin response translator was called for a same-format response")
	}
}
