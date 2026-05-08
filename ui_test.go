package main

import (
	"strings"
	"testing"
)

func TestLoginPageUI(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	body := getBody(t, ts, client, "/login")
	for _, want := range []string{
		"Sign In",
		`form action="/login" method="post"`,
		`name="username"`,
		`name="password"`,
		`value="Sign In"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("login page missing %q", want)
		}
	}
}

func TestRegisterPageUI(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	body := getBody(t, ts, client, "/register")
	for _, want := range []string{
		"Sign Up",
		`form action="/register" method="post"`,
		`name="username"`,
		`name="email"`,
		`name="password"`,
		`name="password2"`,
		`value="Sign Up"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("register page missing %q", want)
		}
	}
}

func TestTimelinePageUI(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	registerAndLogin(t, ts, client, "foo", "default")
	addMessage(t, ts, client, "hello from ui test")

	body := getBody(t, ts, client, "/")
	for _, want := range []string{
		"My Timeline",
		"What's on your mind foo?",
		`form action="/add_message" method=post`,
		`name=text`,
		`value="Share"`,
		`href="/logout"`,
		`href="/public"`,
		`hello from ui test`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("timeline page missing %q", want)
		}
	}
}
