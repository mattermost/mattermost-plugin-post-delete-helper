# Post Delete Helper 

[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-post-delete-helper)](https://github.com/mattermost/mattermost-plugin-post-delete-helper/releases/latest)
[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-post-delete-helper/master)](https://circleci.com/gh/mattermost/mattermost-plugin-post-delete-helper)

## Features

- Adds a post menu option to delete root posts without deleting the replies
    - message text is replaced with a customizable messsage
    - all reactions are removed
- root posts are soft-deleted, retaining the original timestamp, author and profile image
- deleted root posts cannot be editted or reacted to

## Usage 

Use the action menu attached to a root post to select the `Remove Root Post` menu item.

![Remove-root-post-crt](https://github.com/user-attachments/assets/95a31080-80c0-4348-94d1-7803da7aad39)

## Configuration

You may optionally provide a custom string to be used when deleting a root post. This string will replace the original post contents.

![Screenshot from 2024-09-24 00-36-14](https://github.com/user-attachments/assets/3fcad579-902a-45ac-bc0d-55d95c2806c2)
