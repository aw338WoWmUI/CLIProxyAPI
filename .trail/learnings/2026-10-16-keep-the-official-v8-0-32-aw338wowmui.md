---
title: Keep the official v8.0.32 runtime and plugin boundary
author: Aw338WoWmUI
date: 2026-10-16
branch: chore/upgrade-v8.0.32
touches:
  - README.md
  - README_CN.md
  - assets/kedaya.jpg
  - config.example.yaml
  - internal/config/claude_code_test.go
  - internal/config/config_load.go
  - internal/config/config_normalization.go
  - internal/config/config_types.go
  - internal/config/config_yaml.go
  - internal/config/parse.go
---

## Decided
Merge the exact v8.0.32 tag and deploy eceasy/cli-proxy-api:v8.0.32. Runtime source remains identical to upstream; retain the fork's three independent regression tests. Preserve credential priorities, session affinity, model aliases, API-key groups and plugins in the existing mounted configuration.

## Tried and ruled out
Do not deploy a no-plugin release or rebuild the core with old private token patches. The trail check keyword matches in upstream README, Claude and Devin tests do not change local Codex catalogs, discovery configuration or account routing; accepting this release does not reverse those previous decisions.

## Watch out
Back up the database and runtime before changing images. OAuth tokens may rotate after the backup; rollback images/configuration without replacing current credential JSON with an older snapshot.
