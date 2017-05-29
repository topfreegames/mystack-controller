// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package extensions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
func getClientCredentials(credentials models.Credentials) (*oauth2.Config, error) {
	googleOauthConfig.ClientID = credentials.GetID()
	if len(googleOauthConfig.ClientID) == 0 {
		return nil, errors.NewAccessError(
			fmt.Sprintf("Undefined environment variable %s", models.ClientIDEnvVar),
			fmt.Errorf("Define your app's OAuth2 Client ID on %s environment variable and run again", models.ClientIDEnvVar),
		)
	}

	googleOauthConfig.ClientSecret = credentials.GetSecret()
	if len(googleOauthConfig.ClientSecret) == 0 {
		return nil, errors.NewAccessError(
			fmt.Sprintf("Undefined environment variable %s", models.ClientSecretEnvVar),
			fmt.Errorf("Define your app's OAuth2 Client Secret on %s environment variable and run again", models.ClientSecretEnvVar),
		)
	}

	return googleOauthConfig, nil
}

//GenerateLoginURL generates the login url using googleapis OAuth2 Client Secret and OAuth2 Client ID
func GenerateLoginURL(oauthState string, credentials models.Credentials) (string, error) {
	googleOauthConfig, err := getClientCredentials(credentials)
	if err != nil {
		return "", err
	}

	url := googleOauthConfig.AuthCodeURL(oauthState, oauth2.AccessTypeOffline)
	return url, nil
}

//GetAccessToken exchange authorization code with access token
func GetAccessToken(code string) (*oauth2.Token, error) {
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		err := errors.NewAccessError("GoogleCallback: Code exchange failed", err)
		return nil, err
	}
	if !token.Valid() {
		err := errors.NewAccessError("GoogleCallback", fmt.Errorf("Invalid token received from Authorization Server"))
		return nil, err
	}

	return token, nil
}

//Authenticate authenticates an access token or gets a new one with the refresh token
//The returned string is either the error message or the user email
func Authenticate(
	token *oauth2.Token,
	credentials models.Credentials,
	db models.DB,
) (string, int, error) {
	var email string
	var status int

	googleOauthConfig, err := getClientCredentials(credentials)
	if err != nil {
		return email, status, errors.NewAccessError("error getting access token", err)
	}

	newToken := new(oauth2.Token)
	*newToken = *token
	expired := time.Now().UTC().After(token.Expiry)
	if expired {
		var err error
		newToken, err = googleOauthConfig.TokenSource(oauth2.NoContext, token).Token()
		if err != nil {
			return email, status, errors.NewAccessError("error getting access token", err)
		}
	}

	client := googleOauthConfig.Client(oauth2.NoContext, newToken)
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=%s", newToken.AccessToken)
	resp, err := client.Get(url)

	defer resp.Body.Close()

	status = resp.StatusCode
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return email, status, errors.NewGenericError("error reading google response body", err)
	}

	if status != http.StatusOK {
		return string(bts), status, nil
	}

	var bodyObj map[string]interface{}
	json.Unmarshal(bts, &bodyObj)
	email = bodyObj["email"].(string)
	if expired {
		err = SaveToken(newToken, email, token.AccessToken, db)
		if err != nil {
			return email, http.StatusInternalServerError, errors.NewDatabaseError(err)
		}
	}

	return email, status, nil
}
