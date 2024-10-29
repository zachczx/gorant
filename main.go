package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"gostart/posts"
	"gostart/templates"

	"github.com/a-h/templ"
	"github.com/go-swiss/compress"

	_ "modernc.org/sqlite"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.StarterWelcome(""))
	})

	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.StarterWelcome("Error!"))
	})

	mux.HandleFunc("POST /posts", func(w http.ResponseWriter, r *http.Request) {
		postID := r.FormValue("postID")
		http.Redirect(w, r, "/posts/"+postID, http.StatusSeeOther)
	})

	mux.HandleFunc("GET /posts/{id}", func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")
		comments, err := posts.View(postID)
		if err != nil {
			http.Redirect(w, r, "/error", 500)
		}
		TemplRender(w, r, templates.Post("Posts", comments, postID))
	})

	mux.HandleFunc("GET /posts/{id}/new", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts/{id}", http.StatusSeeOther)
	})

	mux.HandleFunc("POST /posts/{id}/new", func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")
		c := posts.Comment{
			UserID:    "1",
			Name:      r.FormValue("name"),
			Content:   r.FormValue("message"),
			CreatedAt: time.Now().String(),
			PostID:    postID,
		}

		if vErr := posts.Validate(c); vErr != nil {
			fmt.Println("Error: ", vErr)
			comments, err := posts.View(postID)
			if err != nil {
				http.Redirect(w, r, "/error", 500)
			}
			TemplRender(w, r, templates.PartialPostNewError(comments, postID, vErr))
			return
		}

		if err := posts.Insert(c); err != nil {
			fmt.Println("Error inserting")
		}
		comments, err := posts.View(postID)
		if err != nil {
			http.Redirect(w, r, "/error", 500)
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(comments, postID))
		}
	})

	mux.HandleFunc("POST /posts/{id}/comment/{commentID}/{voteAction}", func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("id")
		commentID := r.PathValue("commentID")
		voteAction := r.PathValue("voteAction")

		fmt.Println(commentID)

		var err error
		if voteAction == "upvote" {
			err = posts.UpVote(commentID)
		} else if voteAction == "downvote" {
			err = posts.DownVote(commentID)
		}
		if err != nil {
			fmt.Println("Error executing upvote", err)
		}

		var comments []posts.Comment
		comments, err = posts.View(postID)
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(comments, postID))
	})

	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.About())
	})

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	wrappedMux := StatusLogger(compress.Middleware(mux))
	var p string = os.Getenv("LISTEN_ADDR")
	http.ListenAndServe(p, wrappedMux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	c.Render(r.Context(), w)
}
