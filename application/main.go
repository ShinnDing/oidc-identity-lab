package main

import (
	"context"
	"fmt"
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

	if os.Getenv("OIDC_CLIENT_SECRET") == "" {
		log.Fatal("OIDC_CLIENT_SECRET is not set")
	}

	oauth2Config = oauth2.Config{
		ClientID:     "oidc-lab-app",
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  "http://localhost:3000/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)

	fmt.Println("App running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OIDC app is running.")
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
		http.Error(w, "failed to verify ID token", http.StatusInternalServerError)
		return
	}

	_ = idToken

	var claims struct {
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}

	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "failed to parse claims", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Login successful.\n")
	fmt.Fprintf(w, "Name: %s\n", claims.Name)
	fmt.Fprintf(w, "Username: %s\n", claims.PreferredUsername)
	fmt.Fprintf(w, "Email: %s\n", claims.Email)
}
