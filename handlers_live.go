package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gorant/live"
	"gorant/templates"

	"github.com/google/uuid"
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
		instPID, err := uuid.Parse(r.PathValue("instPID"))
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

func sseHandler(w http.ResponseWriter, r *http.Request) {
	instPID, err := uuid.Parse(r.PathValue("instPID"))
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	if os.Getenv("DEV_ENV") == "TRUE" {
		// fmt.Println("Set Access-Control-Allow-Origin header") // You may need this locally for CORS requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	}

	// Create a channel for client disconnection
	clientGone := r.Context().Done()

	var sb strings.Builder

	// Gzip doesn't play nice with ResponseController. Caused errors with flushing.
	// From Gemini - gzip.Writer buffers data for efficiency.
	// 		The rc.Flush() method from http.ResponseController doesn't interact with this internal buffering of the gzip.Writer.
	// 		It tries to flush the underlying http.ResponseWriter, but the gzip.Writer might not have sent anything to it yet.
	// 		This leads to errors when the connection expects data to be flushed, especially in the context of Server-Sent Events where real-time updates are crucial.
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	var instComments []live.InstantComment
	var lastCheckTime time.Time
	fmt.Println(lastCheckTime)
	// Necessary to do this assertion and see if this handler can flush data, else streaming wouldn't work.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-clientGone:
			fmt.Println(time.Now(), "Client disconnected!")
			return
		case <-t.C:
			instComments, err = live.ViewLivePost(instPID)
			if err != nil {
				fmt.Println(err)
			}
			// String builder in lieu of concatenation because it works better.
			sb.Reset()
			if len(instComments) == 0 {
				fmt.Println("No live comments")
				sb.WriteString("event:instant\ndata:No data found!\n\n")
				if _, err := w.Write([]byte(sb.String())); err != nil {
					log.Fatal(err)
				}
				flusher.Flush()
				return
			}
			// Using instComments[0].CreatedAt instead of instComments[len(instComments) - 1].CreatedAt,
			// because I'm sorting by DESC order.
			latestPostTime := instComments[0].CreatedAt
			// This check is important, otherwise JS on frontend will keep receiving successful SSE responses,
			// and my JS logic resets the input value and sets focus to it. Making it impossible to use.
			if latestPostTime.After(lastCheckTime) {
				lastCheckTime = latestPostTime
				sb.WriteString("event: instant\ndata:")
				for _, v := range instComments {
					sb.WriteString(fmt.Sprintf("<div class='flex'><div class='avatar'><span>%s</span></div><div class='comment'><span class='user'>%s<span class='time'>%v (%v)</span></span><span class='content'>%s</span></div></div>", v.PreferredNameInitials(), v.PreferredName, v.CreatedAt.Format(time.Kitchen), v.CreatedAt.Format("02 Jan"), v.Content))
				}
				sb.WriteString("\n\n")

				if _, err := w.Write([]byte(sb.String())); err != nil {
					fmt.Println("write error: ", err)
				}
				flusher.Flush()
			}
		}
	}
}
