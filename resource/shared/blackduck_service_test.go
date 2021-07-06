package shared

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

const validCookie = "AUTHORIZATION_BEARER=TOKEN; Max-Age=7200; Expires=Fri, 26-Apr-2019 13:37:14 GMT; Path=/; secure; Secure; HttpOnly"

func TestGetProjectQueriesForProject(t *testing.T) {
	var (
		calledProjects      bool
		calledAuthenticated bool
	)
	const (
		expectedPassword = "TEST_PASSWORD"
		expectedUser     = "TEST_USER"
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			calledProjects = true
			assert.Equal(t, "AUTHORIZATION_BEARER=TOKEN", r.Header.Get("Cookie"))

			b, err := ioutil.ReadFile("testdata/projects.json")
			assert.NoError(t, err)

			_, err = w.Write(b)
			assert.NoError(t, err)
		} else if r.Method == http.MethodPost {
			calledAuthenticated = true

			assert.NoError(t, r.ParseForm())

			assert.Equal(t, r.Form.Get("j_username"), expectedUser)
			assert.Equal(t, r.Form.Get("j_password"), expectedPassword)
			assert.True(t, strings.HasSuffix(r.RequestURI, "/j_spring_security_check"), "Expected url to end with /j_spring_security_check , but was "+r.RequestURI)
			w.Header().Set("Set-Cookie", validCookie)
			w.WriteHeader(http.StatusNoContent)
		}

	}))
	defer ts.Close()

	r := NewBlackduck()
	project, err := r.GetProjectByName(Source{
		Url:      ts.URL,
		Name:     "project1",
		Username: expectedUser,
		Password: expectedPassword,
	})
	require.NoError(t, err)
	require.Equal(t, "project1", project.Name)
	require.Equal(t, 9, len(project.Meta.Links))
	require.True(t, calledProjects)
	require.True(t, calledAuthenticated)

	clean(t)
}

func TestGetProjectCachesTheResult(t *testing.T) {
	var (
		calledProjects     int
		calledAuthenticate int
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			calledAuthenticate++
			w.Header().Set("Set-Cookie", validCookie)
			w.WriteHeader(http.StatusNoContent)
		} else {
			calledProjects++
			b, err := ioutil.ReadFile("testdata/projects.json")
			require.NoError(t, err)
			_, err = w.Write(b)
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	r := NewBlackduck()
	project1, err := r.GetProjectByName(Source{
		Url:  ts.URL,
		Name: "project1",
	})
	require.NoError(t, err)

	r2 := NewBlackduck()
	project2, err := r2.GetProjectByName(Source{
		Url:  ts.URL,
		Name: "project1",
	})
	require.NoError(t, err)

	require.GreaterOrEqual(t, 1, calledProjects)
	require.Equal(t, project1, project2)

	clean(t)
}

func TestGetProjectErrorsWhenResponseIsCorrupted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	r := NewBlackduck()
	_, err := r.GetProjectByName(Source{
		Url:  ts.URL,
		Name: "project1",
	})
	expError := "GetProjectByName: Authentication: authentication failed"
	require.EqualError(t, err, expError)
}

func TestGetProjectErrorsWhenAuthenticationFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Set-Cookie", validCookie)
			w.WriteHeader(http.StatusNoContent)
		} else {
			_, err := w.Write([]byte(`{]`))
			assert.NoError(t, err)
		}
	}))
	defer ts.Close()

	r := NewBlackduck()
	_, err := r.GetProjectByName(Source{
		Url:  ts.URL,
		Name: "project1",
	})
	expError := "GetProjectByName: Decode: invalid character ']' looking for beginning of object key string"
	require.EqualError(t, err, expError)
}

func TestGetProjectByNameSetsInsecureHttpWhenInsecure(t *testing.T) {
	r := NewBlackduck()
	called := false
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := r.GetProjectByName(Source{
		Url:  ts.URL,
		Name: "project1",
	})
	require.Error(t, err)
	require.True(t, strings.HasSuffix(err.Error(), "x509: certificate signed by unknown authority"), err.Error())
	require.False(t, called)

	_, _ = r.GetProjectByName(Source{
		Url:      ts.URL,
		Name:     "project1",
		Insecure: true,
	})
	require.True(t, called)
}

func TestQueriesForTheLatestVersionsInChronologicalOrder(t *testing.T) {
	var calledVersions bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Set-Cookie", validCookie)
			w.WriteHeader(http.StatusNoContent)
		} else {
			calledVersions = true
			require.Equal(t, "AUTHORIZATION_BEARER=TOKEN", r.Header.Get("Cookie"))
			b, err := ioutil.ReadFile("testdata/versions.json")
			require.NoError(t, err)
			_, err = w.Write(b)
			require.NoError(t, err)
		}
	}))
	defer ts.Close()

	r := NewBlackduck()
	refs, err := r.GetProjectVersions(Source{
		Url:  ts.URL,
		Name: "project1",
	}, &Project{
		Name: "",
		Meta: Meta{
			Links: []Link{
				{
					Rel:  "versions",
					Href: ts.URL,
				},
			},
		},
	})
	require.NoError(t, err)

	require.Equal(t, 2,  len(refs))
	require.Equal(t, "2019-04-18 09:12:48.511 +0000 UTC", refs[0].Updated.String())
	require.Equal(t, "2019-04-20 09:12:48.511 +0000 UTC", refs[1].Updated.String())
	require.True(t, calledVersions)
}
