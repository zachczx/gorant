package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
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

var ctx context.Context = context.Background()

func main() {
	service := NewAuthService(
		os.Getenv("STYTCH_PROJECT_ID"),
		os.Getenv("STYTCH_SECRET"),
	)

	mux := http.NewServeMux()
	mux.Handle("/", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := posts.ListPosts()
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		fmt.Println("r.Context()........", r.Context())
		fmt.Println("ctx........", ctx)
		TemplRender(w, r, templates.StarterWelcome(p))
	})))

	//------------------------------------------------------------
	// Tried out lazy loading, but quite jarring with the loaders
	//------------------------------------------------------------

	/* 	mux.HandleFunc("GET /landing", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	   		TemplRender(w, r, templates.Landing())
	   	}))

	   	mux.Handle("POST /posts/list", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	   		// time.Sleep(10 * time.Second)
	   		p, err := posts.ListPosts()
	   		if err != nil {
	   			fmt.Println("Error fetching posts", err)
	   		}

	   		TemplRender(w, r, templates.PartialLandingListPost(p))
	   	})))

	   	mux.Handle("POST /user/list", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	   		// time.Sleep(5 * time.Second)

	   		TemplRender(w, r, templates.PartialNavUser())
	   	}))) */

	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.Error("Oops something went wrong."))
	})

	mux.Handle("/posts", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("validation") == "error" {
			p, err := posts.ListPosts()
			if err != nil {
				fmt.Println("Error fetching posts", err)
			}
			TemplRender(w, r, templates.StarterWelcomeError(p))
			return
		}

		postID := r.FormValue("post-id")

		if v := posts.ValidatePost(postID); v != nil {
			fmt.Println(v)
			http.Redirect(w, r, "/posts?validation=error", http.StatusSeeOther)
			return
		}

		exists := posts.VerifyPostID(postID)
		if exists {
			http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
		}

		if err := posts.NewPost(postID, r.Context().Value("currentUser").(string)); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/login?r=new", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
	})))

	mux.Handle("/posts/{postID}", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		post, err := posts.GetPost(postID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error("Error!"))
			return
		}

		sort := posts.Sort{Type: r.URL.Query().Get("sort"), Direction: r.URL.Query().Get("d")}

		comments, err := posts.NewGetComments(postID, r.Context().Value("currentUser").(string), sort)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error("Error!"))
			return
		}

		if r.Header.Get("Hx-Request") != "" {
			TemplRender(w, r, templates.PartialPostNew(comments, postID, ""))
			return
		}

		TemplRender(w, r, templates.Post("Posts", post, comments, postID, ""))
	})))

	mux.HandleFunc("GET /posts/{postID}/new", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts/{postID}", http.StatusSeeOther)
	})

	mux.Handle("POST /posts/{postID}/new", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")

		if r.Context().Value("currentUser").(string) == "" {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
			if err != nil {
				fmt.Println(err)
				TemplRender(w, r, templates.Error("Error!"))
				return
			}
			TemplRender(w, r, templates.PartialPostNewErrorLogin(comments, postID))
			return
		}

		if !posts.VerifyPostID(postID) {
			fmt.Println("Error verifying post exists")
			TemplRender(w, r, templates.Error("Error! Post doesn't exist!"))
			return
		}

		c := posts.Comment{
			UserID:    r.Context().Value("currentUser").(string),
			Content:   r.FormValue("message"),
			CreatedAt: time.Now().Format(time.RFC3339),
			PostID:    postID,
		}

		if v := posts.Validate(c); v != nil {
			fmt.Println("Error: ", v)
			comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
			if err != nil {
				fmt.Println("Error fetching posts")
				TemplRender(w, r, templates.Error("Oops, something went wrong."))
				return
			}
			TemplRender(w, r, templates.PartialPostNewError(comments, postID, v))
			return
		}

		var insertedID string
		insertedID, err := posts.Insert(c)
		if err != nil {
			fmt.Println("Error inserting: ", err)
		}
		comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
		if err != nil {
			TemplRender(w, r, templates.Error("Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(comments, postID, insertedID))
		}
	})))

	mux.Handle("POST /posts/{postID}/filter", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		q := r.FormValue("query")

		if r.Context().Value("currentUser").(string) == "" {
			fmt.Println("Not authenticated")
			var comments []posts.JoinComment
			comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
			if err != nil {
				fmt.Println(err)
				TemplRender(w, r, templates.Error("Error!"))
				return
			}
			TemplRender(w, r, templates.PartialPostNewErrorLogin(comments, postID))
			return
		}

		comments, err := posts.FilterComments(q, postID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error("Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNew(comments, postID, ""))
		}
	})))

	mux.Handle("POST /posts/{postID}/delete", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		if err := posts.DeletePost(postID, r.Context().Value("currentUser").(string)); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})))

	mux.Handle("POST /posts/{postID}/mood/edit/{newMood}", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		newMood := r.PathValue("newMood")

		if r.Context().Value("currentUser").(string) == "" {
			post, err := posts.GetPost(postID, r.Context().Value("currentUser").(string))
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

		post, err := posts.GetPost(postID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Issue with getting post info: ", err)
		}

		TemplRender(w, r, templates.MoodMapper(postID, post.UserID, post.Mood))
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/upvote", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		if r.Context().Value("currentUser").(string) == "" {
			comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
			if err != nil {
				fmt.Println("Error fetching posts: ", err)
			}
			TemplRender(w, r, templates.PartialPostVoteError(comments, postID))
			return
		}

		var err error
		err = posts.UpVote(commentID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Error executing upvote", err)
		}

		var comments []posts.JoinComment
		comments, err = posts.GetComments(postID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(comments, postID, commentID))
	})))

	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value("currentUser").(string) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		c, err := posts.GetComment(commentID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEdit(c))
	})))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/edit", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value("currentUser").(string) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")
		e := r.FormValue("edit-content")

		if err := posts.EditComment(commentID, e, r.Context().Value("currentUser").(string)); err != nil {
			fmt.Println(err)
			return
		}

		c, err := posts.GetComment(commentID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEditSuccess(c))
	})))

	mux.Handle("GET /posts/{postID}/comment/{commentID}/edit/cancel", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value("currentUser").(string) == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		commentID := r.PathValue("commentID")

		c, err := posts.GetComment(commentID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEditSuccess(c))
	})))

	// comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
	// 	if err != nil {
	// 		fmt.Println("Error fetching posts", err)
	// 	}
	// 	TemplRender(w, r, templates.PartialPostDelete(comments, postID))

	mux.Handle("POST /posts/{postID}/comment/{commentID}/delete", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		if r.Context().Value("currentUser").(string) == "" {
			fmt.Println("I'm inside unauthenticated")
			comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
			if err != nil {
				fmt.Println(err)
			}
			TemplRender(w, r, templates.PartialPostDeleteError(comments, postID))
			return
		}

		if err := posts.Delete(commentID, r.Context().Value("currentUser").(string)); err != nil {
			fmt.Println("Error deleting comment: ", err)
			return
		}

		comments, err := posts.GetComments(postID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.PartialPostDelete(comments, postID))
	})))

	mux.Handle("POST /posts/{postID}/description/edit", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		description := r.FormValue("post-description-input")

		err := posts.EditPostDescription(postID, description)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error("Something went wrong while editing the post!"))
		}

		post, err := posts.GetPost(postID, r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Error fetching post info", err)
		}
		TemplRender(w, r, templates.PartialEditDescriptionResponse(postID, post))
	})))

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.About())
	})

	mux.HandleFunc("GET /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") == "TRUE" {
			err := database.Reset()
			if err != nil {
				fmt.Println(err)
				w.Write([]byte("Reset failed, errored out"))
				return
			}

			t := time.Now().String()

			TemplRender(w, r, templates.Reset("", t))
		} else {
			w.Write([]byte("Not allowed!"))
		}
	})

	mux.Handle("/settings", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("r")
		s, err := users.GetSettings(r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		switch ref {
		case "firstlogin":
			fmt.Println("in switch")
			TemplRender(w, r, templates.SettingsFirstLogin(s))
			return
		}
		TemplRender(w, r, templates.Settings(s))
	})))

	mux.Handle("POST /settings/edit", service.CheckAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := users.Settings{
			PreferredName: r.FormValue("preferred-name"),
			ContactMe:     r.FormValue("contact-me"),
			Avatar:        r.FormValue("avatar-radio"),
		}

		if err := users.Validate(f); err != nil {
			fmt.Println("Error: ", err)
			s, err := users.GetSettings(r.Context().Value("currentUser").(string))
			if err != nil {
				fmt.Println("Error fetching settings: ", err)
				http.Redirect(w, r, "/error", http.StatusSeeOther)
			}
			TemplRender(w, r, templates.PartialSettingsEditError(s))
			return
		}

		if err := users.SaveSettings(r.Context().Value("currentUser").(string), f); err != nil {
			fmt.Println("Error saving: ", err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		s, err := users.GetSettings(r.Context().Value("currentUser").(string))
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		TemplRender(w, r, templates.PartialSettingsEditSuccess(s))
	})))

	//--------------------------------------
	// Auth handles
	//--------------------------------------
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("r")

		switch ref {
		case "new":
			TemplRender(w, r, templates.Login("error", "You need to login before you can create a new post"))
			return
		}
		TemplRender(w, r, templates.Login("", ""))
	})

	mux.HandleFunc("POST /login/sendlink", service.sendMagicLinkHandler)

	mux.Handle("/authenticate", service.authenticateHandler(ctx))

	mux.Handle("/logout", service.logout(ctx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.LoggedOut())
	})))

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	var p string = os.Getenv("LISTEN_ADDR")
	wrappedMux := StatusLogger(ExcludeCompression(mux))
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}
