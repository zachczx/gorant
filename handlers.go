package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"gorant/database"
	"gorant/posts"
	"gorant/templates"
	"gorant/upload"
	"gorant/users"

	"github.com/google/uuid"
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
	if err := r.ParseForm(); err != nil {
		w.Header().Set("Hx-Redirect", "/error")
		return
	}
	m := r.Form["mood"]
	t := r.Form["tags"]
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

func viewFileHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := bc.PublicAccessDomain
		fileID := r.PathValue("fileID")
		TemplRender(w, r, templates.ViewFile(domain, fileID))
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
			fmt.Println("Inside saving")
			s, err := users.SaveSortComments(k.currentUser.UserID, sort)
			if err != nil {
				fmt.Println(err)
			}
			k.currentUser.SortComments = s
			if err := k.UpdateSessionStore(w, r); err != nil {
				fmt.Println(err)
				w.Header().Set("Hx-Redirect", "/error")
				return
			}
		}
		comments, err := posts.ListCommentsFilterSort(postID, k.currentUser.UserID, k.currentUser.SortComments, filter)
		if err != nil {
			fmt.Println(err)
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		TemplRender(w, r, templates.PartialPostNewSorted(k.currentUser, comments, ""))
	})
}

func (k *keycloak) UpdateSessionStore(w http.ResponseWriter, r *http.Request) error {
	// get session from gorilla sessions
	session, err := k.store.Get(r, "grumplr_kc_session")
	// Err cannot be nil here since we're verifying token
	if err != nil || session == nil {
		*k.currentUser = users.User{}
		return fmt.Errorf("error with getting gorilla session store: %w", err)
	}
	if err := SetSettingsCookie(k.currentUser, session, k.currentUser.UserID, true); err != nil {
		return fmt.Errorf("error with setting settings cookie: %w", err)
	}
	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("error with saving gorilla session: %w", err)
	}
	return nil
}

