package main

import "github.com/mattermost/mattermost/server/public/model"

// userHasRemovePermissionsToPost checks if the user has permissions to delete a post
// based on the post ID, the user ID, and the channel ID.
// Returns an error message if the user does not have permissions, or an empty string if the user has permissions.
func (p *Plugin) userHasRemovePermissionsToPost(userID, channelID, postID string) string {
	// Check if the post exists
	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		p.API.LogError("error fetching post", "post_id", postID, "err", appErr.Error())
		return "Post does not exist"
	}

	// Check if the user is the post author or has permissions to edit others posts
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		p.API.LogError("error fetching user", "user_id", userID, "err", appErr.Error())
		return "Internal error, check with your system administrator for assistance"
	}

	var permission *model.Permission
	if post.UserId == user.Id {
		permission = model.PermissionEditPost
	} else {
		permission = model.PermissionEditOthersPosts
	}

	if !p.API.HasPermissionToChannel(userID, channelID, permission) {
		return "Not authorized"
	}

	// Check if the post is editable at this point in time
	config := p.API.GetConfig()
	if config.ServiceSettings.PostEditTimeLimit != nil &&
		*config.ServiceSettings.PostEditTimeLimit > 0 &&
		model.GetMillis() > post.CreateAt+int64(*config.ServiceSettings.PostEditTimeLimit*1000) {
		return "Post is too old to edit"
	}

	return ""
}
