---
title: Antigravity rejects optional annotations in full OMO tool schemas
author: Aw338WoWmUI
date: 2026-10-17
branch: fix/antigravity-optional-schema
touches:
  - internal/util/gemini_schema.go
  - internal/runtime/executor/antigravity_optional_schema_test.go
  - internal/util/gemini_schema_optional_test.go
---

## Decided
Strip the non-standard optional annotation in the shared Gemini/Antigravity schema cleaner. Optionality is already represented by the parent required array. Preserve literal properties named optional and tool-call arguments in replayed conversation history.

## Tried and ruled out
A plain OMO prompt and a read-only tool allowlist both succeeded against Antigravity Opus, while OMO's full tool declarations consistently returned Google INVALID_ARGUMENT for optional fields in function declarations 17 and 18. A single function with optional:true reproduced the same upstream HTTP 400. Unit tests at the shared cleaner and the Antigravity request-schema boundary failed before adding optional to removeUnsupportedKeywords, then passed with the one-key fix.

## Watch out
A minimal text-only health check does not prove compatibility with a coding agent's complete tool set. Exercise a full tool request, and keep required membership, a literal optional property, nested array item schemas, and replayed tool arguments in the regression checks. Clean schema subtrees only, never a whole conversation request.
