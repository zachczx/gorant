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

	"github.com/zitadel/zitadel-go/v3/pkg/authentication"
	openid "github.com/zitadel/zitadel-go/v3/pkg/authentication/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type User struct {
	Username string
	LoggedIn string
}

type ZitadelConfig struct {
	Domain      string
	Key         string
	ClientID    string
	RedirectURI string
}

func main() {
	user := &User{Username: "", LoggedIn: "false"}
	////////////////////////////////////////////////
	// Zitadel
	///////////////////////////////////////////////
	z := &ZitadelConfig{os.Getenv("DOMAIN"), os.Getenv("KEY"), os.Getenv("CLIENT_ID"), os.Getenv("REDIRECT_URI")}

	flag.Parse()

	ctx := context.Background()

	// Initiate the authentication by providing a zitadel configuration and handler.
	// This example will use OIDC/OAuth2 PKCE Flow, therefore you will also need to initialize that with the generated client_id:
	authN, err := authentication.New(ctx, zitadel.New(z.Domain), z.Key,
		openid.DefaultAuthentication(z.ClientID, z.RedirectURI, z.Key),
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
	mux.Handle("/", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authentication.IsAuthenticated(r.Context()) {
			fmt.Println("Logged in!")
			user.Username = mw.Context(r.Context()).UserInfo.PreferredUsername
			user.LoggedIn = "true"
		}
		p, err := posts.ListPosts()
		if err != nil {
			fmt.Println("Error fetching posts")
		}

		TemplRender(w, r, templates.StarterWelcome("", p, user.Username, user.LoggedIn))
	})))

	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.Error("Oops something went wrong."))
	})

	mux.Handle("POST /posts", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.FormValue("post-id")
		fmt.Println("Form value received: ", postID)

		if !authentication.IsAuthenticated(r.Context()) {
			exists := posts.VerifyPostID(postID)

			if !exists {
				http.Redirect(w, r, "/error", http.StatusSeeOther)
			}
			http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
			return
		}
		username := mw.Context(r.Context()).UserInfo.PreferredUsername

		err := posts.NewPost(postID, username)
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}
		http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
	})))

	mux.Handle("/posts/{postID}", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authentication.IsAuthenticated(r.Context()) {
			fmt.Println("Logged in!")
			user.Username = mw.Context(r.Context()).UserInfo.PreferredUsername
			user.LoggedIn = "true"
		}

		postID := r.PathValue("postID")
		comments, err := posts.View(postID, user.Username)
		if err != nil {
			TemplRender(w, r, templates.Error("Error!"))
			return
		}

		TemplRender(w, r, templates.Post("Posts", comments, postID, user.Username, user.LoggedIn))
	})))

	mux.HandleFunc("GET /posts/{postID}/new", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts/{postID}", http.StatusSeeOther)
	})

	mux.Handle("POST /posts/{postID}/new", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		authCtx := mw.Context(r.Context())
		user.Username = mw.Context(r.Context()).UserInfo.PreferredUsername

		if !authCtx.IsAuthenticated() {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err = posts.View(postID, user.Username)
			TemplRender(w, r, templates.PartialPostNewErrorLogin(comments, postID, user.Username))
			return
		}

		if !posts.VerifyPostID(postID) {
			fmt.Println("Error verifying post exists")
			TemplRender(w, r, templates.Error("Error! Post doesn't exist!"))
			return
		}

		c := posts.Comment{
			UserID:    user.Username,
			Content:   r.FormValue("message"),
			CreatedAt: time.Now().String(),
			PostID:    postID,
		}

		if vErr := posts.Validate(c); vErr != nil {
			fmt.Println("Error: ", vErr)
			comments, err := posts.View(postID, user.Username)
			if err != nil {
				fmt.Println("Error fetching posts")
				TemplRender(w, r, templates.Error("Oops, something went wrong."))
				return
			}
			TemplRender(w, r, templates.PartialPostNewError(comments, postID, user.Username, vErr))
			return
		}

		if err := posts.Insert(c); err != nil {
			fmt.Println("Error inserting")
		}
		comments, err := posts.View(postID, user.Username)
		if err != nil {
			TemplRender(w, r, templates.Error("Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(comments, postID, user.Username))
		}
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		authCtx := mw.Context(r.Context())
		user.Username = mw.Context(r.Context()).UserInfo.PreferredUsername

		if !authCtx.IsAuthenticated() {
			comments, err := posts.View(postID, user.Username)
			if err != nil {
				fmt.Println("Error fetching posts: ", err)
			}
			TemplRender(w, r, templates.PartialPostVoteError(comments, postID, user.Username))
			return
		}

		var err error

		err = posts.UpVote(commentID, user.Username)
		if err != nil {
			fmt.Println("Error executing upvote", err)
		}

		var comments []posts.JoinComment
		comments, err = posts.View(postID, user.Username)
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(comments, postID, user.Username))
	})))

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.About())
	})

	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", mw.CheckAuthentication()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")
		user.Username = mw.Context(r.Context()).UserInfo.PreferredUsername

		authCtx := mw.Context(r.Context())
		if !authCtx.IsAuthenticated() {
			fmt.Println("I'm inside unauthenticated")
			comments, err := posts.View(postID, user.Username)
			if err != nil {
				fmt.Println(err)
			}
			TemplRender(w, r, templates.PartialPostDeleteError(comments, postID, user.Username))
			return
		}

		if err := posts.Delete(commentID, user.Username); err != nil {
			fmt.Println("Error deleting comment: ", err)
			return
		}

		comments, err := posts.View(postID, user.Username)
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.PartialPostDelete(comments, postID, user.Username))
	})))

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	mux.HandleFunc("/admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") == "TRUE" {
			err := posts.ResetDB()
			if err != nil {
				w.Write([]byte("Reset failed, errored out"))
				return
			}

			t := time.Now().String()

			TemplRender(w, r, templates.Reset("", t))
		}
	})

	wrappedMux := StatusLogger(ExcludeCompression(mux))
	var p string = os.Getenv("LISTEN_ADDR")
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}
