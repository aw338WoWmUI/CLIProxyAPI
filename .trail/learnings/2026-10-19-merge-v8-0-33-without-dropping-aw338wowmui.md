---
title: Merge v8.0.33 without dropping the OMO schema fix
author: Aw338WoWmUI
date: 2026-10-19
branch: chore/upgrade-v8.0.33
touches:
  - README.md
  - README_CN.md
  - README_JA.md
  - config.example.yaml
  - internal/api/handlers/management/auth_files.go
  - internal/api/handlers/management/auth_files_crud.go
  - internal/api/handlers/management/auth_files_excluded_models_test.go
  - internal/api/handlers/management/auth_files_fields.go
  - internal/discovery/discovery_test.go
  - internal/discovery/service.go
---

## Decided
Merge upstream v8.0.33 into the gateway fork. Preserve the optional schema annotation removal and its regression tests together with upstream contentMediaType/contentEncoding removal. Keep plugin-enabled Debian/glibc Docker packaging and existing native plugin mounts.

## Tried and ruled out
Trail check keyword matches in upstream README, exclusion tests and discovery tests are not changes to local client catalogs, endpoint discovery or account routing. The upgrade does not reverse those earlier decisions. Do not remove upstream regression coverage to silence keyword matches.

## Watch out
Keep request-performance history on the existing logs volume across core replacement. Recreate only the core container so the temporary Cloudflare address stays stable. Verify native plugin registration after the host upgrade.
