// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package extensions

import (
	"fmt"
	"github.com/topfreegames/mystack-controller/errors"
	"github.com/topfreegames/mystack-controller/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL: "http://localhost:57459/google-callback",
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint: google.Endpoint,
	}
)

//getClientCredentials receive Credentials interface and fills ClientID and ClientSecret
func getClientCredentials(credentials models.Credentials) {
	googleOauthConfig.ClientID = credentials.GetID()
	googleOauthConfig.ClientSecret = credentials.GetSecret()
}

//GenerateLoginURL generates the login url using googleapis OAuth2 Client Secret and OAuth2 Client ID
func GenerateLoginURL(oauthState string, credentials models.Credentials) (string, error) {
	getClientCredentials(credentials)

	if len(googleOauthConfig.ClientID) == 0 {
		return "", errors.NewAccessError(
			fmt.Sprintf("Undefined environment variable %s", models.ClientIDEnvVar),
			fmt.Errorf("Define your app's OAuth2 Client ID on %s environment variable and run again", models.ClientIDEnvVar),
		)
	}

	if len(googleOauthConfig.ClientSecret) == 0 {
		return "", errors.NewAccessError(
			fmt.Sprintf("Undefined environment variable %s", models.ClientSecretEnvVar),
			fmt.Errorf("Define your app's OAuth2 Client Secret on %s environment variable and run again", models.ClientSecretEnvVar),
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
