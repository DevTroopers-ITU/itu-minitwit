package main

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

// Setup a test server with a fresh temp database
func setupTestServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()

	// Create a temp database
	tmpFile, err := os.CreateTemp("", "minitwit-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	// Open the temp database
	db, err = openDB(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Create tables
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatal(err)
	}

	store = newStore()

	ts := httptest.NewServer(setupRouter())

	// Client with cookie jar — follows redirects automatically
	jar, _ := cookiejar.New(nil)
	client := ts.Client()
	client.Jar = jar

	return ts, client
}

// Helper: read response body as string
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

// Helper: register a user
func register(t *testing.T, ts *httptest.Server, client *http.Client, username, password, password2, email string) string {
	t.Helper()
	if password2 == "" {
		password2 = password
	}
	if email == "" {
		email = username + "@example.com"
	}
	resp, err := client.PostForm(ts.URL+"/register", url.Values{
		"username":  {username},
		"password":  {password},
		"password2": {password2},
		"email":     {email},
	})
	if err != nil {
		t.Fatal(err)
	}
	return readBody(t, resp)
}

// Helper: login
func login(t *testing.T, ts *httptest.Server, client *http.Client, username, password string) string {
	t.Helper()
	resp, err := client.PostForm(ts.URL+"/login", url.Values{
		"username": {username},
		"password": {password},
	})
	if err != nil {
		t.Fatal(err)
	}
	return readBody(t, resp)
}

// Helper: register and login
func registerAndLogin(t *testing.T, ts *httptest.Server, client *http.Client, username, password string) string {
	t.Helper()
	register(t, ts, client, username, password, "", "")
	return login(t, ts, client, username, password)
}

// Helper: logout
func doLogout(t *testing.T, ts *httptest.Server, client *http.Client) string {
	t.Helper()
	resp, err := client.Get(ts.URL + "/logout")
	if err != nil {
		t.Fatal(err)
	}
	return readBody(t, resp)
}

// Helper: add a message
func addMessage(t *testing.T, ts *httptest.Server, client *http.Client, text string) string {
	t.Helper()
	resp, err := client.PostForm(ts.URL+"/add_message", url.Values{
		"text": {text},
	})
	if err != nil {
		t.Fatal(err)
	}
	return readBody(t, resp)
}

// Helper: GET a page and return body
func getBody(t *testing.T, ts *httptest.Server, client *http.Client, path string) string {
	t.Helper()
	resp, err := client.Get(ts.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	return readBody(t, resp)
}

func TestRegister(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	// Successful registration
	body := register(t, ts, client, "user1", "default", "", "")
	if !strings.Contains(body, "You were successfully registered and can login now") {
		t.Error("Expected successful registration message")
	}

	// Duplicate username
	body = register(t, ts, client, "user1", "default", "", "")
	if !strings.Contains(body, "The username is already taken") {
		t.Error("Expected 'username already taken' message")
	}

	// Empty username
	body = register(t, ts, client, "", "default", "", "test@example.com")
	if !strings.Contains(body, "You have to enter a username") {
		t.Error("Expected 'enter a username' message")
	}

	// Empty password
	body = register(t, ts, client, "meh", "", "", "meh@example.com")
	if !strings.Contains(body, "You have to enter a password") {
		t.Error("Expected 'enter a password' message")
	}

	// Mismatched passwords
	body = register(t, ts, client, "meh", "x", "y", "meh@example.com")
	if !strings.Contains(body, "The two passwords do not match") {
		t.Error("Expected 'passwords do not match' message")
	}

	// Invalid email
	body = register(t, ts, client, "meh", "foo", "", "broken")
	if !strings.Contains(body, "You have to enter a valid email address") {
		t.Error("Expected 'valid email address' message")
	}
}

func TestLoginLogout(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	// Register and login
	body := registerAndLogin(t, ts, client, "user1", "default")
	if !strings.Contains(body, "You were logged in") {
		t.Error("Expected 'logged in' message")
	}

	// Logout
	body = doLogout(t, ts, client)
	if !strings.Contains(body, "You were logged out") {
		t.Error("Expected 'logged out' message")
	}

	// Wrong password
	body = login(t, ts, client, "user1", "wrongpassword")
	if !strings.Contains(body, "Invalid password") {
		t.Error("Expected 'Invalid password' message")
	}

	// Wrong username
	body = login(t, ts, client, "user2", "wrongpassword")
	if !strings.Contains(body, "Invalid username") {
		t.Error("Expected 'Invalid username' message")
	}
}

func TestMessageRecording(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	registerAndLogin(t, ts, client, "foo", "default")
	addMessage(t, ts, client, "test message 1")
	addMessage(t, ts, client, "<test message 2>")

	body := getBody(t, ts, client, "/")
	if !strings.Contains(body, "test message 1") {
		t.Error("Expected 'test message 1' on timeline")
	}
	if !strings.Contains(body, "&lt;test message 2&gt;") {
		t.Error("Expected HTML-escaped message on timeline")
	}
}

func TestTimelines(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	// foo posts a message
	registerAndLogin(t, ts, client, "foo", "default")
	addMessage(t, ts, client, "the message by foo")
	doLogout(t, ts, client)

	// bar posts a message
	registerAndLogin(t, ts, client, "bar", "default")
	addMessage(t, ts, client, "the message by bar")

	// Public timeline shows both
	body := getBody(t, ts, client, "/public")
	if !strings.Contains(body, "the message by foo") {
		t.Error("Expected foo's message on public timeline")
	}
	if !strings.Contains(body, "the message by bar") {
		t.Error("Expected bar's message on public timeline")
	}

	// Bar's timeline shows only bar's message
	body = getBody(t, ts, client, "/")
	if strings.Contains(body, "the message by foo") {
		t.Error("Did not expect foo's message on bar's timeline")
	}
	if !strings.Contains(body, "the message by bar") {
		t.Error("Expected bar's message on bar's timeline")
	}

	// Follow foo
	body = getBody(t, ts, client, "/foo/follow")
	if !strings.Contains(body, "You are now following") {
		t.Error("Expected follow confirmation message")
	}

	// Now bar's timeline should show both
	body = getBody(t, ts, client, "/")
	if !strings.Contains(body, "the message by foo") {
		t.Error("Expected foo's message after following")
	}
	if !strings.Contains(body, "the message by bar") {
		t.Error("Expected bar's message on own timeline")
	}

	// User page shows only that user's messages
	body = getBody(t, ts, client, "/bar")
	if strings.Contains(body, "the message by foo") {
		t.Error("Did not expect foo's message on bar's user page")
	}
	if !strings.Contains(body, "the message by bar") {
		t.Error("Expected bar's message on bar's user page")
	}

	body = getBody(t, ts, client, "/foo")
	if !strings.Contains(body, "the message by foo") {
		t.Error("Expected foo's message on foo's user page")
	}
	if strings.Contains(body, "the message by bar") {
		t.Error("Did not expect bar's message on foo's user page")
	}

	// Unfollow foo
	body = getBody(t, ts, client, "/foo/unfollow")
	if !strings.Contains(body, "You are no longer following") {
		t.Error("Expected unfollow confirmation message")
	}

	// Bar's timeline should only show bar's message again
	body = getBody(t, ts, client, "/")
	if strings.Contains(body, "the message by foo") {
		t.Error("Did not expect foo's message after unfollowing")
	}
	if !strings.Contains(body, "the message by bar") {
		t.Error("Expected bar's message on own timeline")
	}
}
