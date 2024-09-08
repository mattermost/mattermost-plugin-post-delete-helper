# Post Delete Helper 

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-post-delete-helper/master)](https://circleci.com/gh/mattermost/mattermost-plugin-post-delete-helper)

## Features

- Adds a post menu option to delete root posts without deleting the replies
    - message text is replaced with a customizable messsage
    - all reactions are removed
- root posts are soft-deleted, retaining the original timestamp, author and profile image
- deleted root posts cannot be editted or reacted to
