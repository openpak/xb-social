# Deploy: xb-social

Config: `XB_FEATURE_ADDR`, `XB_STATE_DIR`, `XB_PRESENCE_TTL_SECS`, `XB_PRESENCE_OFFLINE_ON_CLOSE`, `XB_LISTEN_ADDR`, `XB_TLS_CERT_FILE`, `XB_TLS_KEY_FILE`,
`XB_CLIENT_CA_FILE`, `XB_ACCOUNT_URL`, `XB_ACCOUNT_CA_FILE`, `XB_CORE_ADDR`,
`XB_CORE_KEY_FILE`, and `XB_CORE_CA_FILE`/`XB_CORE_SERVER_NAME` or
`XB_CORE_PLAINTEXT_INTERNAL=1`.

## Repository

- Module `github.com/openpak/xb-social`, image `ghcr.io/openpak/xb-social` (static, scratch, uid 10001).
- Shared packages: `github.com/openpak/xb-common` (private; see its README for access). Service map: workspace docs `xbox/native-service-architecture.md`.
- Build and test: `GOWORK=off go build ./... && go vet ./... && go test -race ./...` (with `GOPRIVATE=github.com/openpak/*` and `gh auth setup-git`).
- Image: `podman build --secret id=netrc,src=$HOME/.netrc -t ghcr.io/openpak/xb-social:dev .`
- Release: push a `vX.Y.Z` tag; `.github/workflows/release.yml` (tags and manual dispatch only) tests and publishes the multi-arch image. CI reads xb-common with the job's `GITHUB_TOKEN` through a netrc build secret; nothing secret is in git or in the image.
- `deploy/` holds the env example and a Podman quadlet example (digests in it are examples; production pins its own).
