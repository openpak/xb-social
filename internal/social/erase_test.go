// SPDX-License-Identifier: AGPL-3.0-only

package social

import (
	"context"
	"testing"

	"github.com/openpak/xb-social/internal/coretest"
	"github.com/openpak/xb-social/internal/pgdb"
)

func TestEraseAccount(t *testing.T) {
	ctx := context.Background()
	pg, err := OpenPG(pgdb.TestDSN(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pg.Close()
	c := coretest.New()
	c.AddAccount("alice", "Alice")
	c.AddAccount("bob", "Bob")
	c.LinkXbox("alice", xA, "Alice")
	c.LinkXbox("bob", xB, "Bob")
	now := pg.now().Unix()
	pg.AddAnswer(ctx, "alice", Answer{XUID: xB, At: now}) // bob answered alice
	pg.AddAnswer(ctx, "bob", Answer{XUID: xA, At: now})   // alice answered bob
	pg.AddAnswer(ctx, "bob", Answer{XUID: xC, At: now})
	e := &Eraser{PG: pg, Core: c}
	n, err := e.Erase(ctx, "alice")
	if err != nil || n != 2 {
		t.Fatalf("erase: %d %v", n, err)
	}
	if n, err := e.Erase(ctx, "alice"); err != nil || n != 0 {
		t.Fatalf("second erase: %d %v", n, err)
	}
	if b, _ := pg.TakeAnswers(ctx, "bob"); len(b) != 1 || b[0].XUID != xC {
		t.Fatalf("bob's queue: %+v", b)
	}
	// event path: own rows only
	pg.AddAnswer(ctx, "bob", Answer{XUID: xB, At: now})
	pg.AddAnswer(ctx, "alice", Answer{XUID: xC, At: now})
	if n, err := e.EraseOwn(ctx, "alice"); err != nil || n != 1 {
		t.Fatalf("own: %d %v", n, err)
	}
	if b, _ := pg.TakeAnswers(ctx, "bob"); len(b) != 1 {
		t.Fatalf("bob's queue after own erase: %+v", b)
	}
}

func TestPollerErasesOnAccountDeleted(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	cur := &MemCursor{}
	cur.Save(0)
	var got []string
	p := &Poller{Core: e.c, Cursor: cur, Hub: &Hub{}, Presence: e.p,
		Erase: func(_ context.Context, a string) error { got = append(got, a); return nil }}
	e.c.Emit("account_banned", "alice", "", "")
	e.c.Emit("account_deleted", "alice", "", "")
	if _, err := p.Step(ctx); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "alice" {
		t.Fatalf("erased %v", got)
	}
}
