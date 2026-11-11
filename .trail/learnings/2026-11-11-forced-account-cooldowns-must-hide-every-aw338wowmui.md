---
title: Forced account cooldowns must hide every advertised model
author: Aw338WoWmUI
date: 2026-11-11
branch: chore/upgrade-v8.0.67-20261112
id: l-ba81cd0e
scope: repo
status: active
src: agent
touches:
  - sdk/cliproxy/auth/conductor_models.go
---

## Decided
An explicit account-wide force cooldown applies even when a credential already has per-model states. Model-list projections must use the same effective force deadline as the selector and keep a longer per-model deadline if present.

## Tried and ruled out
The ordinary auth-level availability aggregate is intentionally ignored for unmatched models when per-model states exist. Applying every aggregate globally would hide otherwise usable models. Only the explicit force deadline is an independent gate.

## Watch out
A shorter per-model suspension must not cause the registry to relist a model while an account-level force deadline remains active. Disabled and unauthorized states remain open-ended. Test both the pure projection and MarkResult through the registry listing; a projection-only test does not establish publication.

In the code: sdk/cliproxy/auth/conductor_models.go clientModelProjectionForAuth, conductor_cooldown.go effectiveAuthForceUntil, conductor_models_force_projection_test.go. The original implementation failed unmatched-model visibility and shorter-model-deadline cases, including a real registry listing after MarkResult.
