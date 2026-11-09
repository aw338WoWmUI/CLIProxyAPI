package helps

import (
	"testing"

	"github.com/MoeForge/moego/ojson"
)

func TestNormalizeCodexSIWCAgentMessagesDoc(t *testing.T) {
	doc := ojson.NewDoc([]byte(`{"input":[{"type":"agent_message","id":"amsg_1","author":"/root/reviewer","recipient":"/root","content":[{"type":"input_text","text":"keep finding"}]},{"type":"agent_message","id":"msg_2","content":[]},{"type":"message","id":"existing","role":"user","content":"hi"},{"type":"agent_message","content":[]}]}`))
	NormalizeCodexSIWCAgentMessagesDoc(doc)
	if doc.Get("input.0.id").String() != "msg_amsg_1" || doc.Get("input.0.type").String() != "message" || doc.Get("input.0.content.0.text").String() != "keep finding" || doc.Get("input.0.author").Exists() {
		t.Fatalf("invalid converted message: %s", doc.Bytes())
	}
	if doc.Get("input.1.id").String() != "msg_2" || doc.Get("input.2.id").String() != "existing" || doc.Get("input.3.id").Exists() {
		t.Fatalf("unrelated or missing IDs changed: %s", doc.Bytes())
	}
	before := string(doc.Bytes())
	NormalizeCodexSIWCAgentMessagesDoc(doc)
	if string(doc.Bytes()) != before {
		t.Fatal("normalization is not idempotent")
	}
	NormalizeCodexSIWCAgentMessagesDoc(nil)
}
