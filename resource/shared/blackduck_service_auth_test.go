package shared

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetsAuthenticationTokenFromBlackDuck(t *testing.T) {
	var calledAuthenticated bool
	const expectedPassword = "TEST_PASSWORD"
	const expectedUser = "TEST_USER"
	const expectedToken = "TEST_TOKEN"

	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledAuthenticated = true
		r.ParseForm()
		gotUser := r.Form.Get("j_username")
		if gotUser != expectedUser {
			t.Errorf("Expected username %v , but got %v", expectedUser, gotUser)
		}
		gotPw := r.Form.Get("j_password")
		if gotPw != expectedPassword {
			t.Errorf("Expected password %v , but got %v", expectedPassword, gotPw)
		}
		if !strings.HasSuffix(r.RequestURI, "/j_spring_security_check") {
			t.Errorf("Expected url to end with /j_spring_security_check , but was %v", r.RequestURI)
		}
		w.Header().Set("Set-Cookie", "Test=a; AUTHORIZATION_BEARER="+expectedToken+"; Max-Age=7200; Expires=Fri, "+
			"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer h.Close()

	r := NewBlackduck()
	err := r.Authenticate(h.URL, expectedUser, expectedPassword)
	if err != nil {
		t.Error(err)
	}

	if expectedToken != r.token {
		t.Errorf("Expected token %v but was %v", expectedToken, r.token)
	}

	if !calledAuthenticated {
		t.Error("Didn't call the Blackduck api for token")
	}
}

func TestErrorsWithoutAuthenticationTokenFromBlackDuck(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "Max-Age=7200; Expires=Fri, "+
			"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer h.Close()

	r := NewBlackduck()
	err := r.Authenticate(h.URL, "user", "password")
	if err == nil {
		t.Errorf("Expected error but was nil")
	}

	if len(r.token) != 0 {
		t.Errorf("Expected token to be empty but was %v", r.token)
	}

}

func TestAuthenticationTokenFromBlackDuckAuthenticationFaild(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer h.Close()

	r := NewBlackduck()
	err := r.Authenticate(h.URL, "user", "password")
	if err.Error() != "authentication failed" {
		t.Errorf("Expected error was wrong")
	}

	if len(r.token) != 0 {
		t.Errorf("Expected token to be empty but was %v", r.token)
	}

}
