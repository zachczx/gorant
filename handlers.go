package main

import (
	"fmt"
	"net/http"
	"strings"

	"gorant/posts"
	"gorant/templates"
	"gorant/users"
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

func (k *keycloak) filterSortPostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		filter := r.FormValue("f")
		sort := r.FormValue("sort")

		fmt.Println("Form value sort: ", r.FormValue("sort"))
		fmt.Println("Filter value: ", filter)

		// By default the radio buttons aren't checked, so there's no default value when the filter is posted
		if sort != "" {
			s, err := users.SaveSortComments(k.currentUser.UserID, sort)
			if err != nil {
				fmt.Println(err)
			}

			k.currentUser.SortComments = s
		}

		comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Error(k.currentUser, "Error!"))
			return
		}

		TemplRender(w, r, templates.PartialPostNewSorted(k.currentUser, comments, ""))
	})
}

func (k *keycloak) newCommentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")

		if k.currentUser.UserID == "" {
			fmt.Println("Not authenticated")
			var comments []posts.Comment
			comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, "")
			if err != nil {
				fmt.Println(err)
				TemplRender(w, r, templates.Error(k.currentUser, "Error!"))
				return
			}
			TemplRender(w, r, templates.PartialPostNewErrorLogin(k.currentUser, comments))
			return
		}

		if exists, _ := posts.VerifyPostID(postID); !exists {
			fmt.Println("Error verifying post exists")
			TemplRender(w, r, templates.Error(k.currentUser, "Error! Post doesn't exist!"))
			return
		}

		c := posts.Comment{
			UserID:  k.currentUser.UserID,
			Content: r.FormValue("message"),
			PostID:  postID,
		}

		if v := posts.Validate(c); v != nil {
			fmt.Println("Error: ", v)
			comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, "")
			if err != nil {
				fmt.Println("Error fetching posts")
				TemplRender(w, r, templates.Error(k.currentUser, "Oops, something went wrong."))
				return
			}
			TemplRender(w, r, templates.PartialPostNewError(k.currentUser, comments, v))
			return
		}

		var insertedID string
		insertedID, err := posts.Insert(c)
		if err != nil {
			fmt.Println("Error inserting: ", err)
		}

		comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, "")
		if err != nil {
			TemplRender(w, r, templates.Error(k.currentUser, "Oops, something went wrong."))
			return
		}
		if hd := r.Header.Get("Hx-Request"); hd != "" {
			TemplRender(w, r, templates.PartialPostNewSuccess(k.currentUser, comments, insertedID))
		}
	})
}

func getTagsHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("postID")
	p, err := posts.GetTags(postID)
	if err != nil {
		fmt.Println(err)
	}
	p.ID = postID

	TemplRender(w, r, templates.ShowTags(p))
}

func (k *keycloak) editTagsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		post, err := posts.GetPost(postID, k.currentUser.UserID)
		if err != nil {
			fmt.Println(err)
		}

		TemplRender(w, r, templates.PartialEditTags(post))
	})
}

func (k *keycloak) saveTagsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}

func (k *keycloak) deletePostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		if err := posts.DeletePost(postID, k.currentUser.UserID); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
