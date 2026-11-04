# Changelog

## v0.2.0 - 2026-11-04

- Account erasure (card #120). `DELETE /internal/accounts/{account_id}/data` (bearer `ACCOUNT_ERASE_KEY`, listener `ERASE_ADDR`, default `:9099`, container network only, not served while the key is unset) deletes the account's queued invitation answers and every answer the account gave that still waits in another inviter's queue (matched by its Xbox id from the core link), and ends its live presence session. The `account_deleted` core event runs the same erase with the account-keyed rows only (the link is gone by then).
- Kept: friends, blocks and invitations are account-core data and are erased there. The event cursor holds no account data.

## v0.1.38 - 2026-11-03

- The core event cursor moved from the `events.cursor` file to Postgres (schema `xb_social` of the shared `xbox` database). `XB_SOCIAL_DATABASE_URL` is required; `XB_STATE_DIR` is gone.
- Invitation answers waiting for the inviter's next poll are now stored in Postgres too, so a restart no longer drops them.
- Tests run against Postgres (`XB_TEST_DATABASE_URL`) and cover a restart on the same database.
