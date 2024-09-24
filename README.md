# Post Delete Helper 

[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-post-delete-helper)](https://github.com/mattermost/mattermost-plugin-post-delete-helper/releases/latest)
[![Build Status](https://github.com/mattermost/mattermost-plugin-post-delete-helper/actions/workflows/ci.yml/badge.svg)](https://github.com/mattermost/mattermost-plugin-post-delete-helper/actions/workflows/ci.yml)

## Features

- Adds a post menu option to delete root posts without deleting the replies
    - message text is replaced with a customizable messsage
    - all reactions are removed
- root posts are soft-deleted, retaining the original timestamp, author and profile image
- deleted root posts cannot be editted or reacted to
