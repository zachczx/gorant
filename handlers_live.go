package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"gorant/live"
	"gorant/templates"
)

func (k *keycloak) mainLivePageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instPosts, err := live.ListLivePosts()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(instPosts)
		instComments, err := live.ListLiveComments()
		if err != nil {
			fmt.Println(err)
		}
		TemplRender(w, r, templates.ViewInstantMain(k.currentUser, instPosts, instComments))
	})
}

func (k *keycloak) newInstantPostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		new := live.InstantPost{UserID: k.currentUser.UserID, Title: title}
		err := live.CreateInstantPost(new)
		if err != nil {
			fmt.Println(err)
		}
		instP, err := live.ListLivePosts()
		if err != nil {
			fmt.Println(err)
		}
		TemplRender(w, r, templates.ListInstantPosts(instP))
	})
}

func (k *keycloak) viewInstantCommentsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("instPID")
		// TemplRender(w, r, templates.LiveScreenBase(k.currentUser))
		TemplRender(w, r, templates.ViewInstantPost(k.currentUser, id))
	})
}

func (k *keycloak) newInstantCommentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instPID, err := strconv.Atoi(r.PathValue("instPID"))
		if err != nil {
			fmt.Println(err)
		}
		c := r.FormValue("content")
		new := live.InstantComment{InstantPostID: instPID, Content: c, UserID: k.currentUser.UserID}
		err = live.CreateInstantComment(new)
		if err != nil {
			fmt.Println(err)
		}
		// instC, err := live.ListLiveComments()
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// TemplRender(w, r, templates.ViewListInstantComments(instC))
	})
}

func sseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instPID, err := strconv.Atoi(r.PathValue("instPID"))
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		if os.Getenv("DEV_ENV") == "TRUE" {
			fmt.Println("Set Access-Control-Allow-Origin header") // You may need this locally for CORS requests
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
		}

		// Create a channel for client disconnection
		clientGone := r.Context().Done()
		rc := http.NewResponseController(w)
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		var holder string
		var instComments []live.InstantComment
		var lastID int = 0

		for {
			select {
			case <-clientGone:
				fmt.Println("Client disconnected!")
				return
			case <-t.C:
				instComments, err = live.ViewLivePost(instPID)
				if len(instComments) == 0 {
					fmt.Println("No live comments")
					holder = "event:instant\ndata:No data found!\n\n"
					w.Write([]byte(holder))
					err = rc.Flush()
					if err != nil {
						fmt.Println(err)
						return
					}
					return
				}
				// Using instComments[0].ID instead of instComments[len(instComments) - 1].ID because I'm sorting by DESC order.
				if lastID < instComments[0].ID {
					lastID = instComments[0].ID
					holder = "event: instant\ndata:"
					if err != nil {
						fmt.Println(err)
					}
					for _, v := range instComments {
						holder = holder + fmt.Sprintf("<div class='flex'><div class='avatar'><span>%s</span></div><div class='comment'><span class='user'>%s<span class='time'>%v (%v)</span></span><span class='content'>%s</span></div></div>", v.PreferredNameInitials(), v.PreferredName, v.CreatedAt.Format(time.Kitchen), v.CreatedAt.Format("02 Jan"), v.Content)
					}
					holder = holder + "\n\n"
					fmt.Println(holder)
					w.Write([]byte(holder))

					err = rc.Flush()
					if err != nil {
						fmt.Println(err)
						return
					}
				}

			}
		}
	})
}
