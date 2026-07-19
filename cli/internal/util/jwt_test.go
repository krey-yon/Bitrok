package util

import "testing"

func TestUsernameFromTokenEmailFallback(t *testing.T) {
	tok := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJJeXpNcnlwdEdvWlR5WlZFZzRkNEpKanNjMXc0QTN1QSIsInR5cGUiOiJjbGkiLCJlbWFpbCI6IjEwMTcwMjkwMkB1c2Vycy5ub3JlcGx5LmdpdGh1Yi5jb20iLCJpYXQiOjE3ODQ0MjIxNTYsImV4cCI6MTc4NzAxNDE1NiwiYXVkIjoiYml0cm9rLWNsaSIsImlzcyI6ImJpdHJvayJ9.3iZdThrFjwzaO5SbHkiMC-TzTw9drOeOG5lTZnHc5xY"
	u, err := UsernameFromToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if u != "101702902" {
		t.Fatalf("got %q", u)
	}
}
