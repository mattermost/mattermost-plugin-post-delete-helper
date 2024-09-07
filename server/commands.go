package main

import (
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
		return p.createErrorCommandResponse("Invalid number of arguments, use `/deleterootpost [postID]`."), nil
	}

	postID := commandSplit[1]

	// Check if the post ID is a valid ID
	if !model.IsValidId(postID) {
		return p.createErrorCommandResponse("Invalid post ID"), nil
	}

	// Check if the user has permissions to remove the post
	if errReason := p.userHasRemovePermissionsToPost(args.UserId, args.ChannelId, postID); errReason != "" {
		return p.createErrorCommandResponse(errReason), nil
	}

	// Create an interactive dialog to confirm the action
	if err := p.API.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       "/plugins/" + manifest.Id + "/api/v1/delete_root_post?post_id=" + postID,
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
