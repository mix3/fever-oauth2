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

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/joho/godotenv"
	"github.com/mix3/fever"
	"github.com/mix3/fever-oauth2"
	"github.com/mix3/fever-sessions"
	"github.com/mix3/fever/mux"
)

func getEnv(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		v = defaultValue
	}
	return v
}

func main() {
	//Loads environment variables from a .env file
	godotenv.Load("examples/environment")

	var (
		clientID     = getEnv("OAUTH2_CLIENT_ID", "client_id")
		clientSecret = getEnv("OAUTH2_CLIENT_SECRET", "client_secret")
		redirectURL  = getEnv("OAUTH2_REDIRECT_URL", "redirect_url")
	)

	store := sessions.NewMemoryStore()
	defer func() {
		_ = store.Close()
	}()
	ss := sessions.New(store, "myapp_session")

	m := mux.New()
	m.Use(ss.Middleware)
	m.Use(oauth2.Google(&oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
	}))
	m.Get("/").ThenFunc(func(c context.Context, w http.ResponseWriter, req *http.Request) {
		token := oauth2.GetToken(c)
		if token == nil || !token.Valid() {
			fmt.Fprintf(w, "not logged in, or the access token is expired")
			return
		}
		fmt.Fprintf(w, "logged in")
		return
	})
	m.Get("/restrict").Use(oauth2.LoginRequired()).ThenFunc(func(c context.Context, w http.ResponseWriter, req *http.Request) {
		token := oauth2.GetToken(c)
		fmt.Fprintf(w, "OK: %s", token.Access())
	})

	fever.Run(":19300", 10*time.Second, m)
}
