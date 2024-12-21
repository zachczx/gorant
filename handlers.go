package main

import (
	"fmt"
	"net/http"
	"strings"

	"gorant/posts"
	"gorant/templates"
)

func (k *keycloak) landingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := posts.ListPosts()
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		t, err := posts.ListTags()
		if err != nil {
			fmt.Println("Error fetching tags", err)
		}
		fmt.Println("Tags: ", t)
		TemplRender(w, r, templates.StarterWelcome(k.currentUser, p, t))
	})
}

func postFilterHandler(w http.ResponseWriter, r *http.Request) {
	// Requires r.ParseForm() because r.FormValue only grabs first value, not other values of same named checkboxes
	r.ParseForm()
	m := r.Form["mood"]
	t := r.Form["tags"]

	fmt.Println("Mood: ", m)
	fmt.Println("Tags: ", t)

	p, err := posts.ListPostsFilter(m, t)
	if err != nil {
		fmt.Println("Error fetching posts", err)
	}

	TemplRender(w, r, templates.ListPosts(p))
}

/*
func (k *keycloak) xxx() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
*/

func (k *keycloak) postsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("validation") == "error" {
			p, err := posts.ListPosts()
			if err != nil {
				fmt.Println("Error fetching posts", err)
			}

			t, err := posts.ListTags()
			if err != nil {
				fmt.Println("Error fetching tags", err)
			}
			TemplRender(w, r, templates.StarterWelcomeError(k.currentUser, p, t))
			return
		}
	})
}

func (k *keycloak) newPostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			UserID: k.currentUser.UserID,
			Mood:   m,
		}

		var t []string
		if tags != "" {
			t = strings.Split(tags, ",")
		}

		if k.currentUser.UserID == "" {
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
	})
}

func (k *keycloak) viewPostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		post, err := posts.GetPost(postID, k.currentUser.UserID)
		if err != nil {
			fmt.Println("Get post error: ", err)
			TemplRender(w, r, templates.Error(k.currentUser, "Error!"))
			return
		}

		var filter string

		comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(k.currentUser, "Error!"))
			return
		}

		TemplRender(w, r, templates.Post(k.currentUser, "Posts", post, comments, "", k.currentUser.SortComments))
	})
}
