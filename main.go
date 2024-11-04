package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"gostart/posts"
	"gostart/templates"

	"github.com/a-h/templ"

	_ "modernc.org/sqlite"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/authentication"
	openid "github.com/zitadel/zitadel-go/v3/pkg/authentication/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

func main() {
	////////////////////////////////////////////////
	// Zitadel
	///////////////////////////////////////////////
	var (
		domain      string = os.Getenv("DOMAIN")
		key         string = os.Getenv("KEY")
		clientID    string = os.Getenv("CLIENT_ID")
		redirectURI string = os.Getenv("REDIRECT_URI")
	)

	flag.Parse()

	ctx := context.Background()

	// Initiate the authentication by providing a zitadel configuration and handler.
	// This example will use OIDC/OAuth2 PKCE Flow, therefore you will also need to initialize that with the generated client_id:
	authN, err := authentication.New(ctx, zitadel.New(domain), key,
		openid.DefaultAuthentication(clientID, redirectURI, key),
	)
	if err != nil {
		slog.Error("zitadel sdk could not initialize", "error", err)
		os.Exit(1)
	}
	// Initialize the middleware by providing the sdk
	mw := authentication.Middleware(authN)

	mux := http.NewServeMux()
	mux.Handle("/auth/", authN)
	// Register the authentication handler on your desired path.
	// It will register the following handlers on it:
	// - /login (starts the authentication process to the Login UI)
	// - /callback (handles the redirect back from the Login UI)
	// - /logout (handles the logout process)

	///////////////////////////////////////////////////
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.StarterWelcome(""))
	})

	mux.Handle("GET /error", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if authentication.IsAuthenticated(r.Context()) {
		// 	http.Redirect(w, r, "/", http.StatusSeeOther)
		// }
		TemplRender(w, r, templates.Error("Oops something went wrong."))
	})))

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		postID := r.FormValue("postID")
		http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
	})

	mux.Handle("/posts/{id}", mw.RequireAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authCtx *openid.UserInfoContext[*oidc.IDTokenClaims, *oidc.UserInfo]
		var u string
		loggedIn := "false"

		if authentication.IsAuthenticated(r.Context()) {
			authCtx = mw.Context(r.Context())
			u = authCtx.UserInfo.PreferredUsername
			loggedIn = "true"
		}
		postID := r.PathValue("id")
		comments, err := posts.View(postID) //
		if err != nil {
			TemplRender(w, r, templates.Error("Error!"))
			return
		}
		TemplRender(w, r, templates.Post("Posts", comments, postID, u, loggedIn))
	})))

	mux.HandleFunc("GET /posts/{id}/new", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts/{id}", http.StatusSeeOther)
	})

	mux.Handle("POST /posts/{id}/new", mw.RequireAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")
		authCtx := mw.Context(r.Context())
		u := authCtx.UserInfo.PreferredUsername

		c := posts.Comment{
			UserID:    u,
			Name:      r.FormValue("name"),
			Content:   r.FormValue("message"),
			CreatedAt: time.Now().String(),
			PostID:    postID,
		}

		if !authCtx.IsAuthenticated() || authCtx.GetUserInfo().PreferredUsername == "" {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err = posts.View(postID)
			TemplRender(w, r, templates.PartialPostVoteError(comments, postID))
			return
		}

		if vErr := posts.Validate(c); vErr != nil {
			fmt.Println("Error: ", vErr)
			comments, err := posts.View(postID)
			if err != nil {
				TemplRender(w, r, templates.Error("Oops, something went wrong."))
				return
			}
			TemplRender(w, r, templates.PartialPostNewError(comments, postID, vErr))
			return
		}

		if err := posts.Insert(c); err != nil {
			fmt.Println("Error inserting")
		}
		comments, err := posts.View(postID)
		if err != nil {
			TemplRender(w, r, templates.Error("Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(comments, postID))
		}
	})))

	mux.Handle("POST /posts/{id}/comment/{commentID}/upvote", mw.RequireAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")
		commentID := r.PathValue("commentID")

		authCtx := mw.Context(r.Context())
		u := authCtx.UserInfo.PreferredUsername

		if !authCtx.IsAuthenticated() || authCtx.GetUserInfo().PreferredUsername == "" {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err = posts.View(postID)
			TemplRender(w, r, templates.PartialPostVoteError(comments, postID))
			return
		}

		var err error

		err = posts.UpVote(commentID, u)
		if err != nil {
			fmt.Println("Error executing upvote", err)
		}

		var comments []posts.JoinComment
		comments, err = posts.View(postID)
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(comments, postID))
	})))

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.About())
	})

	mux.HandleFunc("POST /posts/{id}/comment/{commentID}/delete", func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")
		commentID := r.PathValue("commentID")

		if err := posts.Delete(commentID); err != nil {
			fmt.Println("Error deleting comment: ", err)
			return
		}

		comments, err := posts.View(postID)
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.PartialPostVote(comments, postID))
	})

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	wrappedMux := StatusLogger(ExcludeCompression(mux))
	var p string = os.Getenv("LISTEN_ADDR")
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}
