package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorant/database"
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
	mux.Handle("GET /{$}", k.AltCheckAuthentication()(k.landingHandler()))
	mux.Handle("GET /navbar-profile-badge", k.AltCheckAuthentication()(k.viewNavbarProfileBadge()))
	mux.HandleFunc("POST /anonymous", viewAnonymousHandler)

	// Post routes
	mux.Handle("GET /posts", k.AltCheckAuthentication()(k.postsHandler()))
	mux.HandleFunc("POST /filter", postFilterHandler)
	mux.Handle("POST /posts/new", k.AltCheckAuthentication()(k.newPostHandler()))
	mux.Handle("GET /posts/{postID}", k.AltCheckAuthentication()(k.viewPostHandler()))
	mux.Handle("POST /posts/{postID}", k.AltCheckAuthentication()(k.filterSortPostHandler()))
	mux.HandleFunc("GET /posts/{postID}/new", newPostWrongMethodHandler)
	mux.Handle("POST /posts/{postID}/new", k.AltCheckAuthentication()(k.newCommentHandler()))
	mux.Handle("POST /posts/{postID}/delete", k.AltCheckAuthentication()(k.deletePostHandler()))
	mux.HandleFunc("GET /posts/{postID}/tags", getTagsHandler)
	mux.Handle("GET /posts/{postID}/tags/edit", k.AltCheckAuthentication()(k.editTagsHandler()))
	mux.Handle("POST /posts/{postID}/tags/save", k.AltCheckAuthentication()(k.saveTagsHandler()))
	mux.Handle("POST /posts/{postID}/mood/edit/{newMood}", k.AltCheckAuthentication()(k.editMoodHandler()))

	// Comment routes
	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", k.AltCheckAuthentication()(k.upvoteCommentHandler()))
	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit", k.AltCheckAuthentication()(k.editCommentViewHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/edit", k.AltCheckAuthentication()(k.editCommentSaveHandler()))
	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit/cancel", k.AltCheckAuthentication()(k.editCommentCancelHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", k.AltCheckAuthentication()(k.deleteCommentHandler()))
	mux.Handle("POST /posts/{postID}/description/edit", k.AltCheckAuthentication()(k.editPostDescriptionHandler()))
	mux.Handle("POST /posts/{postID}/like", k.AltCheckAuthentication()(k.likePostHandler()))

	// User and misc routes
	mux.Handle("GET /settings", k.AltCheckAuthentication()(k.viewSettingsHandler()))
	mux.Handle("POST /settings/edit", k.AltCheckAuthentication()(k.editSettingsHandler()))
	mux.HandleFunc("GET /error", k.viewErrorHandler)
	mux.HandleFunc("GET /admin/reset", resetAdmin)

	// Auth routes
	mux.HandleFunc("GET /login", viewLoginHandler)
	mux.Handle("POST /authenticate", k.LoginHandler(currentUser))
	mux.HandleFunc("GET /register", viewRegisterHandler)
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
