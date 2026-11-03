# Changelog

## v0.1.38 - 2026-11-03

- The core event cursor moved from the `events.cursor` file to Postgres (schema `xb_social` of the shared `xbox` database). `XB_SOCIAL_DATABASE_URL` is required; `XB_STATE_DIR` is gone.
- Invitation answers waiting for the inviter's next poll are now stored in Postgres too, so a restart no longer drops them.
- Tests run against Postgres (`XB_TEST_DATABASE_URL`) and cover a restart on the same database.
