// SPDX-License-Identifier: AGPL-3.0-only

package social

import (
	"context"
	"errors"

	"github.com/openpak/xb-social/internal/core"
)

// EraseAccount deletes the account's queued invitation answers and every
// answer it gave that still waits in someone else's queue (xuids are the
// account's Xbox ids). Idempotent; returns the rows removed.
func (p *PG) EraseAccount(ctx context.Context, account string, xuids ...string) (int, error) {
	total := 0
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	r, err := tx.ExecContext(ctx, `DELETE FROM invite_answers WHERE account=?`, account)
	if err != nil {
		return 0, err
	}
	n, _ := r.RowsAffected()
	total += int(n)
	for _, x := range xuids {
		if x == "" {
			continue
		}
		r, err := tx.ExecContext(ctx, `DELETE FROM invite_answers WHERE xuid=?`, x)
		if err != nil {
			return 0, err
		}
		n, _ := r.RowsAffected()
		total += int(n)
	}
	return total, tx.Commit()
}

// Eraser erases an account from xb-social: Postgres answers, and the live
// presence session. Friends, blocks and invitations live in account core and
// are erased there.
type Eraser struct {
	PG       *PG
	Core     core.API // resolves the account's Xbox id while the link still exists
	Presence *Presences
}

// Erase is the full erasure (HTTP endpoint): the core link is still present.
func (e *Eraser) Erase(ctx context.Context, account string) (int, error) {
	var xuids []string
	if e.Core != nil {
		l, err := e.Core.ActiveLink(ctx, account)
		switch {
		case err == nil:
			xuids = append(xuids, l.SubjectID)
		case !errors.Is(err, core.ErrNotFound):
			return 0, err
		}
	}
	return e.erase(ctx, account, xuids)
}

// EraseOwn is the account_deleted event path: the link is gone, only the
// account-keyed rows can be found.
func (e *Eraser) EraseOwn(ctx context.Context, account string) (int, error) {
	return e.erase(ctx, account, nil)
}

func (e *Eraser) erase(ctx context.Context, account string, xuids []string) (int, error) {
	if e.Presence != nil {
		if err := e.Presence.Revoke(ctx, account); err != nil {
			return 0, err
		}
	}
	return e.PG.EraseAccount(ctx, account, xuids...)
}
