package main

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// MessageWillBeUpdated is invoked when a message is being updated
func (p *Plugin) MessageWillBeUpdated(c *plugin.Context, newPost *model.Post, oldPost *model.Post) (*model.Post, string) {
	if shouldPreventEdit(oldPost) {
		return nil, "Deleted root post cannot be edited"
	}
	return newPost, ""
}

// ReactionHasBeenAdded is involked when a reaction has been added to a post.
// There currently is not way to block adding a reaction (no `ReactionWillBeAdded`)
// so we'll just delete the new reaction immediately to simulate a deleted post that
// disallows reactions.
func (p *Plugin) ReactionHasBeenAdded(c *plugin.Context, reaction *model.Reaction) {
	if reaction == nil {
		return
	}

	post, appErr := p.API.GetPost(reaction.PostId)
	if appErr != nil {
		// ignore the error and leave the reaction alone.
		p.API.LogDebug("cannot fetch post to check for reaction blocking", "err", appErr.Error())
		return
	}

	if !shouldPreventEdit(post) {
		return
	}

	if appErr = p.API.RemoveReaction(reaction); appErr != nil {
		p.API.LogError("cannot remove reaction", "err", appErr.Error())
	}
}

// shouldPreventEdit returns true if a post has been deleted as root, meaning it contains the property "del=true".
func shouldPreventEdit(post *model.Post) bool {
	// check if the post has the DeletedRootPostPropKey property set
	val := post.GetProp(DeletedRootPostPropKey)
	deleted, ok := val.(bool)
	return ok && deleted
}
