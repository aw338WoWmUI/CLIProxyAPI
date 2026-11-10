---
title: Finish translated Claude streams before waiting for a separator
author: Aw338WoWmUI
date: 2026-11-10
branch: chore/upgrade-v8.0.63
id: l-e5c7f649
scope: repo
status: active
src: agent
touches:
  - internal/runtime/executor/antigravity_reasoning_replay_test.go
  - internal/runtime/executor/claude_executor_stream.go
  - internal/runtime/executor/claude_executor_stream_terminal_test.go
  - sdk/auth/manager.go
  - sdk/auth/manager_login_proxy_test.go
---

## Decided
Claude translated streams must recognize a valid message_stop data line before its blank separator. Translation may emit response.completed immediately, so waiting for the separator can keep Chunks open indefinitely. This is independent of connection reuse or draining policy. Also apply an explicit login proxy when an SDK manager returns a credential without a token store.

## Tried and ruled out
Focused tests reproduced both upstream v8.0.63 defects before the fixes. The third review concern about countTokens replay degradation did not reproduce with the supplied mismatched-call fixture: conservative replay rejects the mismatched call. Keep a read-only pipeline/cache preservation test, without changing production replay behavior. Compare normalized cache entries against a before snapshot; comparing raw fixture JSON falsely reported mutation.

## Watch out
In the code: claude_executor_stream.go keeps JSON validity checks and does not restore detach/drain behavior. sdk/auth/manager.go still reapplies the proxy after persisted metadata merging on stored logins. antigravity_reasoning_replay_test.go must use the caller-scoped request key so tests actually reach the seeded cache.
