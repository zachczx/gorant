package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorant/database"
	"gorant/live"
	"gorant/users"

	"github.com/a-h/templ"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Username string
}

const (
	anonymousUserID = "anonymous@rantkit.com"
)

var (
	ctx       context.Context = context.Background()
	emptyUser users.User
)

func main() {
	var err error
	pg := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	database.DB, err = sqlx.Open("pgx", pg)
	if err != nil {
		log.Fatal(err)
	}

	// Init Keycloak client
	k := newKeycloak()
	currentUser := &users.User{SortComments: "upvote;desc"}

	mux := http.NewServeMux()

	// Landing page routes
	mux.Handle("GET /{$}", k.CheckAuthentication()(k.landingHandler()))
	mux.Handle("GET /navbar-profile-badge", k.CheckAuthentication()(k.viewNavbarProfileBadge()))
	mux.HandleFunc("POST /anonymous", viewAnonymousHandler)

	// Post routes
	mux.Handle("GET /posts", k.CheckAuthentication()(k.postsHandler()))
	mux.HandleFunc("POST /filter", postFilterHandler)
	mux.Handle("POST /posts/new", k.CheckAuthentication()(k.newPostHandler()))
	mux.Handle("GET /posts/{postID}", k.CheckAuthentication()(k.viewPostHandler()))
	mux.Handle("POST /posts/{postID}", k.CheckAuthentication()(k.filterSortPostHandler()))
	mux.HandleFunc("GET /posts/{postID}/new", newPostWrongMethodHandler)
	mux.Handle("POST /posts/{postID}/new", k.CheckAuthentication()(k.newCommentHandler()))
	mux.Handle("POST /posts/{postID}/delete", k.CheckAuthentication()(k.deletePostHandler()))
	mux.HandleFunc("GET /posts/{postID}/tags", getTagsHandler)
	mux.Handle("GET /posts/{postID}/tags/edit", k.CheckAuthentication()(k.editTagsHandler()))
	mux.Handle("POST /posts/{postID}/tags/save", k.RequireAuthentication()(k.saveTagsHandler()))
	mux.Handle("POST /posts/{postID}/mood/edit/{newMood}", k.CheckAuthentication()(k.editMoodHandler()))

	// Comment routes
	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", k.CheckAuthentication()(k.upvoteCommentHandler()))
	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit", k.CheckAuthentication()(k.editCommentViewHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/edit", k.CheckAuthentication()(k.editCommentSaveHandler()))
	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit/cancel", k.CheckAuthentication()(k.editCommentCancelHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", k.CheckAuthentication()(k.deleteCommentHandler()))
	mux.Handle("POST /posts/{postID}/description/edit", k.CheckAuthentication()(k.editPostDescriptionHandler()))
	mux.Handle("POST /posts/{postID}/like", k.CheckAuthentication()(k.likePostHandler()))

	// Live
	mux.Handle("GET /live", k.CheckAuthentication()(k.mainLivePageHandler()))
	mux.Handle("POST /live/new", k.CheckAuthentication()(k.newInstantPostHandler()))
	mux.Handle("GET /live/{instPID}", k.CheckAuthentication()(k.viewInstantCommentsHandler()))
	mux.Handle("POST /live/{instPID}/new", k.RequireAuthentication()(k.newInstantCommentHandler()))
	mux.HandleFunc("GET /live/reset-db", func(w http.ResponseWriter, r *http.Request) {
		err := live.ResetDB()
		if err != nil {
			fmt.Println(err)
		}

		w.Write([]byte("Successfully created live DB!"))
	})
	mux.Handle("GET /event/{instPID}", k.CheckAuthentication()(sseHandler()))

	// User and misc routes
	mux.Handle("GET /settings", k.CheckAuthentication()(k.viewSettingsHandler()))
	mux.Handle("POST /settings/edit", k.CheckAuthentication()(k.editSettingsHandler()))
	mux.HandleFunc("GET /error", k.viewErrorHandler)
	mux.HandleFunc("GET /error-unauthorized", k.viewErrorUnauthorizedHandler)
	mux.HandleFunc("GET /admin/reset", resetAdmin)
	mux.Handle("GET /testing", k.OnlyAuthenticated()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected route here!"))
	})))

	// Auth routes
	mux.HandleFunc("GET /login", viewLoginHandler)
	mux.Handle("POST /authenticate", k.LoginHandler(currentUser))
	mux.HandleFunc("GET /register", viewRegisterHandler)
	mux.HandleFunc("POST /register-check-username", registerCheckUsernameHandler)
	mux.HandleFunc("GET /reset-password", viewResetPassword)
	mux.Handle("POST /reset-verification", k.resetPasswordVerificationHandler())
	mux.Handle("POST /registration", k.processRegistrationHandler(currentUser))
	mux.Handle("GET /logout", k.Logout(currentUser))

	// File Server
	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	var p string = os.Getenv("LISTEN_ADDR")
	wrappedMux := StatusLogger(ExcludeCompression(SetCacheControl(mux)))
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}

// For chaining middleware.Iterating in reverse to start from the 2nd argument onwards.
// Each middleware function wraps the next handler. To apply them in the order they are provided in the Compose call,
// we need to start with the innermost wrapper and work our way outwards.
// func chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
// 	for i := len(middleware) - 1; i >= 0; i-- {
// 		h = middleware[i](h)
// 	}
// 	return h
// }
