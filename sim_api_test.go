package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

const simAuth = "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh"

// simRequest is a helper that builds a request with simulator auth + JSON content type.
func simRequest(t *testing.T, method, url string, body interface{}) *http.Request {
	t.Helper()
	var r *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		r, _ = http.NewRequest(method, url, bytes.NewReader(b))
	} else {
		r, _ = http.NewRequest(method, url, nil)
	}
	r.Header.Set("Authorization", simAuth)
	r.Header.Set("Content-Type", "application/json")
	return r
}

// simDo sends a request and returns status code + parsed JSON body.
func simDo(t *testing.T, client *http.Client, req *http.Request) (int, map[string]interface{}) {
	t.Helper()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(data, &result) // may fail for 204 or arrays — that's ok
	return resp.StatusCode, result
}

// simDoArray sends a request and returns status code + parsed JSON array.
func simDoArray(t *testing.T, client *http.Client, req *http.Request) (int, []map[string]interface{}) {
	t.Helper()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result []map[string]interface{}
	json.Unmarshal(data, &result)
	return resp.StatusCode, result
}

// simRegisterUser registers a user via the sim API.
func simRegisterUser(t *testing.T, ts *httptest.Server, client *http.Client, username, email, pwd string, latestVal int) int {
	t.Helper()
	body := map[string]string{"username": username, "email": email, "pwd": pwd}
	req := simRequest(t, "POST", ts.URL+"/register?latest="+itoa(latestVal), body)
	code, _ := simDo(t, client, req)
	return code
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func TestSimLatest(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()
	latest = -1

	// Register a user to trigger latest update
	code := simRegisterUser(t, ts, client, "test", "test@test.com", "foo", 1337)
	if code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}

	// Verify latest was updated
	req := simRequest(t, "GET", ts.URL+"/latest", nil)
	status, body := simDo(t, client, req)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if int(body["latest"].(float64)) != 1337 {
		t.Errorf("expected latest=1337, got %v", body["latest"])
	}
}

func TestSimRegister(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()
	latest = -1

	code := simRegisterUser(t, ts, client, "a", "a@a.a", "a", 1)
	if code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}

	// Verify latest
	req := simRequest(t, "GET", ts.URL+"/latest", nil)
	_, body := simDo(t, client, req)
	if int(body["latest"].(float64)) != 1 {
		t.Errorf("expected latest=1, got %v", body["latest"])
	}
}

func TestSimCreateMsg(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()
	latest = -1

	simRegisterUser(t, ts, client, "a", "a@a.a", "a", 1)

	// Post a message
	msg := map[string]string{"content": "Blub!"}
	req := simRequest(t, "POST", ts.URL+"/msgs/a?latest=2", msg)
	code, _ := simDo(t, client, req)
	if code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}

	// Verify latest
	req = simRequest(t, "GET", ts.URL+"/latest", nil)
	_, body := simDo(t, client, req)
	if int(body["latest"].(float64)) != 2 {
		t.Errorf("expected latest=2, got %v", body["latest"])
	}
}

func TestSimGetUserMsgs(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()
	latest = -1

	simRegisterUser(t, ts, client, "a", "a@a.a", "a", 1)

	// Post a message
	msg := map[string]string{"content": "Blub!"}
	req := simRequest(t, "POST", ts.URL+"/msgs/a?latest=2", msg)
	simDo(t, client, req)

	// Get user messages
	req = simRequest(t, "GET", ts.URL+"/msgs/a?no=20&latest=3", nil)
	status, msgs := simDoArray(t, client, req)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	found := false
	for _, m := range msgs {
		if m["content"] == "Blub!" && m["user"] == "a" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find message 'Blub!' from user 'a'")
	}

	// Verify latest
	req = simRequest(t, "GET", ts.URL+"/latest", nil)
	_, body := simDo(t, client, req)
	if int(body["latest"].(float64)) != 3 {
		t.Errorf("expected latest=3, got %v", body["latest"])
	}
}

func TestSimGetAllMsgs(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()
	latest = -1

	simRegisterUser(t, ts, client, "a", "a@a.a", "a", 1)

	msg := map[string]string{"content": "Blub!"}
	req := simRequest(t, "POST", ts.URL+"/msgs/a?latest=2", msg)
	simDo(t, client, req)

	// Get all messages
	req = simRequest(t, "GET", ts.URL+"/msgs?no=20&latest=4", nil)
	status, msgs := simDoArray(t, client, req)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	found := false
	for _, m := range msgs {
		if m["content"] == "Blub!" && m["user"] == "a" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find message 'Blub!' from user 'a' in all messages")
	}

	// Verify latest
	req = simRequest(t, "GET", ts.URL+"/latest", nil)
	_, body := simDo(t, client, req)
	if int(body["latest"].(float64)) != 4 {
		t.Errorf("expected latest=4, got %v", body["latest"])
	}
}

func TestSimFollowUnfollow(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()
	latest = -1

	// Register three users
	simRegisterUser(t, ts, client, "a", "a@a.a", "a", 1)
	simRegisterUser(t, ts, client, "b", "b@b.b", "b", 2)
	simRegisterUser(t, ts, client, "c", "c@c.c", "c", 3)

	// a follows b
	req := simRequest(t, "POST", ts.URL+"/fllws/a?latest=4", map[string]string{"follow": "b"})
	code, _ := simDo(t, client, req)
	if code != http.StatusNoContent {
		t.Fatalf("expected 204 for follow b, got %d", code)
	}

	// a follows c
	req = simRequest(t, "POST", ts.URL+"/fllws/a?latest=5", map[string]string{"follow": "c"})
	code, _ = simDo(t, client, req)
	if code != http.StatusNoContent {
		t.Fatalf("expected 204 for follow c, got %d", code)
	}

	// Verify a's follows
	req = simRequest(t, "GET", ts.URL+"/fllws/a?no=20&latest=6", nil)
	status, body := simDo(t, client, req)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}

	follows, ok := body["follows"].([]interface{})
	if !ok {
		t.Fatal("expected 'follows' to be an array")
	}
	followSet := make(map[string]bool)
	for _, f := range follows {
		followSet[f.(string)] = true
	}
	if !followSet["b"] {
		t.Error("expected 'b' in follows")
	}
	if !followSet["c"] {
		t.Error("expected 'c' in follows")
	}

	// a unfollows b
	req = simRequest(t, "POST", ts.URL+"/fllws/a?latest=7", map[string]string{"unfollow": "b"})
	code, _ = simDo(t, client, req)
	if code != http.StatusNoContent {
		t.Fatalf("expected 204 for unfollow, got %d", code)
	}

	// Verify b is no longer followed
	req = simRequest(t, "GET", ts.URL+"/fllws/a?no=20&latest=8", nil)
	_, body = simDo(t, client, req)
	follows = body["follows"].([]interface{})
	for _, f := range follows {
		if f.(string) == "b" {
			t.Error("expected 'b' to no longer be in follows")
		}
	}

	// Verify latest
	req = simRequest(t, "GET", ts.URL+"/latest", nil)
	_, body = simDo(t, client, req)
	if int(body["latest"].(float64)) != 8 {
		t.Errorf("expected latest=8, got %v", body["latest"])
	}
}

func TestSimAuthRequired(t *testing.T) {
	ts, client := setupTestServer(t)
	defer ts.Close()

	// Request without auth header should get 403
	req, _ := http.NewRequest("GET", ts.URL+"/msgs", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 without auth, got %d", resp.StatusCode)
	}
}
