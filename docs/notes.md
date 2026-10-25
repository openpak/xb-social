# Notes: xb-social

Native Xbox friend list, presence and invitations for OpenPak. The
canonical friend graph, blocks and live sessions stay in account core (Social,
Sessions, Links over gRPC). Nothing is copied or cached here.

Not implemented: avatars, rich presence strings, nicknames, multi-user Alive
(needs several bound XUIDs per stream), list-full checks for the other side.

Invitation answers live in this process only (core has no answer path).

Provenance and licence: AGPL-3.0-only, see `LICENSE` and `NOTICE`. Derived from the Xombie project in part; the legacy Rust workspace remains in https://github.com/openpak/xbox-live.
