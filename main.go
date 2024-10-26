package main

import (
	"fmt"
	"net/http"

	"gostart/posts"
	"gostart/templates"

	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func main() {
	var p string = ":7000"

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.StarterWelcome("Hello world!"))
	})

	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.StarterWelcome("Error!"))
	})

	mux.HandleFunc("GET /posts", func(w http.ResponseWriter, r *http.Request) {
		posts, err := posts.View()
		if err != nil {
			http.Redirect(w, r, "/error", 500)
		}
		TemplRender(w, r, templates.Post("Posts", posts))
	})

	mux.HandleFunc("GET /posts/new", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	})

	mux.HandleFunc("POST /posts/new", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		msg := r.FormValue("message")
		fmt.Println(name)
		fmt.Println(msg)
		if err := posts.Insert(name, msg); err != nil {
			fmt.Println("Error inserting")
		}
		posts, err := posts.View()
		if err != nil {
			http.Redirect(w, r, "/error", 500)
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNew(posts))
		}
	})

	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(p, mux)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	posts.Connect()
	c.Render(r.Context(), w)
}
