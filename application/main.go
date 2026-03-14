package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

var (
	oidcProvider    *oidc.Provider
	oauth2Config    oauth2.Config
	idTokenVerifier *oidc.IDTokenVerifier
)

type UserClaims struct {
	Email             string
	PreferredUsername string
	Name              string
}

var currentUser UserClaims

func main() {
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

	fmt.Println("App running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html><body>
		<h2>OIDC app is running.</h2>
		<p><a href="/login">Login with Keycloak</a></p>
		<p><a href="/profile">View profile</a></p>
	</body></html>`)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := oauth2Config.AuthCodeURL("random-state")
	http.Redirect(w, r, url, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("token exchange failed: %v", err)
		http.Error(w, fmt.Sprintf("failed to exchange code for token: %v", err), http.StatusInternalServerError)
		return
	}

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

	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}

	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "failed to parse claims", http.StatusInternalServerError)
		return
	}

	currentUser = UserClaims{
		Email:             claims.Email,
		PreferredUsername: claims.PreferredUsername,
		Name:              claims.Name,
	}

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	if currentUser.Email == "" && currentUser.PreferredUsername == "" && currentUser.Name == "" {
		http.Error(w, "no user is logged in yet", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, `<html><body>
		<h2>Profile</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Username:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><a href="/">Back to home</a></p>
	</body></html>`,
		html.EscapeString(currentUser.Name),
		html.EscapeString(currentUser.PreferredUsername),
		html.EscapeString(currentUser.Email),
	)
}
