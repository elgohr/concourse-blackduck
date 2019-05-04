package shared_test

import (
	"github.com/elgohr/blackduck-resource/shared"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func clean(t *testing.T) {
	if err := os.Remove(shared.ProjectCacheName); err != nil {
		t.Error(err)
	}
}

func TestGetProjectByNameQueriesForProject(t *testing.T) {
	var calledProjects bool
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledProjects = true
		b, err := ioutil.ReadFile("testdata/projects.json")
		if err != nil {
			t.Error(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Error(err)
		}
	}))
	defer h.Close()

	r := shared.Blackduck{}
	project, err := r.GetProjectByName(h.URL, "project1")
	if err != nil {
		t.Error(err)
	}
	if project.Name != "project1" {
		t.Errorf("Expected project1 to be loaded, but was %v", project.Name)
	}
	l := len(project.Meta.Links)
	if l != 9 {
		t.Errorf("Expected project to contain 9 links, but had %v", l)
	}
	if !calledProjects {
		t.Error("Didn't call the Blackduck api for projects")
	}

	clean(t)
}

func TestAuthenticationTokenFromBlackDuck(t *testing.T) {
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
		w.Header().Set("Set-Cookie", "Test=a; AUTHORIZATION_BEARER="+expectedToken+"; Max-Age=7200; Expires=Fri, " +
			"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer h.Close()

	r := shared.Blackduck{}
	token, err := r.GetAuthenticationToken(h.URL, expectedUser, expectedPassword)

	if err != nil {
		t.Error(err)
	}
	if expectedToken != token {
		t.Errorf("Expected token %v but was %v",expectedToken, token)
	}

	if !calledAuthenticated {
		t.Error("Didn't call the Blackduck api for token")
	}
}

func TestAuthenticationTokenFromBlackDuckTokenNotFound(t *testing.T) {
	var calledAuthenticated bool
	const expectedPassword = "TEST_PASSWORD"
	const expectedUser = "TEST_USER"
	const expectedToken = ""

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
		w.Header().Set("Set-Cookie", "Max-Age=7200; Expires=Fri, " +
			"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer h.Close()

	r := shared.Blackduck{}
	token, err := r.GetAuthenticationToken(h.URL, expectedUser, expectedPassword)

	if err == nil {
		t.Errorf("Expected error but was nil")
	}

	if expectedToken != token {
		t.Errorf("Expected token to be empty but was %v", token)
	}

	if !calledAuthenticated {
		t.Error("Didn't call the Blackduck api for token")
	}
}

func TestAuthenticationTokenFromBlackDuckAuthenticationFaild(t *testing.T) {
	var calledAuthenticated bool
	const expectedPassword = "TEST_PASSWORD"
	const expectedUser = "TEST_USER"
	const expectedToken = ""

	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledAuthenticated = true
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer h.Close()

	r := shared.Blackduck{}
	token, err := r.GetAuthenticationToken(h.URL, expectedUser, expectedPassword)

	if err.Error() != "authentication failed" {
		t.Errorf("Expected error was wrong")
	}

	if expectedToken != token {
		t.Errorf("Expected token to be empty but was %v", token)
	}

	if !calledAuthenticated {
		t.Error("Didn't call the Blackduck api for token")
	}
}

func TestGetProjectByNameCachesTheResult(t *testing.T) {
	var calledProjects int
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledProjects++
		b, err := ioutil.ReadFile("testdata/projects.json")
		if err != nil {
			t.Error(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Error(err)
		}
	}))
	defer h.Close()

	r := shared.Blackduck{}
	project1, err := r.GetProjectByName(h.URL, "project1")
	if err != nil {
		t.Error(err)
	}

	r2 := shared.Blackduck{}
	project2, err := r2.GetProjectByName(h.URL, "project1")
	if err != nil {
		t.Error(err)
	}

	if calledProjects > 1 {
		t.Error("Called blackduck multiple times")
	}
	if project1.Name != project2.Name || len(project1.Meta.Links) != len(project2.Meta.Links) {
		t.Error("Response didn't match")
	}

	clean(t)
}

func TestGetProjectByNameErrorsWhenResponseIsCorrupted(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{]`))
	}))
	defer h.Close()

	r := shared.Blackduck{}
	_, err := r.GetProjectByName(h.URL, "project1")
	if err == nil {
		t.Error("Should have errored, but didn't")
	}
}

func TestQueriesForTheLatestVersionsInChronologicalOrder(t *testing.T) {
	var calledVersions bool
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledVersions = true
		b, err := ioutil.ReadFile("testdata/versions.json")
		if err != nil {
			t.Error(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Error(err)
		}
	}))
	defer h.Close()

	r := shared.Blackduck{}
	refs, err := r.GetProjectVersions(&shared.Project{
		Name: "",
		Meta: shared.Meta{
			Links: []shared.Link{
				{
					Rel:  "versions",
					Href: h.URL,
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
	}

	if len(refs) != 2 {
		t.Errorf("Expected refs to contain 2 elements, but where %v", len(refs))
	}
	if refs[0].Ref != "0.1.1-DEVELOPMENT" {
		t.Error("Expected older entry first")
	}
	if refs[1].Ref != "0.1.2-DEVELOPMENT" {
		t.Error("Expected newer entry second")
	}
	if !calledVersions {
		t.Error("Didn't call the Blackduck api for versions")
	}
}
