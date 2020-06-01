package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/afiore/gcs-proxy/config"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.

const oauthGoogleURLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
const sessionCookieName = "gcs-proxy-session"
const loginTargetCookieName = "loginTargetPath"

//GoogleOAuthLoginPath is the canonical path for the Google oauth2 login handler
const GoogleOAuthLoginPath = "/auth/google/login"

//GoogleOAuthCallbackPath is the canonical path for the Google oauth2 callback
const GoogleOAuthCallbackPath = "/auth/google/callback"

const userHostedDomainKey = "hostedDomain"

type sessionValidationResult int

const (
	//denotes a successful validation
	valid sessionValidationResult = iota
	//denotes a failed validation (i.e. invalid user host domain or cookie signature mismatch)
	invalid
	//denotes the absense of a session cookie
	noCookie
)

func setSessionCookie(secret string, u userInfo, w http.ResponseWriter) error {
	signedCookie := securecookie.New([]byte(secret), nil)
	value := map[string]string{
		userHostedDomainKey: u.HostedDomain,
	}
	encoded, err := signedCookie.Encode(sessionCookieName, value)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:  sessionCookieName,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
	return err
}

type userInfo struct {
	HostedDomain string `json:"hd"`
}

func userInfoFromSessionCookie(secret string, r *http.Request) (userInfo, error) {
	var u userInfo
	signedCookie := securecookie.New([]byte(secret), nil)
	value := make(map[string]string)
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return u, err
	}
	if err = signedCookie.Decode(sessionCookieName, cookie.Value, &value); err == nil {
		hostedDomain, ok := value[userHostedDomainKey]
		if ok {
			u = userInfo{HostedDomain: hostedDomain}
		}
	}
	return u, err
}

func resolve(path string, r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, path)
}

func validateSession(validDomains []string, sessionSecret string, r *http.Request) (sessionValidationResult, error) {
	result := invalid
	u, err := userInfoFromSessionCookie(sessionSecret, r)
	if err != nil {
		return noCookie, nil
	}
	for _, domain := range validDomains {
		if domain == u.HostedDomain {
			result = valid
			break
		}
	}
	return result, nil
}

//ValidatingSession validates the session cookie redirecting to /auth/google/login if this is missing
func ValidatingSession(allowedHostDomains []string, sessionSecret string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := validateSession(allowedHostDomains, sessionSecret, r)
		if err != nil {
			log.Fatal(err)
		}
		switch result {
		case valid:
			handler(w, r)
		case invalid:
			http.Error(w, "Forbidden", http.StatusForbidden)
		case noCookie:
			//set a cookie to preserve the original target path across serveral requests
			targetPathCookie := http.Cookie{
				Name:  loginTargetCookieName,
				Path:  "/",
				Value: r.RequestURI,
			}
			log.Printf("setting %s cookie to: %s. Request: %v", loginTargetCookieName, r.RequestURI, r.URL)
			http.SetCookie(w, &targetPathCookie)
			http.Redirect(w, r, resolve(GoogleOAuthLoginPath, r), http.StatusTemporaryRedirect)
		}
	}
}

//GoogleOAuthHandlers provides server handlers for Google OAuth2 login and callback
type GoogleOAuthHandlers struct {
	Login    func(w http.ResponseWriter, r *http.Request)
	Callback func(w http.ResponseWriter, r *http.Request)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

//Handlers constructs the Google OAuth2 server handlers from the supplied configuration.
//
// Implementation is adapted from https://dev.to/douglasmakey/oauth2-example-with-go-3n8a
func Handlers(c config.ProgramConfig) GoogleOAuthHandlers {
	googleOauthConfig := func(r *http.Request) oauth2.Config {
		return oauth2.Config{
			RedirectURL:  resolve(GoogleOAuthCallbackPath, r),
			ClientID:     c.Web.OAuth.ClientID,
			ClientSecret: c.Web.OAuth.ClientSecret,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		}
	}

	getUserData := func(config *oauth2.Config, code string) (userInfo, error) {
		var u userInfo
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			return u, fmt.Errorf("code exchange wrong: %s", err.Error())
		}
		response, err := http.Get(oauthGoogleURLAPI + token.AccessToken)
		if err != nil {
			return u, fmt.Errorf("failed getting user info: %s", err.Error())
		}
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return u, fmt.Errorf("failed read response: %s", err.Error())
		}
		if err := json.Unmarshal(contents, &u); err != nil {
			log.Fatal(err)
		}

		return u, nil
	}

	login := func(w http.ResponseWriter, r *http.Request) {
		// Create oauthState cookie
		oauthState := generateStateOauthCookie(w)
		config := googleOauthConfig(r)

		/*
		   AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and

		   validate that it matches the the state query parameter on your redirect callback.
		*/
		u := config.AuthCodeURL(oauthState)
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}

	callback := func(w http.ResponseWriter, r *http.Request) {
		// Read oauthState from Cookie
		oauthState, _ := r.Cookie("oauthstate")
		config := googleOauthConfig(r)

		if r.FormValue("state") != oauthState.Value {
			log.Println("invalid oauth google state")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		userData, err := getUserData(&config, r.FormValue("code"))
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		setSessionCookie(c.Web.OAuth.SessionSecret, userData, w)

		cookie, err := r.Cookie(loginTargetCookieName)
		if err != nil {
			log.Printf("cookie not found: %v", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		} else {

			log.Printf("found cookie %s. Value: %s", loginTargetCookieName, cookie.Value)
			expired := http.Cookie{
				Value:  "",
				Path:   "/",
				Name:   loginTargetCookieName,
				MaxAge: -1,
			}
			http.SetCookie(w, &expired)
			http.Redirect(w, r, cookie.Value, http.StatusTemporaryRedirect)
		}

	}

	return GoogleOAuthHandlers{
		Login:    login,
		Callback: callback,
	}

}
