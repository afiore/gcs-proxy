package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
)

const testSecret = "test-session-secret"

func TestValidatingSessionRedirectsToLoginWhenNoCookie(t *testing.T) {
	r, err := http.NewRequest("GET", "/path/to/resource?q=some%20param", nil)
	r.Host = "example.com"

	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	innerHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hello")
	}

	handler := http.HandlerFunc(ValidatingSession([]string{"lenses.io"}, testSecret, http.HandlerFunc(innerHandler)))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	location, _ := resp.Location()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
	if location.String() != "http://example.com/auth/google/login" {
		t.Errorf("Unexpected location %s", location.String())
	}
}

func TestValidatingSessionRedirectsToLoginWhenCookieIsTampered(t *testing.T) {
	r, err := http.NewRequest("GET", "/path/to/resource", nil)
	r.Host = "example.com"
	signedCookie := securecookie.New([]byte("wrong-secret"), nil)
	value := map[string]string{
		userHostedDomainKey: "acme.com",
	}
	encoded, err := signedCookie.Encode(sessionCookieName, value)
	if err != nil {
		t.Error(err)
	}
	cookie := http.Cookie{
		Name:  sessionCookieName,
		Value: encoded,
		Path:  "/",
	}
	r.AddCookie(&cookie)

	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	innerHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hello")
	}

	handler := http.HandlerFunc(ValidatingSession([]string{"lenses.io"}, testSecret, http.HandlerFunc(innerHandler)))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	location, err := resp.Location()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
	if location.String() != "http://example.com/auth/google/login" {
		t.Errorf("Unexpected location %s", location.String())
	}
}

func TestValidatingSessionForbiddenWhenInvalid(t *testing.T) {
	r, err := http.NewRequest("GET", "/path/to/resource", nil)
	signedCookie := securecookie.New([]byte(testSecret), nil)
	value := map[string]string{
		userHostedDomainKey: "acme.com",
	}
	encoded, err := signedCookie.Encode(sessionCookieName, value)
	if err != nil {
		t.Error(err)
	}
	cookie := http.Cookie{
		Name:  sessionCookieName,
		Value: encoded,
		Path:  "/",
	}
	r.AddCookie(&cookie)

	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	innerHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hello")
	}

	handler := http.HandlerFunc(ValidatingSession([]string{"lenses.io"}, testSecret, http.HandlerFunc(innerHandler)))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}

func TestValidatingSessionOkWhenValid(t *testing.T) {
	r, err := http.NewRequest("GET", "/path/to/resource", nil)
	signedCookie := securecookie.New([]byte(testSecret), nil)
	value := map[string]string{
		userHostedDomainKey: "lenses.io",
	}
	encoded, err := signedCookie.Encode(sessionCookieName, value)
	if err != nil {
		t.Error(err)
	}
	cookie := http.Cookie{
		Name:  sessionCookieName,
		Value: encoded,
		Path:  "/",
	}
	r.AddCookie(&cookie)

	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	innerHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hello")
	}

	handler := http.HandlerFunc(ValidatingSession([]string{"lenses.io"}, testSecret, http.HandlerFunc(innerHandler)))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code %d", resp.StatusCode)
	}
}
