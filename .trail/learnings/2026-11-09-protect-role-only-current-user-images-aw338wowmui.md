---
title: Protect role-only current user images in Codex compaction
author: Aw338WoWmUI
date: 2026-11-09
branch: chore/upgrade-v8.0.61
touches:
  - internal/runtime/executor/helps/codex_inline_images.go
  - internal/runtime/executor/helps/codex_inline_images_test.go
---

## Decided
Recognize both explicit message items and role-only user messages when protecting current Codex images. Responses message.type is optional; all images in the latest user message must remain.

## Tried and ruled out
The new three-image role-only regression failed with one image removed before the fix. The one-condition fix passes while existing old-history compaction coverage remains enabled. Do not disable all compaction or change account routing to address this boundary.

## Watch out
This appliance upgrade imports upstream shared handlers unchanged, including plugin-store refresh behavior. It does not introduce a separate fork feature on deprecated v0 routes; retain upstream compatibility rather than maintain a version-specific handler fork.
