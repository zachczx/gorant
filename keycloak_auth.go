package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"gorant/database"
	"gorant/templates"
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
	gocloak gocloak.GoCloak
	config  keycloakConfig
	store   *gorillaSessions.CookieStore
}

var regex *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func newKeycloak() *keycloak {
	return &keycloak{
		gocloak: *gocloak.NewClient(os.Getenv("GOCLOAK_URL")),
		config: keycloakConfig{
			clientID:     os.Getenv("GOCLOAK_CLIENT_ID"),
			clientSecret: os.Getenv("GOCLOAK_CLIENT_SECRET"),
			realm:        os.Getenv("GOCLOAK_REALM"),
		},
		store: gorillaSessions.NewCookieStore([]byte(os.Getenv("GORILLA_SESSION_KEY"))),
	}
}

func (k *keycloak) keycloakRegisterHandler(currentUser *users.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if !regex.MatchString(username) {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "Please provide a valid email address."))
			return
		}

		// Get a token for an admin account to create the new account
		// To avoid inserting it carelessly where I don't intend to, I chose not to add the username/password in the keycloak struct.
		adminToken, err := k.gocloak.LoginAdmin(ctx, os.Getenv("GOCLOAK_ADMIN_USER"), os.Getenv("GOCLOAK_ADMIN_PASSWORD"), k.config.realm)
		if err != nil {
			fmt.Println("Error getting admin token!")
			fmt.Println(err)
			http.Error(w, "Error registering!", http.StatusInternalServerError)
			return
		}

		fmt.Println(adminToken.Scope)

		newUser := gocloak.User{
			Username:    gocloak.StringP(username),
			Email:       gocloak.StringP(username),
			Enabled:     gocloak.BoolP(true),
			Credentials: &[]gocloak.CredentialRepresentation{{Type: gocloak.StringP("password"), Value: gocloak.StringP(password), Temporary: gocloak.BoolP(false)}},
		}

		userID, err := k.gocloak.CreateUser(ctx, adminToken.AccessToken, k.config.realm, newUser)
		if err != nil {
			fmt.Println("Error creating user!")
			fmt.Println(err)
			http.Error(w, "Error registering!", http.StatusInternalServerError)
			return
		}

		fmt.Println("Registration successful! UserID: ", userID)

		// Login part

		jwt, err := k.gocloak.Login(ctx, k.config.clientID, k.config.clientSecret, k.config.realm, username, password)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Bad request", http.StatusForbidden)
			return
		}
		fmt.Println("Access Token: ", jwt.AccessToken)
		fmt.Println("Refresh Token: ", jwt.RefreshToken)
		fmt.Println("Expires in: ", jwt.ExpiresIn)

		// Get a session. We're ignoring the error resulted from decoding an existing session:
		//		- Get() always returns a session, even if empty.
		// See: https://github.com/gorilla/sessions
		session, _ := k.store.Get(r, "grumplr_kc_session")
		session.Values["token"] = jwt.AccessToken
		session.Values["username"] = username
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if jwt.AccessToken != "" {
			currentUser.UserID = username
		}

		// Add an entry to the Grumplr DB if the account is new.
		// Then redirect to firstlogin page to configure settings.
		firstLogin, err := SyncUserLocalDB(username)
		if err != nil {
			fmt.Println("Error!! ", err)
		}

		if firstLogin {
			w.Header().Set("HX-Redirect", "/settings?r=firstlogin")
		} else {
			w.Header().Set("HX-Redirect", "/")
		}
	})
}

func (k *keycloak) keycloakLoginHandler(currentUser *users.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Break up login handler from handler and login logic
		// TODO: reuse login logic in registration
		username := r.FormValue("username")
		password := r.FormValue("password")

		if !regex.MatchString(username) {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "Please provide a valid email address."))
			return
		}

		jwt, err := k.gocloak.Login(ctx, k.config.clientID, k.config.clientSecret, k.config.realm, username, password)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Bad request", http.StatusForbidden)
			return
		}
		fmt.Println("Access Token: ", jwt.AccessToken)
		fmt.Println("Refresh Token: ", jwt.RefreshToken)
		fmt.Println("Expires in: ", jwt.ExpiresIn)

		// Get a session. We're ignoring the error resulted from decoding an existing session:
		//		- Get() always returns a session, even if empty.
		// See: https://github.com/gorilla/sessions
		session, _ := k.store.Get(r, "grumplr_kc_session")
		session.Values["token"] = jwt.AccessToken
		session.Values["username"] = username
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if jwt.AccessToken != "" {
			currentUser.UserID = username
		}

		w.Write([]byte("Successfully authenticated!\r\n\r\n"))

		info := fmt.Sprintf("Username: %v\r\n\r\nAccess Token: %s\r\n\r\nRefresh Token: %s\r\n\r\nExpires in: %v", currentUser.UserID, jwt.AccessToken, jwt.RefreshToken, jwt.ExpiresIn)
		w.Write([]byte(info))

		w.Header().Set("HX-Redirect", "/")
	})
}

