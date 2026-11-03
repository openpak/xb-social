// SPDX-License-Identifier: AGPL-3.0-only

package social

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/openpak/xb-social/internal/pgdb"
)

// PG is xb-social's durable state in Postgres (schema xb_social of the shared
// "xbox" database): the core event cursor and the invitation answers waiting
// for the inviter's next poll. It implements CursorStore and AnswerStore.
type PG struct {
	db  *pgdb.DB
	now func() time.Time
}

const cursorName = "core-events"

// answerTTL drops answers nobody polled for; an invitation lives minutes.
const answerTTL = 24 * time.Hour

// OpenPG connects to the Postgres at dsn (schema xb_social unless the DSN sets a
// search_path) and creates the tables.
func OpenPG(dsn string) (*PG, error) {
	db, err := pgdb.Open(dsn, "xb_social", `
CREATE TABLE IF NOT EXISTS cursors(name TEXT PRIMARY KEY, version BIGINT NOT NULL);
CREATE TABLE IF NOT EXISTS invite_answers(
 id BIGSERIAL PRIMARY KEY, account TEXT NOT NULL, xuid TEXT NOT NULL, payload BYTEA NOT NULL,
 accepted BOOLEAN NOT NULL, at BIGINT NOT NULL);
CREATE INDEX IF NOT EXISTS invite_answers_account ON invite_answers(account, id);`)
	if err != nil {
		return nil, err
	}
	return &PG{db: db, now: time.Now}, nil
}

func (p *PG) Close() error { return p.db.Close() }

// Load returns ErrNoCursor when nothing was ever saved.
func (p *PG) Load() (uint64, error) {
	var v int64
	err := p.db.QueryRow(`SELECT version FROM cursors WHERE name=?`, cursorName).Scan(&v)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNoCursor
		}
		return 0, err
	}
	return uint64(v), nil
}

// Save records the last fully handled event version.
func (p *PG) Save(v uint64) error {
	_, err := p.db.Exec(`INSERT INTO cursors(name,version) VALUES(?,?) ON CONFLICT(name) DO UPDATE SET version=excluded.version`, cursorName, int64(v))
	return err
}

func (p *PG) AddAnswer(ctx context.Context, account string, a Answer) error {
	payload := a.Payload
	if payload == nil {
		payload = []byte{}
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM invite_answers WHERE at<?`, p.now().Add(-answerTTL).Unix()); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO invite_answers(account,xuid,payload,accepted,at) VALUES(?,?,?,?,?)`,
		account, a.XUID, payload, a.Accepted, a.At); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM invite_answers WHERE account=? AND id NOT IN (SELECT id FROM invite_answers WHERE account=? ORDER BY id DESC LIMIT ?)`,
		account, account, maxAnswers); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *PG) TakeAnswers(ctx context.Context, account string) ([]Answer, error) {
	rows, err := p.db.QueryContext(ctx, `DELETE FROM invite_answers WHERE account=? RETURNING id,xuid,payload,accepted,at`, account)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type row struct {
		id int64
		a  Answer
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.a.XUID, &r.a.Payload, &r.a.Accepted, &r.a.At); err != nil {
			return nil, err
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := 1; i < len(got); i++ { // RETURNING order is unspecified: oldest first
		for j := i; j > 0 && got[j].id < got[j-1].id; j-- {
			got[j], got[j-1] = got[j-1], got[j]
		}
	}
	var out []Answer
	for _, r := range got {
		out = append(out, r.a)
	}
	return out, nil
}
