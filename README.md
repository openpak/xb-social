# Xbox Social Server

OpenPak's Xbox Social Server.

| | |
| --- | --- |
| Network | Original Xbox |
| Role | Native Xbox friend list, presence and invitations |
| Depends on | [`xb-account`](https://github.com/openpak/xb-account), [account core](https://github.com/openpak/account) (gRPC) |
| Used by | [`xb-gateway`](https://github.com/openpak/xb-gateway) |
| Image | `ghcr.io/openpak/xb-social` |
| Based on | [Xombie](https://github.com/XombieOnline/xombie) (AGPL-3.0), in part |
| Licence | AGPL-3.0-only |

## What it serves

| Endpoint | What |
| --- | --- |
| `GET /internal/v1/friends?viewer=<xuid>` | The viewer's accepted friends (mTLS, gateway only) |
| `xbfeature` stream, service id 1 | Presence, friends and invitations for the gateway (mTLS, gateway only) |

## Run

```sh
GOWORK=off go build ./... && go vet ./... && go test -race ./...
```

Build and test. Needs `GOPRIVATE=github.com/openpak/*` and `gh auth setup-git`;
the image, configuration and release are in [Deploy](docs/deploy.md).

## Docs

- [Protocol](docs/protocol.md)
- [Evidence](docs/evidence.md)
- [Deploy](docs/deploy.md)
- [Notes](docs/notes.md)

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