func (k *keycloak) keycloakResetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")

		if !regex.MatchString(username) {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "Please provide a valid email address."))
			return
		}

		// Get a token for an admin account to create the new account
		adminToken, err := k.gocloak.LoginAdmin(ctx, os.Getenv("GOCLOAK_ADMIN_USER"), os.Getenv("GOCLOAK_ADMIN_PASSWORD"), k.config.realm)
		if err != nil {
			fmt.Println("Error getting admin token!")
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
			return
		}

		params := gocloak.GetUsersParams{Username: &username}

		info, err := k.gocloak.GetUsers(ctx, adminToken.AccessToken, k.config.realm, params)
		if err != nil {
			fmt.Println("Error querying user!")
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
			return
		}

		var keycloakUUID string
		if len(info) > 0 {
			keycloakUUID = *info[0].ID
		}
		fmt.Println(keycloakUUID)

		actions := []string{"UPDATE_PASSWORD"}

		paramsExecute := gocloak.ExecuteActionsEmail{UserID: &keycloakUUID, ClientID: gocloak.StringP(k.config.clientID), Actions: &actions}

		err = k.gocloak.ExecuteActionsEmail(ctx, adminToken.AccessToken, k.config.realm, paramsExecute)
		if err != nil {
			fmt.Println("Error triggering actions email!")
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
			return
		}

		fmt.Println("Successfully triggered reset email!")
	})
}

func (k *keycloak) keycloakCheckAuthentication(currentUser *users.User, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieStart := time.Now()

		session, err := k.store.Get(r, "grumplr_kc_session")
		// Err cannot be nil here since we're verifying token
		if err != nil || session == nil {
			*currentUser = users.User{}
			// http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		cookieSince := time.Since(cookieStart)

		token, ok := session.Values["token"].(string)
		if token == "" || !ok {
			currentUser.UserID = ""
			fmt.Println("No token found!")
			next.ServeHTTP(w, r)
			return
		}

		cookieUsername, ok := session.Values["username"].(string)
		if cookieUsername == "" || !ok {
			fmt.Println("No username cookie found!")
		}
		currentUser.UserID = cookieUsername

		authStart := time.Now()
		result, err := k.gocloak.RetrospectToken(ctx, token, k.config.clientID, k.config.clientSecret, k.config.realm)
		if err != nil || !*result.Active {
			fmt.Println("Token inspection failed!")
			*currentUser = users.User{}
			next.ServeHTTP(w, r)
			return
		}
		authDuration := time.Since(authStart)

		settingsStart := time.Now()

		// Load user settings from cookie or DB.
		// If loaded from DB, then store in cookie to be saved.
		if err := SetSettingsCookie(currentUser, session, cookieUsername); err != nil {
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
					Text:        fmt.Sprintf("UserID: %v", currentUser.UserID),
					TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
					BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
					Bullet:      ">",
				},
				{
					Level:       1,
					Text:        fmt.Sprintf("PreferredName: %v", currentUser.PreferredName),
					TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
					BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
					Bullet:      ">",
				},
				{
					Level:       1,
					Text:        fmt.Sprintf("SortComments: %v", currentUser.SortComments),
					TextStyle:   pterm.NewStyle(pterm.FgLightWhite),
					BulletStyle: pterm.NewStyle(pterm.FgLightWhite),
					Bullet:      ">",
				},
			}
			fmt.Println("###################")
			pterm.DefaultSection.Println("Benchmarks!")
			pterm.DefaultBulletList.WithItems(bulletListItems).Render()
		}

		next.ServeHTTP(w, r)
	})
}

func (k *keycloak) keycloakLogout(currentUser *users.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// invalidate token
		// clear session store
		session, err := k.store.Get(r, "grumplr_kc_session")
		if err != nil {
			fmt.Println("Error getting store: grumplr_kc_session!")
		}

		session.Options.MaxAge = -1
		err = session.Save(r, w)
		if err != nil {
			fmt.Println("Failed to delete grumplr_kc_session", err)
		}

		// clear KUser struct
		*currentUser = users.User{}

		fmt.Println("Successfully logged out!")

		TemplRender(w, r, templates.LoggedOut(currentUser))
	})
}

func SyncUserLocalDB(username string) (bool, error) {
	firstLogin := false
	var exists bool
	err := database.DB.QueryRow("SELECT * FROM users WHERE user_id=$1;", username).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err := database.DB.Exec("INSERT INTO users (user_id, email, preferred_name) VALUES ($1, $2, $3);", username, username, username)
			if err != nil {

				log.Printf("Error inserting new user into DB")
				return firstLogin, err
			}
			fmt.Println("Successfully created new user in DB")
			firstLogin = true
			return firstLogin, nil
		} else {
			fmt.Println("Something else went wrong, not the issue with existing user being found.")
			return firstLogin, err
		}
	}
	fmt.Println("User already exists, no DB action needed")
	return false, nil
}

func SetSettingsCookie(currentUser *users.User, session *gorillaSessions.Session, cookieUsername string) error {
	// Check if cookies are filled, if so, store user pref values in currentUser
	var refetch bool
	var ok bool
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
	currentUser.AvatarPath, ok = session.Values["AvatarPath"].(string)
	if currentUser.AvatarPath == "" || !ok {
		fmt.Println("No AvatarPath cookie found!")
		refetch = true
	}
	currentUser.SortComments, ok = session.Values["SortComments"].(string)
	if currentUser.SortComments == "" || !ok {
		fmt.Println("No SortComments cookie found!")
		refetch = true
	}

	// If cookies are empty, then fetch from DB
	if refetch {
		fmt.Println("Fetching from DB")
		err := currentUser.GetSettings(cookieUsername)
		if err != nil {
			return err
		}

		// Once fetched, store inside cookies
		session.Values["PreferredName"] = currentUser.PreferredName
		session.Values["Avatar"] = currentUser.Avatar
		session.Values["AvatarPath"] = currentUser.AvatarPath
		session.Values["SortComments"] = currentUser.SortComments
	}

	return nil
}
