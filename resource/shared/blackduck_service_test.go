package shared

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func clean(t *testing.T) {
	if err := os.Remove(ProjectCacheName); err != nil {
		t.Error(err)
	}
}

func TestGetProjectQueriesForProject(t *testing.T) {
	var (
		calledProjects      bool
		calledAuthenticated bool
	)
	const (
		expectedPassword = "TEST_PASSWORD"
		expectedUser     = "TEST_USER"
		expectedToken    = "TEST_TOKEN"
	)
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			calledProjects = true
			if r.Header.Get("Cookie") != "AUTHORIZATION_BEARER=TEST_TOKEN" {
				t.Errorf("Expected to get the authentication token, but got %v", r.Header.Get("Cookie"))
			}
			b, err := ioutil.ReadFile("testdata/projects.json")
			if err != nil {
				t.Error(err)
			}
			if _, err := w.Write(b); err != nil {
				t.Error(err)
			}
		} else if r.Method == "POST" {
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
		}

	}))
	defer h.Close()

	r := NewBlackduck()
	project, err := r.GetProjectByName(Source{
		Url:      h.URL,
		Name:     "project1",
		Username: expectedUser,
		Password: expectedPassword,
	})
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
	if !calledAuthenticated {
		t.Error("Didn't call the Blackduck api for authentication token")
	}

	clean(t)
}

func TestGetProjectCachesTheResult(t *testing.T) {
	var (
		calledProjects     int
		calledAuthenticate int
	)
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			calledAuthenticate++
			w.Header().Set("Set-Cookie", "AUTHORIZATION_BEARER=TOKEN; Max-Age=7200; Expires=Fri, "+
				"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
			w.WriteHeader(http.StatusNoContent)
		} else {
			calledProjects++
			b, err := ioutil.ReadFile("testdata/projects.json")
			if err != nil {
				t.Error(err)
			}
			if _, err := w.Write(b); err != nil {
				t.Error(err)
			}
		}
	}))
	defer h.Close()

	r := NewBlackduck()
	project1, err := r.GetProjectByName(Source{
		Url:  h.URL,
		Name: "project1",
	})
	if err != nil {
		t.Error(err)
	}

	r2 := NewBlackduck()
	project2, err := r2.GetProjectByName(Source{
		Url:  h.URL,
		Name: "project1",
	})
	if err != nil {
		t.Error(err)
	}

	if calledProjects > 1 {
		t.Error("Called blackduck multiple times")
	}
	if calledAuthenticate > 1 {
		t.Error("Called blackduck authenticate multiple times")
	}
	if project1.Name != project2.Name || len(project1.Meta.Links) != len(project2.Meta.Links) {
		t.Error("Response didn't match")
	}

	clean(t)
}

func TestGetProjectErrorsWhenResponseIsCorrupted(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer h.Close()

	r := NewBlackduck()
	_, err := r.GetProjectByName(Source{
		Url:  h.URL,
		Name: "project1",
	})
	if err.Error() != "authentication failed" {
		t.Error("Should have errored for authentication, but didn't")
	}
}

func TestGetProjectErrorsWhenAuthenticationFails(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Set-Cookie", "AUTHORIZATION_BEARER=TOKEN; Max-Age=7200; Expires=Fri, "+
				"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
			w.WriteHeader(http.StatusNoContent)
		} else {
			_, err := w.Write([]byte(`{]`))
			if err != nil {
				t.Error(err)
			}
		}
	}))
	defer h.Close()

	r := NewBlackduck()
	_, err := r.GetProjectByName(Source{
		Url:  h.URL,
		Name: "project1",
	})
	if err.Error() != "invalid character ']' looking for beginning of object key string" {
		t.Error("Should have errored, but didn't")
	}
}

func TestQueriesForTheLatestVersionsInChronologicalOrder(t *testing.T) {
	var calledVersions bool
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Set-Cookie", "AUTHORIZATION_BEARER=TOKEN; Max-Age=7200; Expires=Fri, "+
				"26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly")
			w.WriteHeader(http.StatusNoContent)
		} else {
			calledVersions = true
			if r.Header.Get("Cookie") != "AUTHORIZATION_BEARER=TOKEN" {
				t.Errorf("Expected the bearer to be send via Cookie (whoever knows why...), but got %v", r.Header.Get("Cookie"))
			}
			b, err := ioutil.ReadFile("testdata/versions.json")
			if err != nil {
				t.Error(err)
			}
			if _, err := w.Write(b); err != nil {
				t.Error(err)
			}
		}
	}))
	defer h.Close()

	r := NewBlackduck()
	refs, err := r.GetProjectVersions(Source{
		Url:  h.URL,
		Name: "project1",
	}, &Project{
		Name: "",
		Meta: Meta{
			Links: []Link{
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
	if refs[0].Updated.String() != "2019-04-18 09:12:48.511 +0000 UTC" {
		t.Errorf("Expected older entry first, but was %v", refs[0].Updated.String())
	}
	if refs[1].Updated.String() != "2019-04-20 09:12:48.511 +0000 UTC" {
		t.Errorf("Expected newer entry second, but was %v", refs[1].Updated.String())
	}
	if !calledVersions {
		t.Error("Didn't call the Blackduck api for versions")
	}
}
