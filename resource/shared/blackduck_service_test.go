package shared

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func clean(t *testing.T) {
	if err := os.Remove(ProjectCacheName); err != nil {
		t.Error(err)
	}
}

func TestGetProjectQueriesForProject(t *testing.T) {
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

	r := NewBlackduck()
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

func TestGetProjectAddsAuthentication(t *testing.T) {
	const expectedToken = "TEST_TOKEN"
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "AUTHORIZATION_BEARER="+expectedToken {
			t.Errorf("Expected the bearer to be send via Cookie (whoever knows why...), but got %v", r.Header.Get("Cookie"))
		}
		b, err := ioutil.ReadFile("testdata/projects.json")
		if err != nil {
			t.Error(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Error(err)
		}
	}))
	defer h.Close()

	r := NewBlackduck()
	r.token = expectedToken
	_, err := r.GetProjectByName(h.URL, "project1")
	if err != nil {
		t.Error(err)
	}

	clean(t)
}

func TestGetProjectCachesTheResult(t *testing.T) {
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

	r := NewBlackduck()
	project1, err := r.GetProjectByName(h.URL, "project1")
	if err != nil {
		t.Error(err)
	}

	r2 := NewBlackduck()
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

func TestGetProjectErrorsWhenResponseIsCorrupted(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{]`))
	}))
	defer h.Close()

	r := NewBlackduck()
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

	r := NewBlackduck()
	refs, err := r.GetProjectVersions(&Project{
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

func TestQueriesForTheLatestVersionsWithAuthentication(t *testing.T) {
	const expectedToken = "TEST_TOKEN"
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "AUTHORIZATION_BEARER="+expectedToken {
			t.Errorf("Expected the bearer to be send via Cookie (whoever knows why...), but got %v", r.Header.Get("Cookie"))
		}
		b, err := ioutil.ReadFile("testdata/versions.json")
		if err != nil {
			t.Error(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Error(err)
		}
	}))
	defer h.Close()

	r := NewBlackduck()
	r.token = expectedToken
	_, err := r.GetProjectVersions(&Project{
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
}
