// SPDX-License-Identifier: AGPL-3.0-only

package svc1

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/openpak/xb-social/internal/social"
)

// pollKind2 sends a kind 2 poll as alice and returns the reply words and presence records.
func pollKind2(t *testing.T, e *env) (words []uint32, recs [][]byte) {
	t.Helper()
	m := msg(msgNotifyPoll, 5, xAlice, uint16(2), uint32(0), uint16(0x14), uint32(0x800))
	code, out := e.doPath(t, "/xpnfront/xnotification.srf", infoFor(xAlice, "alice"), "xon/1", m)
	if code != ok {
		t.Fatalf("poll %q", code)
	}
	n := int(binary.LittleEndian.Uint16(out[0x34:]))
	nw := int(binary.LittleEndian.Uint32(out[0x3a:]))
	p := 0x3e
	for range nw {
		words = append(words, binary.LittleEndian.Uint32(out[p:]))
		p += 4
	}
	for range n {
		l := 6 + int(binary.LittleEndian.Uint16(out[p+4:]))
		recs = append(recs, out[p:p+l])
		p += l
	}
	if p != len(out) {
		t.Fatalf("reply has %d trailing bytes", len(out)-p)
	}
	return words, recs
}

func befriend(t *testing.T, e *env) {
	t.Helper()
	ctx := context.Background()
	a := fmt.Sprintf("%016x", xAlice)
	if _, err := e.srv.Social.SendFriendRequest(ctx, xBob, "", a); err != nil {
		t.Fatal(err)
	}
	if err := e.srv.Social.AcceptFriendRequest(ctx, a, "", xBob); err != nil {
		t.Fatal(err)
	}
}

func TestKind2PollListsSwitchFriendAndClearsPresenceBit(t *testing.T) {
	e := newEnv(t)
	befriend(t, e)
	e.c.AddForeignSession("bob", "switch", "")
	var last []uint32
	e.srv.Notifier = &social.Notifier{Emit: func(_ uint64, w []uint32) { last = w }}
	e.srv.Notifier.Apply(social.Push{XUID: fmt.Sprintf("%016x", xAlice), Bits: social.PendingPresent})
	if len(last) != 2 || last[1] != social.BitPresent || social.BitPresent != 0x4 {
		t.Fatalf("presence push must set bit 0x4 in the bits word: %x", last)
	}
	pushed := last[0] >> 8

	words, recs := pollKind2(t, e)
	if len(recs) != 1 {
		t.Fatalf("records %d", len(recs))
	}
	body := recs[0][6:]
	if binary.LittleEndian.Uint64(body) != 0x0009000000000002 || binary.LittleEndian.Uint32(body[8:]) != 0 ||
		binary.LittleEndian.Uint32(body[0xc:]) != flagOnline || binary.LittleEndian.Uint16(recs[0][4:]) != presRecBody {
		t.Fatalf("switch friend record % x", recs[0])
	}
	if len(words) != 2 || words[1]&social.BitPresent != 0 || words[0]>>8 <= pushed {
		t.Fatalf("poll must clear 0x4 with a higher version: %x (pushed version %d)", words, pushed)
	}
}

func TestKind2PollSendsOfflineFriendWithFlagsZero(t *testing.T) {
	e := newEnv(t)
	befriend(t, e)
	_, recs := pollKind2(t, e) // bob has no session at all
	if len(recs) != 1 {
		t.Fatalf("records %d", len(recs))
	}
	body := recs[0][6:]
	if binary.LittleEndian.Uint64(body) != 0x0009000000000002 || binary.LittleEndian.Uint32(body[8:]) != 0 ||
		binary.LittleEndian.Uint32(body[0xc:]) != 0 {
		t.Fatalf("offline friend record % x", recs[0])
	}
}

func TestHelloIsPushedAfterAliveReplyOnlyWhileOnline(t *testing.T) {
	for _, tc := range []struct {
		name    string
		signOut bool
	}{{"online", false}, {"signed out before the delay", true}} {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(t)
			got := make(chan []uint32, 4)
			e.srv.HelloDelay = 50 * time.Millisecond
			e.srv.Notifier = &social.Notifier{Emit: func(_ uint64, w []uint32) { got <- w }}
			e.do(t, info(xAlice), "xon/1", aliveFor(t, 0, 0))
			select {
			case w := <-got:
				t.Fatalf("hello pushed before the reply was written: %x", w)
			default:
			}
			if tc.signOut {
				if err := e.srv.Presence.Disconnect(context.Background(), "alice", 1); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case w := <-got:
				if tc.signOut {
					t.Fatalf("hello pushed for a signed-out user: %x", w)
				}
				if len(w) != 2 || w[1] != social.BitPresent {
					t.Fatalf("hello words %x", w)
				}
			case <-time.After(300 * time.Millisecond):
				if !tc.signOut {
					t.Fatal("no hello after the reply")
				}
			}
		})
	}
}
