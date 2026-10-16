---
title: Use official v8.0.31 with plugin-owned Copilot quota
author: Aw338WoWmUI
date: 2026-10-16
branch: chore/upgrade-v8.0.31
touches:
  - internal/api/handlers/management/api_tools.go
  - internal/api/handlers/management/api_tools_copilot_test.go
  - internal/api/handlers/management/config_v8.go
  - internal/api/handlers/management/config_v8_etag_test.go
  - internal/api/server_config_v8_etag_test.go
  - internal/api/server_middleware.go
  - internal/pluginhost/host.go
  - internal/pluginhost/loader_unix.go
  - internal/pluginhost/loader_unix_test.go
  - internal/pluginhost/loader_windows.go
---

## Decided
Merge v8.0.31 and remove the fork-only private credential token fallback from deprecated API-call management. Subscription Tools and Dashboard's normalized quota fallback supply that behavior without modifying the host. Runtime source matches the official release; retain the existing independent regression tests.

## Tried and ruled out
Do not reintroduce provider-specific access to private Storage merely to fill the legacy TOKEN placeholder. The installed native quota plugin is already accepted on the unmodified official host.

## Watch out
Use the plugin-enabled Linux asset and verify its release checksum. The no-plugin Linux asset cannot load existing providers. v8 config If-Match is optional, so current field writers remain supported; native shadow-copy loading must be exercised with existing plugin binaries before deployment.
