package helps

import (
	"context"
	"fmt"
	"strings"

	"github.com/MoeForge/moego/ojson"
	multiagentv2 "github.com/router-for-me/CLIProxyAPI/v8/internal/client/codex/optimize-multi-agent-v2"
)

// NormalizeCodexSIWCAgentMessagesDoc preserves agent messages as SIWC-compatible input.
func NormalizeCodexSIWCAgentMessagesDoc(doc *ojson.Doc) {
	if doc == nil {
		return
	}
	for index, item := range doc.Get("input").Array() {
		id := item.Get("id").String()
		if item.Get("type").String() == "agent_message" && id != "" && !strings.HasPrefix(id, "msg") {
			// WebSocket input validation requires message IDs to start with msg.
			_ = doc.Set(fmt.Sprintf("input.%d.id", index), "msg_"+id)
		}
	}
	multiagentv2.RewriteCodexMultiAgentV2InputDoc(doc, context.Background(), nil, nil, true)
}
