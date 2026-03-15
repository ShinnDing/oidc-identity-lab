package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	oidcProvider    *oidc.Provider
	oauth2Config    oauth2.Config
	idTokenVerifier *oidc.IDTokenVerifier
	store           *sessions.CookieStore
)

type UserClaims struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
}

type SessionUser struct {
	IDTokenClaims  UserClaims `json:"id_token_claims"`
	UserInfoClaims UserClaims `json:"user_info_claims"`
}

func main() {
	gob.Register(UserClaims{})
	gob.Register(SessionUser{})

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()

	var providerErr error
	oidcProvider, providerErr = oidc.NewProvider(ctx, "http://localhost:8080/realms/oidc-lab")
	if providerErr != nil {
		log.Fatalf("failed to get OIDC provider: %v", providerErr)
	}

	idTokenVerifier = oidcProvider.Verifier(&oidc.Config{
		ClientID: "oidc-lab-app",
	})

	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("OIDC_CLIENT_SECRET is not set")
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET is not set")
	}

	store = sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}

	oauth2Config = oauth2.Config{
		ClientID:     "oidc-lab-app",
		ClientSecret: clientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  "http://localhost:3000/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/claims", claimsHandler)
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("App running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html><body>
		<h2>OIDC app is running.</h2>
		<p><a href="/login">Login with Keycloak</a></p>
		<p><a href="/profile">View profile</a></p>
		<p><a href="/claims">View raw claims</a></p>
		<p><a href="/logout">Logout</a></p>
	</body></html>`)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "oidc-session")
	if err != nil {
		http.Error(w, "failed to get session", http.StatusInternalServerError)
		return
	}

	state, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	codeVerifier, err := generateRandomString(32)
	if err != nil {
		http.Error(w, "failed to generate PKCE code verifier", http.StatusInternalServerError)
		return
	}

	codeChallenge := generateCodeChallenge(codeVerifier)

	session.Values["oauth_state"] = state
	session.Values["pkce_code_verifier"] = codeVerifier

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}

	url := oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(w, r, url, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "oidc-session")
	if err != nil {
		http.Error(w, "failed to get session", http.StatusInternalServerError)
		return
	}

	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || expectedState == "" {
		http.Error(w, "missing expected state in session", http.StatusBadRequest)
		return
	}

	returnedState := r.URL.Query().Get("state")
	if returnedState == "" || returnedState != expectedState {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	codeVerifier, ok := session.Values["pkce_code_verifier"].(string)
	if !ok || codeVerifier == "" {
		http.Error(w, "missing PKCE code verifier in session", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := oauth2Config.Exchange(
		r.Context(),
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		log.Printf("token exchange failed: %v", err)
		http.Error(w, fmt.Sprintf("failed to exchange code for token: %v", err), http.StatusInternalServerError)
		return
	}

	delete(session.Values, "oauth_state")
	delete(session.Values, "pkce_code_verifier")

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "missing id_token", http.StatusInternalServerError)
		return
	}

	idToken, err := idTokenVerifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}

	var idClaims UserClaims
	if err := idToken.Claims(&idClaims); err != nil {
		http.Error(w, "failed to parse ID token claims", http.StatusInternalServerError)
		return
	}

	userInfo, err := oidcProvider.UserInfo(
		context.WithValue(r.Context(), oauth2.HTTPClient, http.DefaultClient),
		oauth2.StaticTokenSource(token),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get userinfo: %v", err), http.StatusInternalServerError)
		return
	}

	var userInfoClaims UserClaims
	if err := userInfo.Claims(&userInfoClaims); err != nil {
		http.Error(w, "failed to parse userinfo claims", http.StatusInternalServerError)
		return
	}

	session.Values["user"] = SessionUser{
		IDTokenClaims:  idClaims,
		UserInfoClaims: userInfoClaims,
	}

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "id_token_hint",
		Value:    rawIDToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "failed to save session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser, ok := getSessionUser(w, r)
	if !ok {
		return
	}

	fmt.Fprintf(w, `<html><body>
		<h2>Profile</h2>

		<h3>ID Token Claims</h3>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Username:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>

		<h3>UserInfo Claims</h3>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Username:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>

		<p><a href="/">Back to home</a></p>
		<p><a href="/claims">View raw claims</a></p>
		<p><a href="/logout">Logout</a></p>
	</body></html>`,
		html.EscapeString(sessionUser.IDTokenClaims.Name),
		html.EscapeString(sessionUser.IDTokenClaims.PreferredUsername),
		html.EscapeString(sessionUser.IDTokenClaims.Email),
		html.EscapeString(sessionUser.UserInfoClaims.Name),
		html.EscapeString(sessionUser.UserInfoClaims.PreferredUsername),
		html.EscapeString(sessionUser.UserInfoClaims.Email),
	)
}

func claimsHandler(w http.ResponseWriter, r *http.Request) {
	sessionUser, ok := getSessionUser(w, r)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	idJSON, err := json.MarshalIndent(sessionUser.IDTokenClaims, "", "  ")
	if err != nil {
		http.Error(w, "failed to render ID token claims", http.StatusInternalServerError)
		return
	}

	userInfoJSON, err := json.MarshalIndent(sessionUser.UserInfoClaims, "", "  ")
	if err != nil {
		http.Error(w, "failed to render userinfo claims", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `<html><body>
		<h2>Raw Claims</h2>

		<h3>ID Token Claims</h3>
		<pre>%s</pre>

		<h3>UserInfo Claims</h3>
		<pre>%s</pre>

		<p><a href="/profile">Back to profile</a></p>
		<p><a href="/logout">Logout</a></p>
	</body></html>`,
		html.EscapeString(string(idJSON)),
		html.EscapeString(string(userInfoJSON)),
	)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "oidc-session")
	if err != nil {
		http.Error(w, "failed to get session", http.StatusInternalServerError)
		return
	}

	var rawIDToken string
	if cookie, err := r.Cookie("id_token_hint"); err == nil {
		rawIDToken = cookie.Value
	}

	session.Values = map[interface{}]interface{}{}
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "failed to clear session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oidc-session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "id_token_hint",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	logoutRedirect := os.Getenv("KEYCLOAK_LOGOUT_REDIRECT")
	if logoutRedirect == "" {
		logoutRedirect = "http://localhost:3000/"
	}

	if rawIDToken == "" {
		http.Error(w, "missing id token for logout; log in again and retry", http.StatusBadRequest)
		return
	}

	logoutURL := "http://localhost:8080/realms/oidc-lab/protocol/openid-connect/logout" +
		"?id_token_hint=" + url.QueryEscape(rawIDToken) +
		"&post_logout_redirect_uri=" + url.QueryEscape(logoutRedirect)

	http.Redirect(w, r, logoutURL, http.StatusFound)
}

func getSessionUser(w http.ResponseWriter, r *http.Request) (SessionUser, bool) {
	session, err := store.Get(r, "oidc-session")
	if err != nil {
		http.Error(w, "failed to get session", http.StatusInternalServerError)
		return SessionUser{}, false
	}

	userValue, ok := session.Values["user"]

	if !ok {
		http.Error(w, "no user is logged in", http.StatusUnauthorized)
		return SessionUser{}, false
	}

	user, ok := userValue.(SessionUser)
	if !ok {
		http.Error(w, "invalid session user data", http.StatusInternalServerError)
		return SessionUser{}, false
	}

	return user, true
}

func generateRandomString(numBytes int) (string, error) {
	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func generateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
