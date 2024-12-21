package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gorant/database"
	"gorant/posts"
	"gorant/templates"
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

	// Not using this because everything loads so fast, it's just a flash before it changes, which is uglier.
	// And worse, I incur 2 authentication checks instead of 1.
	// It's more troublesome to have to split out Create Bar and NavProfileBadge, both of which needs current user data, just to load them via HTMX separately.
	mux.Handle("GET /navbar-profile-badge", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.NavProfileBadge(currentUser))
	})))

	mux.Handle("GET /{$}", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := posts.ListPosts()
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		t, err := posts.ListTags()
		if err != nil {
			fmt.Println("Error fetching tags", err)
		}
		fmt.Println("Tags: ", t)
		TemplRender(w, r, templates.StarterWelcome(currentUser, p, t))
	})))

	mux.HandleFunc("POST /anonymous", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Hx-Request") == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		t := r.FormValue("post-title")

		TemplRender(w, r, templates.AnonymousMode(t))
	})

	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.Error(currentUser, "Oops something went wrong."))
	})

	mux.HandleFunc("POST /filter", func(w http.ResponseWriter, r *http.Request) {
		// Requires r.ParseForm() because r.FormValue only grabs first value, not other values of same named checkboxes
		r.ParseForm()
		m := r.Form["mood"]
		t := r.Form["tags"]

		// For when there's a reset of the form
		if len(m) == 0 && len(t) == 0 {
			p, err := posts.ListPosts()
			if err != nil {
				fmt.Println("Error fetching posts", err)
			}
			TemplRender(w, r, templates.ListPosts(p))
			return
		}
		// s := r.FormValue("sort")

		fmt.Println("Mood: ", m)
		fmt.Println("Tags: ", t)
		// fmt.Println("Sort: ", s)

		if len(m) == 0 {
			m = []string{"angry", "upset", "sad", "neutral", "happy", "elated"}
		}

		p, err := posts.ListPostsFilter(m, t)
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.ListPosts(p))
	})

	mux.Handle("GET /posts", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("validation") == "error" {
			p, err := posts.ListPosts()
			if err != nil {
				fmt.Println("Error fetching posts", err)
			}

			t, err := posts.ListTags()
			if err != nil {
				fmt.Println("Error fetching tags", err)
			}
			TemplRender(w, r, templates.StarterWelcomeError(currentUser, p, t))
			return
		}
	})))

	mux.Handle("POST /posts/new", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("post-title")
		m := r.FormValue("mood")
		tags := r.FormValue("tags-data")

		exists, ID := posts.VerifyPostID(title)
		if exists {
			TemplRender(w, r, templates.CreatePostError("Post with the same title already exists, please change it."))
			return
		}

		if v := posts.ValidatePost(title); v != nil {
			fmt.Println(v)
			TemplRender(w, r, templates.CreatePostError(v["postTitle"]))
			return
		}
		// TODO need to validate mood as well

		p := posts.ZPost{
			ID:     ID,
			Title:  title,
			UserID: currentUser.UserID,
			Mood:   m,
		}

		var t []string
		if tags != "" {
			t = strings.Split(tags, ",")
		}

		if currentUser.UserID == "" {
			p.UserID = anonymousUserID
		}

		if r.FormValue("anonymous-mode") == "true" {
			if err := posts.NewPost(p, t); err != nil {
				fmt.Println(err)
				w.Header().Set("HX-Redirect", "/error")
				return
			}
			w.Header().Set("HX-Redirect", "/posts/"+ID)
			return
		}

		if err := posts.NewPost(p, t); err != nil {
			fmt.Println(err)
			w.Header().Set("HX-Redirect", "/login?r=new")
			return
		}
		w.Header().Set("HX-Redirect", "/posts/"+ID)
	})))

	mux.Handle("GET /posts/{postID}", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		post, err := posts.GetPost(postID, currentUser.UserID)
		if err != nil {
			fmt.Println("Get post error: ", err)
			TemplRender(w, r, templates.Error(currentUser, "Error!"))
			return
		}

		var filter string

		comments, err := posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(currentUser, "Error!"))
			return
		}

		TemplRender(w, r, templates.Post(currentUser, "Posts", post, comments, "", currentUser.SortComments))
	})))

	mux.Handle("POST /posts/{postID}", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		filter := r.FormValue("f")
		sort := r.FormValue("sort")

		fmt.Println("Form value sort: ", r.FormValue("sort"))
		fmt.Println("Filter value: ", filter)

		// By default the radio buttons aren't checked, so there's no default value when the filter is posted
		if sort != "" {
			s, err := users.SaveSortComments(currentUser.UserID, sort)
			if err != nil {
				fmt.Println(err)
			}

			currentUser.SortComments = s
		}

		comments, err := posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(currentUser, "Error!"))
			return
		}

		TemplRender(w, r, templates.PartialPostNewSorted(currentUser, comments, ""))
	})))

	mux.HandleFunc("GET /posts/{postID}/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET not allowed on this route.")
		http.Redirect(w, r, "/posts/{postID}", http.StatusSeeOther)
	})

	mux.Handle("POST /posts/{postID}/new", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")

		if currentUser.UserID == "" {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err := posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, "")
			if err != nil {
				fmt.Println(err)
				TemplRender(w, r, templates.Error(currentUser, "Error!"))
				return
			}
			TemplRender(w, r, templates.PartialPostNewErrorLogin(currentUser, comments))
			return
		}

		if exists, _ := posts.VerifyPostID(postID); !exists {
			fmt.Println("Error verifying post exists")
			TemplRender(w, r, templates.Error(currentUser, "Error! Post doesn't exist!"))
			return
		}

		c := posts.Comment{
			UserID:  currentUser.UserID,
			Content: r.FormValue("message"),
			PostID:  postID,
		}

		if v := posts.Validate(c); v != nil {
			fmt.Println("Error: ", v)
			comments, err := posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, "")
			if err != nil {
				fmt.Println("Error fetching posts")
				TemplRender(w, r, templates.Error(currentUser, "Oops, something went wrong."))
				return
			}
			TemplRender(w, r, templates.PartialPostNewError(currentUser, comments, v))
			return
		}

		var insertedID string
		insertedID, err := posts.Insert(c)
		if err != nil {
			fmt.Println("Error inserting: ", err)
		}

		comments, err := posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, "")
		if err != nil {
			TemplRender(w, r, templates.Error(currentUser, "Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(currentUser, comments, insertedID))
		}
	})))

	mux.Handle("POST /posts/{postID}/delete", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		if err := posts.DeletePost(postID, currentUser.UserID); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})))

	mux.Handle("GET /posts/{postID}/tags", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		p, err := posts.GetTags(postID)
		if err != nil {
			fmt.Println(err)
		}
		p.ID = postID

		TemplRender(w, r, templates.ShowTags(p))
	}))

	mux.Handle("GET /posts/{postID}/tags/edit", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		post, err := posts.GetPost(postID, currentUser.UserID)
		if err != nil {
			fmt.Println(err)
		}

		TemplRender(w, r, templates.PartialEditTags(post))
	})))

	mux.Handle("POST /posts/{postID}/tags/save", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		t := r.FormValue("tags-data")
		fmt.Println("Form data: ", t)

		var tags []string
		if t != "" {
			tags = strings.Split(t, ",")
		}

		err := posts.EditTags(postID, tags)
		if err != nil {
			fmt.Println(err)
		}

		p, err := posts.GetTags(postID)
		if err != nil {
			fmt.Println(err)
		}
		p.ID = postID

		TemplRender(w, r, templates.ShowTags(p))
	})))

	mux.Handle("POST /posts/{postID}/mood/edit/{newMood}", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		newMood := r.PathValue("newMood")

		if currentUser.UserID == "" {
			post, err := posts.GetPost(postID, currentUser.UserID)
			if err != nil {
				fmt.Println(err)
			}
			TemplRender(w, r, templates.PartialEditMoodError(postID, post.Mood))
			return
		}

		if err := posts.EditMood(postID, newMood); err != nil {
			fmt.Println(err)
			return
		}

		post, err := posts.GetPost(postID, currentUser.UserID)
		if err != nil {
			fmt.Println("Issue with getting post info: ", err)
		}

		TemplRender(w, r, templates.PartialMoodMapper(currentUser, postID, post.UserID, post.Mood))
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")
		if ctx.Value("currentUser") == "" {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You need to login before upvoting."))
			return
		}
		var err error
		err = posts.UpVote(commentID, currentUser.UserID)
		if err != nil {
			fmt.Println("Error executing upvote", err)
		}

		var comments []posts.JoinComment
		comments, err = posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, "")
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(currentUser, comments, commentID))
	})))

	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		c, err := posts.GetComment(commentID, currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEdit(c))
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/edit", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")
		e := r.FormValue("edit-content")

		if err := posts.EditComment(commentID, e, currentUser.UserID); err != nil {
			fmt.Println(err)
			return
		}

		c, err := posts.GetComment(commentID, currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEditSuccess(c))
	})))

	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit/cancel", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		commentID := r.PathValue("commentID")

		c, err := posts.GetComment(commentID, currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEditSuccess(c))
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		if currentUser.UserID == "" {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You can't delete others' comments!"))
			return
		}

		if err := posts.Delete(commentID, currentUser.UserID); err != nil {
			fmt.Println("Error deleting comment: ", err)
			return
		}

		comments, err := posts.ListCommentsFilterSort(postID, currentUser.UserID, currentUser.SortComments, "")
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.PartialPostDelete(currentUser, comments))
	})))

	mux.Handle("POST /posts/{postID}/description/edit", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		description := r.FormValue("post-description-input")

		err := posts.EditPostDescription(postID, description)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Toast("error", "Something went wrong while editing the post!"))
			return
		}

		post, err := posts.GetPost(postID, currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching post info", err)
		}
		TemplRender(w, r, templates.PartialEditDescriptionResponse(currentUser, post))
	})))

	mux.Handle("POST /posts/{postID}/like", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentUser.UserID == "" {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You need to login before liking a post."))
			return
		}
		postID := r.PathValue("postID")
		score, err := posts.LikePost(postID, currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
			return
		}

		if score == 1 {
			TemplRender(w, r, templates.PartialLikePost(postID, "1"))
		} else {
			TemplRender(w, r, templates.PartialLikePost(postID, "0"))
		}
	})))

	mux.HandleFunc("GET /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") == "TRUE" {
			err := database.Reset()
			if err != nil {
				fmt.Println(err)
				w.Write([]byte("Reset failed, errored out"))
				return
			}

			s := time.Now().Format(time.RFC3339)

			TemplRender(w, r, templates.Reset("", s))
		} else {
			w.Write([]byte("Not allowed!"))
		}
	})

	mux.Handle("GET /settings", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("r")
		err := currentUser.GetSettings(currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		switch ref {
		case "firstlogin":
			fmt.Println("in switch")
			TemplRender(w, r, templates.SettingsFirstLogin(currentUser))
			return
		}
		TemplRender(w, r, templates.Settings(currentUser))
	})))

	mux.Handle("POST /settings/edit", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := users.Settings{
			PreferredName: r.FormValue("preferred-name"),
			ContactMe:     r.FormValue("contact-me"),
			Avatar:        r.FormValue("avatar-radio"),
			SortComments:  r.FormValue("sort-comments"),
		}

		if err := users.Validate(f); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			TemplRender(w, r, templates.Toast("error", "Sorry, an error occurred while saving!"))
			return
		}

		if err := users.SaveSettings(currentUser.UserID, f); err != nil {
			fmt.Println("Error saving: ", err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		err := currentUser.GetSettings(currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		session, err := k.store.Get(r, "grumplr_kc_session")
		if err != nil {
			fmt.Println("Failed to access grumplr_kc_session", err)
		}
		session.Values["PreferredName"] = currentUser.PreferredName
		session.Values["Avatar"] = currentUser.Avatar
		session.Values["AvatarPath"] = currentUser.AvatarPath
		session.Values["SortComments"] = currentUser.SortComments
		err = session.Save(r, w)
		if err != nil {
			fmt.Println("Failed to delete grumplr_kc_session", err)
		}

		TemplRender(w, r, templates.PartialSettingsEditSuccess(*currentUser))
	})))

	//--------------------------------------
	// Auth handles
	//--------------------------------------
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		// ref := r.URL.Query().Get("r")

		// switch ref {
		// case "new":
		// 	TemplRender(w, r, templates.Login(currentUser, "error", "You need to login before you can create a new post"))
		// 	return
		// }
		// TemplRender(w, r, templates.Login(currentUser, "", ""))
		TemplRender(w, r, templates.KeycloakLogin(emptyUser))
	})

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	/////////////////////////////////
	// Gocloak
	////////////////////////////////

	mux.Handle("GET /status", k.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Successfully authenticated!\r\n\r\n"))

		info := fmt.Sprintf("Username: %v\r\n\r\n", currentUser.UserID)
		w.Write([]byte(info))
	})))

	mux.Handle("POST /authenticate", k.LoginHandler(currentUser))

	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.KeycloakRegister(emptyUser))
	})

	mux.HandleFunc("GET /reset-password", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.KeycloakResetPassword(emptyUser))
	})

	mux.Handle("POST /reset-verification", k.ResetHandler())

	mux.Handle("POST /registration", k.RegisterHandler(currentUser))

	mux.Handle("GET /logout", k.Logout(currentUser))

	/////////////////////////////////
	// Gocloak
	////////////////////////////////

	var p string = os.Getenv("LISTEN_ADDR")
	wrappedMux := StatusLogger(ExcludeCompression(SetCacheControl(mux)))
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}
