package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"gorant/database"
	"gorant/users"

	"github.com/Nerzal/gocloak/v13"
	gorillaSessions "github.com/gorilla/sessions"
	"github.com/pterm/pterm"
)

type keycloakConfig struct {
	clientID     string
	clientSecret string
	realm        string
}

type keycloak struct {
	gocloak     gocloak.GoCloak
	config      keycloakConfig
	store       *gorillaSessions.CookieStore
	currentUser *users.User
}

var regex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func newKeycloak() *keycloak {
	return &keycloak{
		gocloak: *gocloak.NewClient(os.Getenv("GOCLOAK_URL")),
		config: keycloakConfig{
			clientID:     os.Getenv("GOCLOAK_CLIENT_ID"),
			clientSecret: os.Getenv("GOCLOAK_CLIENT_SECRET"),
			realm:        os.Getenv("GOCLOAK_REALM"),
		},
		store:       gorillaSessions.NewCookieStore([]byte(os.Getenv("GORILLA_SESSION_KEY"))),
		currentUser: &users.User{SortComments: "upvote;desc"},
	}
}

func (k *keycloak) OnlyAuthenticated() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return k.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if k.currentUser.UserID == "" {
				// w.WriteHeader(http.StatusUnauthorized)
				if r.Header.Get("Hx-Request") != "" {
					fmt.Println("Not authenticated, redirecting to error page.")
					w.Header().Set("Hx-Redirect", "/error-unauthorized")
					return
				}
				fmt.Println("Not authenticated, redirecting to error page.")
				http.Redirect(w, r, "/error-unauthorized", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func (k *keycloak) RequireAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("RequireAuthentication()")
			fmt.Println("UserID: ", k.currentUser.UserID)
			if k.currentUser.UserID == "" {
				// w.WriteHeader(http.StatusUnauthorized)
				if r.Header.Get("Hx-Request") != "" {
					w.Header().Set("Hx-Redirect", "/error")
					return
				}
				http.Redirect(w, r, "/error", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (k *keycloak) CheckAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookieStart := time.Now()
			session, err := k.store.Get(r, "grumplr_kc_session")
			// Err cannot be nil here since we're verifying token
			if err != nil || session == nil {
				*k.currentUser = users.User{}
				// http.Redirect(w, r, "/", http.StatusSeeOther)
				next.ServeHTTP(w, r)
				return
			}
			cookieSince := time.Since(cookieStart)
			token, ok := session.Values["token"].(string)
			if token == "" || !ok {
				k.currentUser.UserID = ""
				fmt.Println("No token found!")
				next.ServeHTTP(w, r)
				return
			}
			cookieUsername, ok := session.Values["username"].(string)
			if cookieUsername == "" || !ok {
				fmt.Println("No username cookie found!")
			}
			k.currentUser.UserID = cookieUsername
			authStart := time.Now()
			result, err := k.gocloak.RetrospectToken(ctx, token, k.config.clientID, k.config.clientSecret, k.config.realm)
			if err != nil || !*result.Active {
				fmt.Println("Token inspection failed!")
				*k.currentUser = users.User{}
				next.ServeHTTP(w, r)
				return
			}
			authDuration := time.Since(authStart)
			settingsStart := time.Now()
			// Load user settings from cookie or DB.
			// If loaded from DB, then store in cookie to be saved.
			if err := SetSettingsCookie(k.currentUser, session, cookieUsername, false); err != nil {
				fmt.Println(err)
			}
			if err := session.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Format benchmarks
			settingsDuration := time.Since(settingsStart)
			if os.Getenv("DEV_ENV") == "TRUE" {
				bulletListItems := []pterm.BulletListItem{
					{
						Level:       0,
						Text:        "Speed",
						TextStyle:   pterm.NewStyle(pterm.FgBlue),
						BulletStyle: pterm.NewStyle(pterm.FgRed),
						Bullet:      " ",
					},
					{
						Level:       1,
						Text:        fmt.Sprintf("Cookie: %v", cookieSince),
						TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
						BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
						Bullet:      ">",
					},
					{
						Level:       1,
						Text:        fmt.Sprintf("Auth: %v", authDuration),
						TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
						BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
						Bullet:      ">",
					},
					{
						Level:       1,
						Text:        fmt.Sprintf("Settings: %v", settingsDuration),
						TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
						BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
						Bullet:      ">",
					},
					{
						Level:       0,
						Text:        "Cookie",
						TextStyle:   pterm.NewStyle(pterm.FgBlue),
						BulletStyle: pterm.NewStyle(pterm.FgRed),
						Bullet:      " ",
					},
					{
						Level:       1,
						Text:        fmt.Sprintf("UserID: %v", k.currentUser.UserID),
						TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
						BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
						Bullet:      ">",
					},
					{
						Level:       1,
						Text:        fmt.Sprintf("PreferredName: %v", k.currentUser.PreferredName),
						TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
						BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
						Bullet:      ">",
					},
					{
						Level:       1,
						Text:        fmt.Sprintf("SortComments: %v", k.currentUser.SortComments),
						TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
						BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
						Bullet:      ">",
					},
				}
				fmt.Println("###################")
				pterm.DefaultSection.Println("Benchmarks!")
				err := pterm.DefaultBulletList.WithItems(bulletListItems).Render()
				if err != nil {
					fmt.Println(err)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SyncUserLocalDB(username string) (bool, error) {
	firstLogin := false
	var exists bool
	err := database.DB.QueryRow("SELECT * FROM users WHERE user_id=$1;", username).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := database.DB.Exec("INSERT INTO users (user_id, email, preferred_name) VALUES ($1, $2, $3);", username, username, username)
			if err != nil {
				return firstLogin, fmt.Errorf("error inserting new user into DB: %w", err)
			}
			fmt.Println("Successfully created new user in DB")
			firstLogin = true
			return firstLogin, nil
		}
		return firstLogin, fmt.Errorf("error: not issue with existing user being found: %w", err)
	}
	fmt.Println("User already exists, no DB action needed")
	return false, nil
}

func SetSettingsCookie(currentUser *users.User, session *gorillaSessions.Session, cookieUsername string, refetch bool) error {
	// Check if cookies are filled, if so, store user pref values in currentUser
	var ok bool
	if !refetch {
		currentUser.PreferredName, ok = session.Values["PreferredName"].(string)
		if currentUser.PreferredName == "" || !ok {
			fmt.Println("No PreferredName cookie found!")
			refetch = true
		}
		currentUser.Avatar, ok = session.Values["Avatar"].(string)
		if currentUser.Avatar == "" || !ok {
			fmt.Println("No Avatar cookie found!")
			refetch = true
		}
		currentUser.SortComments, ok = session.Values["SortComments"].(string)
		if currentUser.SortComments == "" || !ok {
			fmt.Println("No SortComments cookie found!")
			refetch = true
		}
	}

	// If cookies are empty, then fetch from DB
	if refetch {
		fmt.Println("Fetching from DB")
		if err := currentUser.GetSettings(cookieUsername); err != nil {
			return fmt.Errorf("error with getting settings from db: %w", err)
		}
		// Once fetched, store inside cookies
		session.Values["PreferredName"] = currentUser.PreferredName
		session.Values["Avatar"] = currentUser.Avatar
		session.Values["SortComments"] = currentUser.SortComments
	}
	return nil
}
