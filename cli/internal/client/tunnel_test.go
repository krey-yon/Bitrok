package client

import (
	"context"
	"testing"
	"time"
)

func TestTunnelSessionRefreshQueuesReconnect(t *testing.T) {
	session := NewTunnelSession("https://relay.example.com", "token", "tunnel-id", "localhost:3000")

	session.Refresh()
	session.Refresh()

	select {
	case <-session.refresh:
	default:
		t.Fatal("refresh did not queue a reconnect request")
	}
	select {
	case <-session.refresh:
		t.Fatal("duplicate refresh requests should be coalesced")
	default:
	}
}

func TestTunnelSessionRefreshInterruptsReconnectBackoff(t *testing.T) {
	session := NewTunnelSession("https://relay.example.com", "token", "tunnel-id", "localhost:3000")
	done := make(chan bool, 1)
	go func() {
		_, refreshed := session.reconnect.sleepContext(context.Background(), session.refresh)
		done <- refreshed
	}()

	session.Refresh()

	select {
	case refreshed := <-done:
		if !refreshed {
			t.Fatal("refresh did not interrupt reconnect backoff")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("refresh did not wake reconnect backoff promptly")
	}
}
