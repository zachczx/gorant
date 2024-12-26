package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gorant/database"
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
		TemplRender(w, r, templates.MainPage(k.currentUser, p, t))
	})
}

// Not using this because everything loads so fast, it's just a flash before it changes, which is uglier.
// And worse, I incur 2 authentication checks instead of 1.
// It's more troublesome to have to split out Create Bar and NavProfileBadge, both of which needs current user data, just to load them via HTMX separately.
func (k *keycloak) viewNavbarProfileBadge() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TemplRender(w, r, templates.NavProfileBadge(k.currentUser))
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

func viewAnonymousHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Hx-Request") == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	t := r.FormValue("post-title")

	TemplRender(w, r, templates.AnonymousMode(t))
}

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
			TemplRender(w, r, templates.MainPageError(k.currentUser, p, t))
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

		p := posts.Post{
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
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You need to login or post in anonymous mode!"))
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

func newPostWrongMethodHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET not allowed on this route.")
	http.Redirect(w, r, "/posts/{postID}", http.StatusSeeOther)
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

func (k *keycloak) likePostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if k.currentUser.UserID == "" {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You need to login before liking a post."))
			return
		}
		postID := r.PathValue("postID")
		score, err := posts.LikePost(postID, k.currentUser.UserID)
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

func (k *keycloak) editMoodHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		newMood := r.PathValue("newMood")

		if k.currentUser.UserID == "" {
			post, err := posts.GetPost(postID, k.currentUser.UserID)
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

		post, err := posts.GetPost(postID, k.currentUser.UserID)
		if err != nil {
			fmt.Println("Issue with getting post info: ", err)
		}

		TemplRender(w, r, templates.PartialMoodMapper(k.currentUser, postID, post.UserID, post.Mood))
	})
}

func (k *keycloak) upvoteCommentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")
		if ctx.Value("currentUser") == "" {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You need to login before upvoting."))
			return
		}
		var err error
		err = posts.UpVote(commentID, k.currentUser.UserID)
		if err != nil {
			fmt.Println("Error executing upvote", err)
		}

		var comments []posts.Comment
		comments, err = posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, "")
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}

		TemplRender(w, r, templates.PartialPostVote(k.currentUser, comments, commentID))
	})
}

func (k *keycloak) editCommentViewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if k.currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		c, err := posts.GetComment(commentID, k.currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEdit(c))
	})
}

func (k *keycloak) editCommentSaveHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if k.currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")
		e := r.FormValue("edit-content")

		if err := posts.EditComment(commentID, e, k.currentUser.UserID); err != nil {
			fmt.Println(err)
			return
		}

		c, err := posts.GetComment(commentID, k.currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEditSuccess(c))
	})
}

func (k *keycloak) editCommentCancelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if k.currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		commentID := r.PathValue("commentID")

		c, err := posts.GetComment(commentID, k.currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}

		TemplRender(w, r, templates.PartialCommentEditSuccess(c))
	})
}

func (k *keycloak) deleteCommentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		commentID := r.PathValue("commentID")

		if k.currentUser.UserID == "" {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "You can't delete others' comments!"))
			return
		}

		if err := posts.Delete(commentID, k.currentUser.UserID); err != nil {
			fmt.Println("Error deleting comment: ", err)
			return
		}

		comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, "")
		if err != nil {
			fmt.Println("Error fetching posts", err)
		}
		TemplRender(w, r, templates.PartialPostDelete(k.currentUser, comments))
	})
}

func (k *keycloak) editPostDescriptionHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")
		description := r.FormValue("post-description-input")

		err := posts.EditPostDescription(postID, description)
		if err != nil {
			fmt.Println(err)
			TemplRender(w, r, templates.Toast("error", "Something went wrong while editing the post!"))
			return
		}

		post, err := posts.GetPost(postID, k.currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching post info", err)
		}
		TemplRender(w, r, templates.PartialEditDescriptionResponse(k.currentUser, post))
	})
}

func (k *keycloak) viewSettingsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ref := r.URL.Query().Get("r")
		err := k.currentUser.GetSettings(k.currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
		a := users.ReturnAvatars()

		switch ref {
		case "firstlogin":
			fmt.Println("in switch")
			TemplRender(w, r, templates.SettingsFirstLogin(k.currentUser, a))
			return
		}
		TemplRender(w, r, templates.Settings(k.currentUser, a))
	})
}

func (k *keycloak) editSettingsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if err := users.SaveSettings(k.currentUser.UserID, f); err != nil {
			fmt.Println("Error saving: ", err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		err := k.currentUser.GetSettings(k.currentUser.UserID)
		if err != nil {
			fmt.Println("Error fetching settings: ", err)
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		session, err := k.store.Get(r, "grumplr_kc_session")
		if err != nil {
			fmt.Println("Failed to access grumplr_kc_session", err)
		}
		session.Values["PreferredName"] = k.currentUser.PreferredName
		session.Values["Avatar"] = k.currentUser.Avatar
		session.Values["AvatarPath"] = k.currentUser.AvatarPath
		session.Values["SortComments"] = k.currentUser.SortComments
		err = session.Save(r, w)
		if err != nil {
			fmt.Println("Failed to delete grumplr_kc_session", err)
		}

		TemplRender(w, r, templates.PartialSettingsEditSuccess(*k.currentUser))
	})
}

func (k *keycloak) viewErrorHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.Error(k.currentUser, "Oops, something went wrong."))
}

func (k *keycloak) viewErrorUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.ErrorUnauthorized(k.currentUser, "You'll need to login to do that."))
}

func resetAdmin(w http.ResponseWriter, r *http.Request) {
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
}
