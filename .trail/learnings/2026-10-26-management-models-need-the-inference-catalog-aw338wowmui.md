---
title: Management models need the inference catalog route on the management origin
author: Aw338WoWmUI
date: 2026-10-26
branch: chore/upgrade-v8.0.43
touches:
  - README.md
  - README_CN.md
  - README_JA.md
  - cmd/fetch_devin_models/main.go
  - cmd/fetch_devin_models/main_test.go
  - internal/api/handlers/management/auth_files.go
  - internal/api/handlers/management/auth_files_devin_oauth.go
  - internal/api/handlers/management/auth_files_oauth_callback.go
  - internal/api/handlers/management/auth_files_provider_oauth.go
  - internal/api/handlers/management/auth_files_v8.go
---

## Decided
The native management System Info page fetches GET /v1/models on the same origin that serves management.html. Forward that exact GET path through the management reverse proxy while keeping normal inference authentication. The dashboard catch-all must not own this route. Retain the previous POST /v1/responses isolation on the management origin.

## Tried and ruled out
The core inference port returned 200 while the LAN management and public management origins returned 404. This was an appliance reverse-proxy omission, not a missing v8 management models endpoint or a plugin defect. Adding a new management API or modifying the frontend would address the wrong layer. The real native page refresh displayed 497 models after the route change; unauthenticated catalog requests still returned 401.

## Watch out
Merge v8.0.43's complete implicit-open schema handling and tests instead of preserving the old dynamic-schema predicate patch. Utility, executor, translator and full core suites passed with the upstream implementation. Keep the remaining gateway fixes, credentials, routing, native plugins, history volume and existing temporary tunnel when deploying the upgrade.
