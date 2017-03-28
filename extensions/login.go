// mystack-controller api
// https://github.com/topfreegames/mystack
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package extensions

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
)

var (
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("googlekey"),
		ClientSecret: os.Getenv("googlesecret"),
		RedirectURL:  "http://localhost:8989/google-callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
)

//GenerateLoginURL generates the login url using googleapis OAuth2 Client Secret and OAuth2 Client ID
func GenerateLoginURL(oauthState string) string {
	url := googleOauthConfig.AuthCodeURL(oauthState)
	return url
}
