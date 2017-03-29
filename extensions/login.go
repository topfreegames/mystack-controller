// mystack-controller api
// https://github.com/topfreegames/mystack/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package extensions

import (
	"fmt"
	"github.com/topfreegames/mystack/mystack-controller/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"
)

const (
	clientIDEnvVar     = "MYSTACK_GOOGLE_CLIENT_ID"
	clientSecretEnvVar = "MYSTACK_GOOGLE_CLIENT_SECRET"
)

var (
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv(clientIDEnvVar),
		ClientSecret: os.Getenv(clientSecretEnvVar),
		RedirectURL:  "http://localhost:57459/google-callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
)

//GenerateLoginURL generates the login url using googleapis OAuth2 Client Secret and OAuth2 Client ID
func GenerateLoginURL(oauthState string) (string, error) {
	if len(googleOauthConfig.ClientID) == 0 {
		return "", errors.NewAccessError(
			fmt.Sprintf("Undefined environmental variable %s", clientIDEnvVar),
			fmt.Errorf("Define your app's OAuth2 Client ID on %s environmental varianle and run again", clientIDEnvVar),
		)
	}

	if len(googleOauthConfig.ClientSecret) == 0 {
		return "", errors.NewAccessError(
			fmt.Sprintf("Undefined environmental variable %s", clientSecretEnvVar),
			fmt.Errorf("Define your app's OAuth2 Client Secret on %s environmental varianle and run again", clientSecretEnvVar),
		)
	}

	url := googleOauthConfig.AuthCodeURL(oauthState)
	return url, nil
}

//GetAccessToken exchange authorization code with access token
func GetAccessToken(code string) (string, error) {
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		err := errors.NewAccessError("GoogleCallback: Code exchange failed", err)
		return "", err
	}
	if !token.Valid() {
		err := errors.NewAccessError("GoogleCallback", fmt.Errorf("Invalid token received from Authorization Server"))
		return "", err
	}

	return token.AccessToken, nil
}
