package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gorant/database"
	"gorant/posts"
	"gorant/templates"
	"gorant/users"

	"github.com/a-h/templ"

	_ "modernc.org/sqlite"
)

type User struct {
	Username string
}

var (
	ctx       context.Context = context.Background()
	emptyUser users.User
)

func main() {
	service := NewAuthService(
		os.Getenv("STYTCH_PROJECT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)

	currentUser := &users.User{SortComments: "upvote;desc"}

	mux := http.NewServeMux()
	mux.Handle("GET /{$}", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Current user: ", currentUser.UserID)
		p, err := posts.ListPosts()
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.StarterWelcome(*currentUser, p))
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
		TemplRender(w, r, templates.Error(*currentUser, "Oops something went wrong."))
	})

	mux.Handle("GET /posts", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("validation") == "error" {
			p, err := posts.ListPosts()
			if err != nil {
				fmt.Println("Error fetching posts", err)
			}
			TemplRender(w, r, templates.StarterWelcomeError(*currentUser, p))
			return
		}
	})))

	mux.Handle("POST /posts/new", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		p := posts.Post{
			PostID:    ID,
			PostTitle: title,
			UserID:    currentUser.UserID,
			Mood:      m,
		}

		t := strings.Split(tags, ",")

		if currentUser.UserID == "" {
			p.UserID = os.Getenv("ANON_USER_ID")
		}

		if r.FormValue("anonymous-mode") == "true" {
			err := posts.NewPost(p, t)
			if err != nil {
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

	mux.Handle("GET /posts/{postID}", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		post, err := posts.GetPost(postID, currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(*currentUser, "Error!"))
			return
		}

		var filter string

		comments, err := posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(*currentUser, "Error!"))
			return
		}

		TemplRender(w, r, templates.Post(*currentUser, "Posts", post, comments, postID, "", currentUser.SortComments))
	})))

	mux.Handle("POST /posts/{postID}", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		comments, err := posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(*currentUser, "Error!"))
			return
		}

		TemplRender(w, r, templates.PartialPostNewSorted(*currentUser, comments, postID, ""))
	})))

	mux.HandleFunc("GET /posts/{postID}/new", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET not allowed on this route.")
		http.Redirect(w, r, "/posts/{postID}", http.StatusSeeOther)
	})

	mux.Handle("POST /posts/{postID}/new", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")

		if currentUser.UserID == "" {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err := posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, "")
			if err != nil {
				fmt.Println(err)
				TemplRender(w, r, templates.Error(*currentUser, "Error!"))
				return
			}
			TemplRender(w, r, templates.PartialPostNewErrorLogin(*currentUser, comments, postID))
			return
		}

		if exists, _ := posts.VerifyPostID(postID); !exists {
			fmt.Println("Error verifying post exists")
			TemplRender(w, r, templates.Error(*currentUser, "Error! Post doesn't exist!"))
			return
		}

		c := posts.Comment{
			UserID:    currentUser.UserID,
			Content:   r.FormValue("message"),
			CreatedAt: time.Now().Format(time.RFC3339),
			PostID:    postID,
		}

		if v := posts.Validate(c); v != nil {
			fmt.Println("Error: ", v)
			comments, err := posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, "")
			if err != nil {
				fmt.Println("Error fetching posts")
				TemplRender(w, r, templates.Error(*currentUser, "Oops, something went wrong."))
				return
			}
			TemplRender(w, r, templates.PartialPostNewError(*currentUser, comments, postID, v))
			return
		}

		var insertedID string
		insertedID, err := posts.Insert(c)
		if err != nil {
			fmt.Println("Error inserting: ", err)
		}

		comments, err := posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, "")
		if err != nil {
			TemplRender(w, r, templates.Error(*currentUser, "Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(*currentUser, comments, postID, insertedID))
		}
	})))

	mux.Handle("POST /posts/{postID}/delete", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		if err := posts.DeletePost(postID, currentUser.UserID); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})))

	mux.Handle("POST /posts/{postID}/mood/edit/{newMood}", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		TemplRender(w, r, templates.PartialMoodMapper(*currentUser, postID, post.UserID, post.Mood))
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		comments, err = posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, "")
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(*currentUser, comments, postID, commentID))
	})))

	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	mux.Handle("POST /posts/{postID}/comment/{commentID}/edit", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit/cancel", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		comments, err := posts.FilterSortComments(postID, currentUser.UserID, currentUser.SortComments, "")
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.PartialPostDelete(*currentUser, comments, postID))
	})))

	mux.Handle("POST /posts/{postID}/description/edit", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		TemplRender(w, r, templates.PartialEditDescriptionResponse(*currentUser, postID, post))
	})))

	mux.Handle("POST /posts/{postID}/like", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.About(*currentUser))
	})

	mux.HandleFunc("GET /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") == "TRUE" {
			err := database.Reset()
			if err != nil {
				fmt.Println(err)
				w.Write([]byte("Reset failed, errored out"))
				return
			}

			t := time.Now().Format(time.RFC3339)

			TemplRender(w, r, templates.Reset("", t))
		} else {
			w.Write([]byte("Not allowed!"))
		}
	})

	mux.Handle("GET /settings", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("r")
		err := currentUser.GetSettings(currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		switch ref {
		case "firstlogin":
			fmt.Println("in switch")
			TemplRender(w, r, templates.SettingsFirstLogin(*currentUser))
			return
		}
		TemplRender(w, r, templates.Settings(*currentUser))
	})))

	mux.Handle("POST /settings/edit", service.CheckAuthentication(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := users.Settings{
			PreferredName: r.FormValue("preferred-name"),
			ContactMe:     r.FormValue("contact-me"),
			Avatar:        r.FormValue("avatar-radio"),
			SortComments:  r.FormValue("sort-comments"),
		}

		if err := users.Validate(f); err != nil {
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

		TemplRender(w, r, templates.PartialSettingsEditSuccess(*currentUser))
	})))

	//--------------------------------------
	// Auth handles
	//--------------------------------------
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("r")

		switch ref {
		case "new":
			TemplRender(w, r, templates.Login(*currentUser, "error", "You need to login before you can create a new post"))
			return
		}
		TemplRender(w, r, templates.Login(*currentUser, "", ""))
	})

	mux.HandleFunc("POST /login/sendlink", service.sendMagicLinkHandler)

	mux.Handle("GET /authenticate", service.authenticateHandler(*currentUser))

	mux.Handle("GET /logout", service.logout(currentUser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.LoggedOut(*currentUser))
	})))

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	var p string = os.Getenv("LISTEN_ADDR")
	wrappedMux := StatusLogger(ExcludeCompression(mux))
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}