func (k *keycloak) newCommentHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID := r.PathValue("postID")

		if exists, _ := posts.VerifyPostID(postID); !exists {
			fmt.Println("Error verifying post exists")
			TemplRender(w, r, templates.Error(k.currentUser, "Error! Post doesn't exist!"))
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // (32 * 2^20) + 1024 bytes
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "An error occurred!"))
			return
		}
		// Not using r.MultipartForm, because I've only 1 file for 1 input field. If I use r.MultipartForm, I'd need to do
		// mpf.File["upload"][0].Filename, mpf.File["upload"][0].Open() etc.

		var c posts.Comment
		uploadedFile, fileName, thumbnailFileName, uniqueKey, err := k.uploaderHandler(r, bc)
		if err != nil {
			fmt.Println(err)
			if err.Error() == "formfile error: http: no such file" || err.Error() == "error empty file: empty file" {
				c = posts.Comment{
					UserID:  k.currentUser.UserID,
					Content: r.FormValue("message"),
					PostID:  postID,
				}
			} else if err.Error() == "filetype not allowed" {
				w.WriteHeader(http.StatusForbidden)
				TemplRender(w, r, templates.Toast("error", "Uploaded file type not allowed!"))
				return
			} else {
				w.Header().Set("Hx-Redirect", "/error")
				return
			}
		} else {
			// File was uploaded
			c = posts.Comment{
				UserID:  k.currentUser.UserID,
				Content: r.FormValue("message"),
				PostID:  postID,
				File: upload.LookupFile{
					ID:           uniqueKey,
					File:         uploadedFile,
					Key:          fileName,
					ThumbnailKey: thumbnailFileName,
					Store:        bc.Store,
					Bucket:       bc.BucketName,
					BaseURL:      bc.PublicAccessDomain,
				},
			}
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
		insertedID, err = posts.Insert(c)
		if err != nil {
			TemplRender(w, r, templates.Error(k.currentUser, "Oops, something went wrong."))
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

func (k *keycloak) uploaderHandler(r *http.Request, bc *upload.BucketConfig) (multipart.File, string, string, uuid.UUID, error) {
	var uniqueKey uuid.UUID
	var fileName string
	var thumbnailFileName string
	uploadedFile, header, err := r.FormFile("file")
	if err != nil {
		return uploadedFile, fileName, thumbnailFileName, uniqueKey, fmt.Errorf("formfile error: %w", err)
	}
	if header.Size == 0 {
		return uploadedFile, fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error empty file: %w", err)
	}

	defer uploadedFile.Close()
	fileType, err := checkFileType(uploadedFile)
	if err != nil {
		return uploadedFile, fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error checkfiletype(): %w", err)
	}
	fileName, thumbnailFileName, uniqueKey, err = bc.UploadToBucket(uploadedFile, header.Filename, fileType)
	if err != nil {
		return uploadedFile, fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error UploadToBucket(): %w", err)
	}
	return uploadedFile, fileName, thumbnailFileName, uniqueKey, nil
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

func (k *keycloak) editCommentSaveHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if k.currentUser.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		postID := r.PathValue("postID")
		commentID, err := uuid.Parse(r.PathValue("commentID"))
		if err != nil {
			w.Header().Set("Hx-Redirect", "/error")
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // (32 * 2^20) + 1024 bytes
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "An error occurred!"))
			return
		}
		var c posts.Comment
		uploadedFile, fileName, thumbnailFileName, uniqueKey, err := k.uploaderHandler(r, bc)
		if err != nil {
			fmt.Println(err)
			if err.Error() == "http: no such file" || err.Error() == "empty file" {
				c = posts.Comment{
					ID:      commentID,
					UserID:  k.currentUser.UserID,
					Content: r.FormValue("message"),
					PostID:  postID,
				}
			} else if err.Error() == "filetype not allowed" {
				w.WriteHeader(http.StatusForbidden)
				TemplRender(w, r, templates.Toast("error", "Uploaded file type not allowed!"))
				return
			} else {
				w.Header().Set("Hx-Redirect", "/error")
				return
			}
		} else {
			// File was uploaded
			c = posts.Comment{
				ID:      commentID,
				UserID:  k.currentUser.UserID,
				Content: r.FormValue("message"),
				PostID:  postID,
				File: upload.LookupFile{
					ID:           uniqueKey,
					File:         uploadedFile,
					Key:          fileName,
					ThumbnailKey: thumbnailFileName,
					Store:        bc.Store,
					Bucket:       bc.BucketName,
					BaseURL:      bc.PublicAccessDomain,
				},
			}
		}
		if err := posts.EditComment(c); err != nil {
			fmt.Println(err)
			return
		}
		c, err = posts.GetComment(r.PathValue("commentID"), k.currentUser.UserID)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(c)
		TemplRender(w, r, templates.PartialCommentEditSuccess(k.currentUser, c))
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

		TemplRender(w, r, templates.PartialCommentEditSuccess(k.currentUser, c))
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

func (k *keycloak) deleteCommentAttachmentHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commentID := r.PathValue("commentID")
		// Verify if currentUser is owner of comment
		c, err := posts.GetComment(commentID, k.currentUser.UserID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if !c.NullFile.ID.Valid {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		files := []upload.BucketFile{{Key: c.NullFile.Key.String}}
		if err := bc.DeleteBucketFiles(files); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err := upload.DeleteDBFileRecord(c.NullFile.Key.String); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// Get the comment again, since the deletion would have cascaded onto the comments table row.
		c, err = posts.GetComment(commentID, k.currentUser.UserID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		TemplRender(w, r, templates.PartialAttachmentDeleteSuccess(c, bc.PublicAccessDomain))
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

func (k *keycloak) viewUploadHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") != "TRUE" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		f, err := bc.ListBucket()
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}
		TemplRender(w, r, templates.UploadAdmin("testing upload", k.currentUser, f))
	})
}

// TODO: Rename this, since this only seems to be used in admin
func (k *keycloak) uploadFileHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Start uploading")
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // (32 * 2^20) + 1024 bytes
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.WriteHeader(http.StatusForbidden)
			TemplRender(w, r, templates.Toast("error", "An error occurred!"))
			return
		}
		// Not using r.MultipartForm, because I've only 1 file for 1 input field. If I use r.MultipartForm, I'd need to do
		// mpf.File["upload"][0].Filename, mpf.File["upload"][0].Open() etc.

		uploadedFile, header, err := r.FormFile("upload")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer uploadedFile.Close()
		fileType, err := checkFileType(uploadedFile)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		fileName, thumbnailFileName, uniqueKey, err := bc.UploadToBucket(uploadedFile, header.Filename, fileType)
		if err != nil {
			fmt.Println("Upload issue!!! ", err)
		}
		fmt.Println(uniqueKey, thumbnailFileName)
		if r.Header.Get("Hx-Request") == "" {
			TemplRender(w, r, templates.UploadAdmin("testing upload", k.currentUser, nil))
			return
		}

		TemplRender(w, r, templates.SuccessfulUpload(fileName))
	})
}

