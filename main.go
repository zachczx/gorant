package main

import (
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.StarterWelcome("Hello world!"))
	})
	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		posts := posts.View()

		TemplRender(w, r, templates.Post("Posts", posts))
	})
	http.HandleFunc("/posts/insert", func(w http.ResponseWriter, r *http.Request) {
		posts.Insert()
	})
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(p, nil)
}

func TemplRender(w http.ResponseWriter, r *http.Request, c templ.Component) {
	posts.Connect()
	c.Render(r.Context(), w)
}
