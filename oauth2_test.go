// Copyright 2014 GoIncremental Limited. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oauth2_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/mix3/fever-oauth2"
	"github.com/mix3/fever-sessions"
	"github.com/mix3/fever/mux"
)

func Test_LoginRedirect(t *testing.T) {
	recorder := httptest.NewRecorder()
	store := sessions.NewMemoryStore()
	defer func() {
		_ = store.Close()
	}()
	ss := sessions.New(store, "myapp_session")
	m := mux.New()
	m.Use(ss.Middleware)
	m.Use(oauth2.Google(&oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "refresh_url",
		Scopes:       []string{"x", "y"},
	}))
	r, _ := http.NewRequest("GET", "/login", nil)
	m.ServeHTTP(recorder, r)

	location := recorder.HeaderMap["Location"][0]
	if recorder.Code != 302 {
		t.Errorf("Not being redirected to the auth page.")
	}
	t.Logf(location)
	if strings.HasPrefix("https://accounts.google.com/o/oauth2/auth?access_type=online&approval_prompt=auto&client_id=client_id&redirect_uri=refresh_url&response_type=code&scope=x+y&state=", location) {
		t.Errorf("Not being redirected to the right page, %v found", location)
	}
}

func Test_LoginRedirectAfterLoginRequired(t *testing.T) {
	recorder := httptest.NewRecorder()
	store := sessions.NewMemoryStore()
	defer func() {
		_ = store.Close()
	}()
	ss := sessions.New(store, "myapp_session")
	m := mux.New()
	m.Use(ss.Middleware)
	m.Use(oauth2.Google(&oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "refresh_url",
		Scopes:       []string{"x", "y"},
	}))
	m.Use(oauth2.LoginRequired())
	m.Get("/login-required").ThenFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		t.Log("hi there")
		fmt.Fprintf(w, "OK")
	})
	r, _ := http.NewRequest("GET", "/login-required?key=value", nil)
	m.ServeHTTP(recorder, r)

	location := recorder.HeaderMap["Location"][0]
	if recorder.Code != 302 {
		t.Errorf("Not being redirected to the auth page.")
	}
	if location != "/login?next=%2Flogin-required%3Fkey%3Dvalue" {
		t.Errorf("Not being redirected to the right page, %v found", location)
	}
}

func Test_Logout(t *testing.T) {
	recorder := httptest.NewRecorder()
	store := sessions.NewMemoryStore()
	defer func() {
		_ = store.Close()
	}()
	ss := sessions.New(store, "myapp_session")
	m := mux.New()
	m.Use(ss.Middleware)
	m.Use(oauth2.Google(&oauth2.Config{
		ClientID:     "foo",
		ClientSecret: "foo",
		RedirectURL:  "foo",
	}))
	m.Use(oauth2.LoginRequired())
	m.Get("/").ThenFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		oauth2.SetToken(c, "dummy token")
		fmt.Fprintf(w, "OK")
	})
	m.Get("/get").ThenFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		tok := oauth2.GetToken(c)
		if tok != nil {
			t.Errorf("User credentials are still kept in the session.")
		}
		fmt.Fprintf(w, "OK")
	})
	r, _ := http.NewRequest("GET", "/login-required?key=value", nil)
	m.ServeHTTP(recorder, r)

	logout, _ := http.NewRequest("GET", "/logout", nil)
	index, _ := http.NewRequest("GET", "/", nil)

	m.ServeHTTP(httptest.NewRecorder(), index)
	m.ServeHTTP(recorder, logout)

	if recorder.Code != 302 {
		t.Errorf("Not being redirected to the next page.")
	}
}

func Test_LogoutOnAccessTokenExpiration(t *testing.T) {
	recorder := httptest.NewRecorder()
	store := sessions.NewMemoryStore()
	defer func() {
		_ = store.Close()
	}()
	ss := sessions.New(store, "myapp_session")
	m := mux.New()
	m.Use(ss.Middleware)
	m.Use(oauth2.Google(&oauth2.Config{
		ClientID:     "foo",
		ClientSecret: "foo",
		RedirectURL:  "foo",
	}))
	m.Use(oauth2.LoginRequired())
	m.Get("/addtoken").ThenFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		oauth2.SetToken(c, "dummy token")
		fmt.Fprintf(w, "OK")
	})
	m.Get("/").ThenFunc(func(c context.Context, w http.ResponseWriter, req *http.Request) {
		tok := oauth2.GetToken(c)
		if tok != nil {
			t.Errorf("User not logged out although access token is expired. %v\n", tok)
		}
	})

	addtoken, _ := http.NewRequest("GET", "/addtoken", nil)
	index, _ := http.NewRequest("GET", "/", nil)
	m.ServeHTTP(recorder, addtoken)
	m.ServeHTTP(recorder, index)
}

func Test_LoginRequired(t *testing.T) {
	recorder := httptest.NewRecorder()
	store := sessions.NewMemoryStore()
	defer func() {
		_ = store.Close()
	}()
	ss := sessions.New(store, "myapp_session")
	m := mux.New()
	m.Use(ss.Middleware)
	m.Use(oauth2.Google(&oauth2.Config{
		ClientID:     "foo",
		ClientSecret: "foo",
		RedirectURL:  "foo",
	}))
	m.Use(oauth2.LoginRequired())
	m.Get("/").ThenFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	r, _ := http.NewRequest("GET", "/", nil)
	m.ServeHTTP(recorder, r)
	if recorder.Code != 302 {
		t.Errorf("Not being redirected to the auth page although user is not logged in. %d\n", recorder.Code)
	}
}
