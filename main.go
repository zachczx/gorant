package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gorant/database"
	"gorant/upload"
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
	ctx       = context.Background()
	emptyUser users.User
)

func main() {
	var err error
	pg := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	database.DB, err = sqlx.Open("pgx", pg)
	if err != nil {
		log.Fatal(err)
	}

	r2 := upload.NewBucketConfig(
		upload.WithStore(os.Getenv("S3_STORE")),
		upload.WithBucketName(os.Getenv("S3_BUCKET_NAME")),
		upload.WithBaseEndpoint(os.Getenv("S3_BASE_ENDPOINT")),
		upload.WithAccessKeyID(os.Getenv("S3_ACCESS_KEY")),
		upload.WithAccessKeySecret(os.Getenv("S3_SECRET_ACCESS_KEY")),
		upload.WithPublicAccessDomain(os.Getenv("S3_PUBLIC_ACCESS_DOMAIN")),
	)

	// Init Keycloak client
	k := newKeycloak()
	currentUser := &users.User{SortComments: "upvote;desc"}

	mux := http.NewServeMux()

	// Landing page routes
	mux.Handle("GET /{$}", k.CheckAuthentication()(k.landingHandler()))
	mux.Handle("GET /navbar-profile-badge", k.CheckAuthentication()(k.viewNavbarProfileBadge()))
	mux.HandleFunc("POST /anonymous", viewAnonymousHandler)
	mux.Handle("GET /random", k.CheckAuthentication()(k.randomPostsHandler()))
	mux.Handle("GET /latest", k.CheckAuthentication()(k.latestPostsHandler()))
	mux.Handle("GET /latest/{p}", k.CheckAuthentication()(k.latestPostsPagesHandler()))
	mux.Handle("GET /about", k.CheckAuthentication()(k.aboutHandler()))
	mux.Handle("GET /yours", k.CheckAuthentication()(k.yourPostsHandler()))
	mux.Handle("GET /yours/{p}", k.CheckAuthentication()(k.yourPostsPagesHandler()))
	mux.HandleFunc("GET /lucky", feelingLuckyHandler)

	// Post routes
	mux.Handle("GET /posts", k.CheckAuthentication()(k.postsHandler()))
	mux.HandleFunc("POST /filter", postFilterHandler)
	mux.Handle("POST /posts/new", k.CheckAuthentication()(k.newPostHandler()))
	mux.Handle("GET /posts/{postID}", k.CheckAuthentication()(k.viewPostHandler()))
	mux.Handle("POST /posts/{postID}", k.CheckAuthentication()(k.filterSortPostHandler()))
	mux.Handle("POST /posts/{postID}/delete", k.CheckAuthentication()(k.deletePostHandler()))
	mux.Handle("GET /posts/{postID}/tags", k.CheckAuthentication()(k.getTagsHandler()))
	mux.Handle("GET /posts/{postID}/tags/edit", k.CheckAuthentication()(k.editTagsHandler()))
	mux.Handle("POST /posts/{postID}/tags/save", k.RequireAuthentication()(k.saveTagsHandler()))
	mux.Handle("POST /posts/{postID}/mood/edit/{newMood}", k.CheckAuthentication()(k.editMoodHandler()))

	// Comment routes
	mux.Handle("POST /posts/{postID}/new", k.RequireAuthentication()(k.newCommentHandler(r2)))
	mux.HandleFunc("GET /posts/{postID}/new", newPostWrongMethodHandler)
	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", k.CheckAuthentication()(k.upvoteCommentHandler()))
	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit", k.CheckAuthentication()(k.editCommentViewHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/edit", k.CheckAuthentication()(k.editCommentSaveHandler(r2)))
	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit/cancel", k.CheckAuthentication()(k.editCommentCancelHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", k.CheckAuthentication()(k.deleteCommentHandler()))
	mux.Handle("POST /posts/{postID}/description/edit", k.CheckAuthentication()(k.editPostDescriptionHandler()))
	mux.Handle("POST /posts/{postID}/like", k.CheckAuthentication()(k.likePostHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/attachment/delete", k.CheckAuthentication()(k.deleteCommentAttachmentHandler(r2)))

	// Reply routes
	mux.Handle("POST /posts/{postID}/comment/{commentID}/reply", k.CheckAuthentication()(k.replyHandler()))
	mux.Handle("POST /posts/{postID}/comment/{commentID}/reply/{replyID}/delete", k.CheckAuthentication()(k.deleteReplyHandler()))

	// Live
	mux.Handle("GET /live", k.CheckAuthentication()(k.mainLivePageHandler()))
	mux.Handle("POST /live/new", k.CheckAuthentication()(k.newInstantPostHandler()))
	mux.Handle("GET /live/{instPID}", k.CheckAuthentication()(k.viewInstantCommentsHandler()))
	mux.Handle("POST /live/{instPID}/new", k.RequireAuthentication()(k.newInstantCommentHandler()))

	// Implemented own compress with Brotli/Gzip, with extra flushing.
	// An alternative that works out of box is klauspost/compress/gzhttp. Others don't. It's because of flushing.
	mux.Handle("GET /event/{instPID}", ZxCompress())

	// User and misc routes
	mux.Handle("GET /profile", k.CheckAuthentication()(k.profileHandler()))
	mux.Handle("GET /profile/posts/{postPage}", k.CheckAuthentication()(k.profilePostsViewMoreHandler()))
	mux.Handle("GET /settings", k.CheckAuthentication()(k.viewSettingsHandler()))
	mux.Handle("POST /settings/edit", k.CheckAuthentication()(k.editSettingsHandler()))
	mux.Handle("GET /credentials", k.CheckAuthentication()(k.viewCredentials()))
	mux.HandleFunc("GET /error", k.viewErrorHandler)
	mux.HandleFunc("GET /error-unauthorized", k.viewErrorUnauthorizedHandler)
	mux.HandleFunc("GET /admin/reset", resetAdmin)

	// Auth routes
	mux.HandleFunc("GET /login", viewLoginHandler)
	mux.Handle("POST /authenticate", k.LoginHandler(currentUser))
	mux.HandleFunc("GET /register", viewRegisterHandler)
	mux.HandleFunc("POST /register-check-username", registerCheckUsernameHandler)
	mux.HandleFunc("GET /reset-password", viewResetPassword)
	mux.Handle("POST /reset-verification", k.resetPasswordVerificationHandler())
	mux.Handle("POST /registration", k.processRegistrationHandler(currentUser))
	mux.Handle("GET /logout", k.Logout(currentUser))

	// Upload routes
	mux.Handle("GET /admin/upload-inspect", k.CheckAuthentication()(k.adminViewUploadHandler(r2)))
	mux.Handle("GET /admin/view/{fileID}", adminViewFileHandler(r2))
	mux.Handle("POST /upload/process", k.CheckAuthentication()(k.adminUploadFileHandler(r2)))
	mux.Handle("POST /admin/upload/test", k.CheckAuthentication()(k.uploadTestFileHandler()))
	mux.Handle("GET /admin/upload/duplicates", k.CheckAuthentication()(k.adminViewDuplicateFilesHandler()))
	mux.Handle("POST /admin/upload/duplicates/delete", k.CheckAuthentication()(k.adminDeleteDuplicateFilesHandler(r2)))

	// Search routes
	mux.Handle("GET /search", k.CheckAuthentication()(k.searchHandler()))

	// File Server
	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	wrappedMux := currentPageContext(StatusLogger(ExcludeCompression(SetCacheControl(mux))))

	server := &http.Server{
		Addr:              os.Getenv("LISTEN_ADDR"),
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           wrappedMux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	if err := c.Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
