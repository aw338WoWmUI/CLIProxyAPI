---
title: Plugin streams must keep turn and socket lifetimes separate
author: Aw338WoWmUI
date: 2026-11-10
branch: fix/stream-translation-websocket-reuse
id: l-acb20a65
scope: repo
status: active
src: agent
touches:
  - ".github/workflows/auto-retarget-main-pr-to-dev.yml"
  - sdk/api/handlers/openai/openai_responses_websocket.go
  - sdk/translator/registry.go
---

## Decided
Treat executor-plugin responses as per-turn streams even when their nested selected provider uses WebSocket. Keep downstream replay history and do not pass the socket input reader into nested plugin executions. Native Codex routes retain duplex behavior.

Skip fallback response translators when source and target formats match. Before/after normalizers and registered native transforms still run.

## Tried and ruled out
A live three-turn WebSocket probe completed turn one then disconnected. The handler-level regression through nested model execution reproduced two failures: without steering the second turn required HTTP replay; with steering a finite plugin stream was treated as a terminated duplex socket. Both now pass, including a provider switch with complete replay history.

Fresh short prompts did not diagnose the separate long-context slowdown. Same existing Codex chat before compaction rose from roughly 2 TPS at 430k input to 42–52 TPS at 439k–447k input after identity translation was skipped; rates include reasoning tokens and are not an identical-request benchmark.

## Watch out
Do not disable native duplex support or forward a previous response ID to another provider. Test both steering modes, nested auth callbacks, subsequent turns and replay contents.

In the code: sdk/translator/registry.go; sdk/api/handlers/openai/openai_responses_websocket.go and openai_responses_websocket_plugin_turn_test.go.