func (k *keycloak) uploadTestFileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Start uploading")
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // (32 * 2^20) + 1024 bytes
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		// Not using r.MultipartForm, because I've only 1 file for 1 input field. If I use r.MultipartForm, I'd need to do
		// mpf.File["upload"][0].Filename, mpf.File["upload"][0].Open() etc.
		uploadedFile, _, err := r.FormFile("upload")
		if err != nil {
			fmt.Println("form file error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer uploadedFile.Close()
		_, err = checkFileType(uploadedFile)
		if err != nil {
			fmt.Println("check file type", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		fileName, err := upload.ToLocalWebp(uploadedFile)
		if err != nil {
			fmt.Println("Upload issue!!! ", err)
		}
		if r.Header.Get("Hx-Request") == "" {
			TemplRender(w, r, templates.UploadAdmin("testing upload", k.currentUser, nil))
			return
		}
		TemplRender(w, r, templates.SuccessfulTestUpload(fileName))
	})
}

var mimeTypes = []string{"image/png", "image/jpeg", "image/webp", "image/avif", "image/gif"}

func checkFileType(file multipart.File) (string, error) {
	var fileType string
	// Peek into first 512 bytes to get mime/type
	buff := make([]byte, 512)
	_, err := file.Read(buff)
	if err != nil {
		return fileType, fmt.Errorf("error reading file: %w", err)
	}
	fileType = http.DetectContentType(buff)
	accepted := false
	for _, v := range mimeTypes {
		if strings.Contains(fileType, v) {
			accepted = true
			fmt.Println("Matched with ", v)
		}
	}
	if !accepted {
		return fileType, errors.New("filetype not allowed")
	}
	// Need to call Seek() to reset file pointer to beginning of file.
	if _, err := file.Seek(0, 0); err != nil {
		return fileType, fmt.Errorf("error resetting file seek position: %w", err)
	}
	return fileType, nil
}

func (k *keycloak) viewDuplicateFilesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") != "TRUE" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		files, err := upload.GetOrphanFilesDB()
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}

		TemplRender(w, r, templates.ViewOrphanFiles("View Orphan Files", k.currentUser, files))
	})
}

func (k *keycloak) deleteDuplicateFilesHandler(bc *upload.BucketConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("DEV_ENV") != "TRUE" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		files, err := upload.GetOrphanFilesDB()
		if err != nil {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		}
		if err := bc.DeleteBucketFiles(files); err != nil {
			fmt.Println(err)
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		if err := upload.DeleteOrphanFilesDB(files); err != nil {
			fmt.Println(err)
			w.Header().Set("Hx-Redirect", "/error")
			return
		}
		TemplRender(w, r, templates.Toast("success", "Deleted successfully!"))
	})
}

func (k *keycloak) viewErrorHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.Error(k.currentUser, "Oops, something went wrong."))
}

func (k *keycloak) viewErrorUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	TemplRender(w, r, templates.ErrorUnauthorized(k.currentUser, "You'll need to login to do that."))
}

func (k *keycloak) searchHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("invoked handler")
		query := r.URL.Query().Get("q")
		sort := r.URL.Query().Get("s")
		if sort != "relevance" && sort != "recent" {
			if r.Header.Get("Hx-Request") == "" {
				http.Redirect(w, r, "/error", http.StatusSeeOther)
			} else {
				w.Header().Set("Hx-Redirect", "/error")
			}
			return
		}
		results, err := posts.SearchComments(query, sort)
		if err != nil {
			fmt.Println(err)
			w.Header().Set("Hx-Redirect", "/error")
		}
		if r.Header.Get("Hx-Request") == "" {
			TemplRender(w, r, templates.FullSearchResults(k.currentUser, query, sort, results))
			return
		}
		TemplRender(w, r, templates.ResultsList(results, query))
	})
}

func resetAdmin(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DEV_ENV") == "TRUE" {
		err := database.Reset()
		if err != nil {
			fmt.Println(err)
			_, err := w.Write([]byte("Reset failed, errored out\r\n"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			msg := fmt.Sprintf("%w", err)
			if _, err := io.WriteString(w, msg); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		s := time.Now().Format(time.RFC3339)

		TemplRender(w, r, templates.Reset("", s))
	} else {
		if _, err := w.Write([]byte("Not allowed!")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
