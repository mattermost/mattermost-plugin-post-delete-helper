package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) getCommand() *model.Command {
	return &model.Command{
		Trigger:      "deleterootpost",
		DisplayName:  "Delete root post",
		Description:  "Delete a root post without deleting the thread.",
		AutoComplete: false, // Hide from autocomplete
	}
}

func (p *Plugin) createCommandResponse(message string) *model.CommandResponse {
	return &model.CommandResponse{
		Text: message,
	}
}

func (p *Plugin) createErrorCommandResponse(errorMessage string) *model.CommandResponse {
	return &model.CommandResponse{
		Text: "Can't delete root post: " + errorMessage,
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	commandSplit := strings.Split(args.Command, " ")

	// Do not provide command output since it's going to be triggered from the frontend
	if len(commandSplit) != 2 {
		return p.createErrorCommandResponse("invalid number of arguments, use `/deleterootpost [postID]`."), nil
	}

	postID := commandSplit[1]

	// Check if the post ID is a valid ID
	if !model.IsValidId(postID) {
		return p.createErrorCommandResponse("invalid post ID"), nil
	}

	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		return p.createErrorCommandResponse("cannot fetch post - " + appErr.Error()), nil
	}

	// Check if root post can be deleted
	if _, err := p.checkCanDeleteRootPost(args.UserId, post); err != nil {
		return p.createErrorCommandResponse(err.Error()), nil
	}

	// Create an interactive dialog to confirm the action
	if err := p.API.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       "/plugins/" + url.PathEscape(manifest.Id) + "/api/v1/delete_root_post?post_id=" + postID,
		Dialog: model.Dialog{
			Title:            "Delete Root Post",
			IntroductionText: "Are you sure you want to delete this post? The thread will remain.",
			SubmitLabel:      "Delete",
		},
	}); err != nil {
		return p.createCommandResponse(err.Error()), nil
	}

	// Return nothing, let the dialog/api handle the response
	return &model.CommandResponse{}, nil
}

func (p *Plugin) checkCanDeleteRootPost(userID string, post *model.Post) (int, error) {
	// Check if post is already deleted
	if post.DeleteAt != 0 {
		return http.StatusBadRequest, errors.New("post already deleted")
	}

	// Check if post is already root deleted by this plugin
	val := post.GetProp(DeletedRootPostPropKey)
	deleted, ok := val.(bool)
	if ok && deleted {
		return http.StatusBadRequest, errors.New("root post already deleted")
	}

	// Check if the user is the post author or a system admin
	if errReason := p.userHasRemovePermissionsToPost(userID, post.ChannelId, post.Id); errReason != "" {
		return http.StatusForbidden, errors.New(errReason)
	}

	// Check if the post is a root post
	if post.RootId != "" || post.ReplyCount == 0 {
		return http.StatusBadRequest, errors.New("post is not root of a thread")
	}
	return http.StatusOK, nil
}
