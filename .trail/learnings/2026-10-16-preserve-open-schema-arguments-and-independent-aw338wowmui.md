---
title: Preserve open schema arguments and independent cooldowns on 8.0.32
author: Aw338WoWmUI
date: 2026-10-16
branch: chore/upgrade-v8.0.32
---

## Decided
Three reviewed v8.0.32 regressions need small core fixes: omitted additionalProperties permits dynamic keys; a reminder-only string must not produce an empty text block; restoring Devin account quota must not release an unrelated legacy model quota. Each regression fails on the release code and passes after the change. Keep parameterless translator fixtures explicitly closed and open-schema tests preserving all client arguments.

## Tried and ruled out
The official v8.0.32 runtime alone reproduces these bugs, so this upgrade needs a plugin-capable image from the reviewed fork revision. The old decision to use the official image resumes only after upstream contains these fixes.

## Watch out
Changing the dynamic-schema predicate affects nested and composed schemas and both Gemini/Antigravity response translators. Run their suites together with util, executor and auth tests. Retain current OAuth JSON when rolling back an image.
