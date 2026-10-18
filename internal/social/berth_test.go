// SPDX-License-Identifier: AGPL-3.0-only
package social

import (
	"context"
	"errors"
	"github.com/openpak/xb-social/internal/core"
	"testing"
	"time"
)

type clientCapture struct {
	core.Backend
	clients   []string
	loseLease bool
}

func (c *clientCapture) RegisterSession(ctx context.Context, a, t, r, client string, d time.Duration) (string, time.Duration, error) {
	c.clients = append(c.clients, client)
	return c.Backend.RegisterSession(ctx, a, t, r, client, d)
}
func (c *clientCapture) Heartbeat(ctx context.Context, id string) (time.Duration, error) {
	if c.loseLease {
		c.loseLease = false
		return 0, core.ErrNotFound
	}
	return c.Backend.Heartbeat(ctx, id)
}
func TestBerthLabelSurvivesRenewalAndDeviceChanges(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	capture := &clientCapture{Backend: e.c}
	e.p.Core = capture
	e.must(e.p.PublishClient(ctx, "alice", xA, "", "berth", 1, 0))
	e.must(e.p.PublishClient(ctx, "alice", xA, "", "berth", 1, 0))
	if len(capture.clients) != 1 {
		t.Fatal("same-client heartbeat replaced the session")
	}
	capture.loseLease = true
	e.p.Tick(ctx)
	if len(capture.clients) != 2 || capture.clients[1] != "berth" {
		t.Fatal("lease recovery lost the Berth label")
	}
	// A newer physical-console session must clear the emulator label.
	e.must(e.p.Publish(ctx, "alice", xA, "", 2, 0))
	if len(capture.clients) != 3 || capture.clients[2] != "" {
		t.Fatal("physical console inherited Berth label")
	}
	if err := e.p.PublishClient(ctx, "alice", xA, "", "berth", 1, 0); !errors.Is(err, ErrSuperseded) {
		t.Fatal("old emulator session took over")
	}
	e.must(e.p.PublishClient(ctx, "alice", xA, "4d530004", "berth", 3, 0))
	capture.loseLease = true
	e.must(e.p.PublishClient(ctx, "alice", xA, "4d530004", "berth", 3, 0))
	if capture.clients[len(capture.clients)-1] != "berth" {
		t.Fatal("publish recovery lost client")
	}
	e.must(e.p.PublishClient(ctx, "alice", xA, "4d530004", "unknown", 4, 0))
	if capture.clients[len(capture.clients)-1] != "" {
		t.Fatal("unknown client accepted")
	}
}
