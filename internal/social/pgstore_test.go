// SPDX-License-Identifier: AGPL-3.0-only

package social

import (
	"context"
	"errors"
	"testing"

	"github.com/openpak/xb-social/internal/pgdb"
)

// The cursor and the queued answers survive a restart (a new PG on the same DB).
func TestPGStateSurvivesRestart(t *testing.T) {
	dsn := pgdb.TestDSN(t)
	ctx := context.Background()
	p1, err := OpenPG(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p1.Load(); !errors.Is(err, ErrNoCursor) {
		t.Fatalf("fresh cursor: %v", err)
	}
	if err := p1.Save(4310); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxAnswers+3; i++ {
		if err := p1.AddAnswer(ctx, "alice", Answer{XUID: "x", Payload: []byte{byte(i)}, Accepted: i%2 == 0, At: p1.now().Unix()}); err != nil {
			t.Fatal(err)
		}
	}
	p1.AddAnswer(ctx, "bob", Answer{XUID: "y", At: p1.now().Unix()})
	p1.Close()

	p2, err := OpenPG(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer p2.Close()
	if v, err := p2.Load(); err != nil || v != 4310 {
		t.Fatalf("cursor after restart: %d %v", v, err)
	}
	got, err := p2.TakeAnswers(ctx, "alice")
	if err != nil || len(got) != maxAnswers || got[0].Payload[0] != 3 || got[len(got)-1].Payload[0] != byte(maxAnswers+2) {
		t.Fatalf("answers after restart: n=%d %v", len(got), err)
	}
	if again, _ := p2.TakeAnswers(ctx, "alice"); len(again) != 0 {
		t.Fatalf("answers not cleared: %d", len(again))
	}
	if b, _ := p2.TakeAnswers(ctx, "bob"); len(b) != 1 || b[0].XUID != "y" {
		t.Fatalf("bob: %+v", b)
	}
}
