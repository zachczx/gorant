package main

import (
	"fmt"
	"net/http"
	"os"

	"gorant/templates"
	"gorant/users"

	"github.com/Nerzal/gocloak/v13"
)

func (k *keycloak) processRegistrationHandler(currentUser *users.User) http.Handler {
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
			TemplRender(w, r, templates.Toast("error", "Error processing your registration!"))
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

func registerCheckUsernameHandler(w http.ResponseWriter, r *http.Request) {
	u := r.FormValue("username")
	exists, err := users.CheckUsername(u)
	fmt.Println(u, "...", exists, "...", err)
	if err != nil {
		TemplRender(w, r, templates.CheckUsernameMessage("empty"))
		return
	}

	if exists {
		TemplRender(w, r, templates.CheckUsernameMessage("exists"))
	} else if !exists {
		TemplRender(w, r, templates.CheckUsernameMessage("avail"))
	}
}

func (k *keycloak) LoginHandler(currentUser *users.User) http.Handler {
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
			w.WriteHeader(http.StatusUnauthorized)
			TemplRender(w, r, templates.InvalidUsernameOrPasswordMessage())
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

		TemplRender(w, r, templates.SuccessfulLoginMessage())
	})
}

func viewResetPassword(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.KeycloakResetPassword(emptyUser))
}

func (k *keycloak) resetPasswordVerificationHandler() http.Handler {
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

func viewRegisterHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.KeycloakRegister(emptyUser))
}

func (k *keycloak) Logout(currentUser *users.User) http.Handler {
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

func viewLoginHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.KeycloakLogin(emptyUser))
}
