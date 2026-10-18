# xb-social

Native Xbox friend list, presence and invitations for OpenPak. The
canonical friend graph, blocks and live sessions stay in account core (Social,
Sessions, Links over gRPC). Nothing is copied or cached here.

`GET /internal/v1/friends?viewer=<xuid>` (mTLS, `spiffe://openpak/xb-gateway`
only; the gateway binds the viewer from its authenticated connection) returns
the viewer's accepted friends. The viewer is resolved through xb-account
(`xbaccountapi.Resolver`), so banned, unlinked or deleted viewers get 403.

Rules:
- A friend is listed only with an active `xbox` link and a published gamertag;
  the XUID is the link's real subject. Friends without one are counted in
  `omitted`, never listed, and no gamertag is derived from a display name.
- Presence comes from core live sessions. A title id is reported only for
  sessions in the `xbox` namespace; a friend playing on another platform is
  `online` without a title. Presence is only asked for friends the viewer may
  discover (core CheckInteraction), so a blocked friend appears offline.
- Any backend failure returns 503, never a partial list.

## Native service 1 (presence, friends, invitations)

With `XB_FEATURE_ADDR` set, xb-social also serves the gateway's `xbfeature`
stream (mTLS, `spiffe://openpak/xb-gateway` only) for service id 1; point the
gateway at it with `XBGW_BACKENDS=1=<host>:<port>` (it replaces the xb-account
svc5 presence stub). Wire layouts: workspace `docs/xbox/social-protocol.md`.

- `xpresence.srf`: Alive/Alive2 (up to 4 messages per POST, each for the
  connection's own user) publish presence to core as an `xbox` Sessions lease
  and answer AliveReply (0x44d) with real buddy/block lists and versions.
  Friend operations 0x3ec-0x3ef, invitations 0x3f3/0x3f4; outcome in `X-Err`.
- `xnotification.srf`: type 8 poll, kinds 0/1 list, 2 presence, 3 invitations,
  4 answers. Version words are always 0.
- Presence: core allows one live session per (account, namespace, title) and a
  Heartbeat always renews to 30 s, so Alive moves a logical deadline
  (`XB_PRESENCE_TTL_SECS`, default 120) and a 10 s keepalive renews the core
  lease until it. A title change registers the new session then expires the old.
  The newest gateway connection generation owns the account (older: 409).
- Events: `social.Poller` follows core's event stream from a durable cursor
  (`XB_STATE_DIR/events.cursor`, first start = stream head), deduplicated by
  event id; ban, deletion and Xbox unlink end presence. `social.PushSource`
  maps notices to the pending-word bits (0x1 list, 0x8 invites, 0x10 answers,
  0x4 presence) for the gateway's UDP 0x5309 push; no channel to the gateway
  exists yet.
- Unsupported (honest 501): other message types, messaging (18), teams (19).
  Replies' unverified details: the notification reply's type field (0x44e).
  Invitation answers live in this process only (core has no answer path).

Not implemented: avatars, rich presence strings, nicknames, multi-user Alive
(needs several bound XUIDs per stream), list-full checks for the other side.

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
- Provenance and licence: AGPL-3.0-only, see `LICENSE` and `NOTICE`. Derived from the Xombie project in part; the legacy Rust workspace remains in https://github.com/openpak/xbox-live.

### Berth presence

The gateway can attach `client=berth` for a provisioned emulator device. Presence
continues to use the Xbox network and friends graph; the separate client label
makes other platforms show “Berth”. The label survives lease renewal and recovery,
and a newer physical Xbox connection clears it. Unknown labels fall back to Xbox.
Device attribution is configured in xb-gateway; see its `deploy/berth-presence.md`.
