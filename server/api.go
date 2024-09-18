package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/wiggin77/merror"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	DeletedRootPostPropKey = "rootdel"
)

type API struct {
	plugin *Plugin
	router *mux.Router
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *API) handlerDeleteRootPost(w http.ResponseWriter, r *http.Request) {
	// user must be authenticated; this is done by the Mattermost server before being passed here,
	// and the Mattermost server adds the Mattermost-User-ID header only if authentication is successful.
	userID := r.Header.Get("Mattermost-User-ID")
	if !model.IsValidId(userID) {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get post_id from the query parameters
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postID := r.FormValue("post_id")
	if !model.IsValidId(postID) {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, appErr := a.plugin.API.GetPost(postID)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	// Check if post is already deleted
	if post.DeleteAt != 0 {
		http.Error(w, "Post already deleted", http.StatusBadRequest)
		return
	}

	// Check if the user is the post author or a system admin
	if errReason := a.plugin.userHasRemovePermissionsToPost(userID, post.ChannelId, postID); errReason != "" {
		http.Error(w, errReason, http.StatusForbidden)
		return
	}

	// Check if the post is a root post
	if post.RootId != "" || post.ReplyCount == 0 {
		http.Error(w, "Post is not a root post of a thread", http.StatusBadRequest)
		return
	}

	// current requirements are to keep the original author and avatar. To do this without changing webapp,
	// we can only delete the message text, attachments, and reactions.  The post will be kept so it is
	// displayed correctly without contents.
	//
	// This is a temporary solution as a future server version will default to not deleting threads when the
	// root post is deleted.

	originalFileIDs := post.FileIds

	post.Message = a.plugin.getConfiguration().DeletedMessage
	post.MessageSource = ""
	post.FileIds = []string{}
	post.AddProp(DeletedRootPostPropKey, true) // mark the post as a deleted root post
	newPost, appErr := a.plugin.API.UpdatePost(post)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	// remove reactions. This is not very efficient, deleting each reaction individually, however for a plugin
	// that will be obsolete with the next server version (see above comment) it is not worth re-creating all
	// the deletion logic, including websocket events and cache invalidation.
	reactions, appErr := a.plugin.API.GetReactions(newPost.Id)
	if appErr != nil {
		a.plugin.API.LogError("error fetching reactions for post", "post_id", newPost.Id, "err", appErr.Error())
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}
	merr := merror.New()
	for _, reaction := range reactions {
		if appErr = a.plugin.API.RemoveReaction(reaction); appErr != nil {
			merr.Append(appErr)
		}
	}
	if merr.Len() > 0 {
		a.plugin.API.LogError("error removing reactions for post", "post_id", newPost.Id, "err", merr)
		http.Error(w, "Internal server error, check logs", http.StatusInternalServerError)
		return
	}

	// soft-delete the attachments from channel
	merr = merror.New()
	for _, fileID := range originalFileIDs {
		if err := a.plugin.SQLStore.DetachAttachmentFromChannel(fileID); err != nil {
			merr.Append(err)
		}
	}
	if merr.Len() > 0 {
		a.plugin.API.LogError("error detaching attachments from channel", "err", merr)
		http.Error(w, "Internal server error, check logs", http.StatusInternalServerError)
		return
	}

	// attach the original file IDs to the original post so history is not lost
	if err := a.plugin.SQLStore.AttachFileIDsToPost(newPost.OriginalId, originalFileIDs); err != nil {
		a.plugin.API.LogError("error attaching original file IDs", "post_id", newPost.OriginalId, "err", merr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// setupAPI sets up the API for the plugin.
func setupAPI(plugin *Plugin) (*API, error) {
	api := &API{
		plugin: plugin,
		router: mux.NewRouter(),
	}

	group := api.router.PathPrefix("/api/v1").Subrouter()
	group.Use(authorizationRequiredMiddleware)
	group.HandleFunc("/delete_root_post", api.handlerDeleteRootPost).Methods(http.MethodPost)

	return api, nil
}

func authorizationRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}
