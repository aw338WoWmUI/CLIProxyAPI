---
title: Keep CONNECT buffers and await hijacked test handlers
author: Aw338WoWmUI
date: 2026-11-09
branch: chore/upgrade-v8.0.61
touches:
  - internal/api/server_test.go
  - internal/runtime/executor/codex_executor_auth_test.go
---

## Decided
Keep upstream v8.0.61 runtime behavior and stabilize two test fixtures. The mock CONNECT proxy must flush its response and relay the hijacked buffered reader. The reauthorization WebSocket test must wait for its hijacked handler before TempDir cleanup.

## Tried and ruled out
The OAuth mock stress failed three times in 20 iterations before the buffered relay fix and passed 20 after. The WebSocket test failed cleanup eight times in 20 iterations before waiting for handlers and passed 20 after. No production transport retry or OAuth refresh policy was changed. Trail keyword conflicts in upstream docs/examples do not reintroduce rejected local catalog, routing or storage configuration.

## Watch out
Use GODEBUG=http2client=0 only for dependency downloads, never for the test suite: HTTP/2 assertions depend on normal negotiation. Preserve payload finalization ordering and the deployed Codex image/T3 fixes when merging upstream.
