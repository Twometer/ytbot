# ytbot

ytbot is a self-hosted YouTube Streaming Bot for Discord written in Go

## About

Currently, most major music bots are stopping their YouTube streaming functionality or can no longer
be added to new servers due to having their verification status revoked.

This project aims to be an open-source and self-hosted replacement for these kinds of bots. It
uses [yt-dlp](https://github.com/yt-dlp/yt-dlp)
as its YouTube backend, and features a custom Discord Client library for API v10.

## Setup

The ytbot requires Windows or Linux, Go >= 1.18, and an [ffmpeg](https://ffmpeg.org/) installation. It can be built like
any other Go application. On the first run, it will download a copy of yt-dlp into the working directory.

For the bot to start, the following environment variables have to be set

| Variable name         | Description                                                                                                                                                                 |
|-----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `YTB_AUTH_TOKEN`      | A Discord Bot authentication token. [Register an application](https://discord.com/developers/applications) at Discord, create a bot for it, and you will get your own token |
| `YTB_FFMPEG_LOCATION` | The path to the ffmpeg executable (not the installation directory)                                                                                                          |

## Usage

The bot is controlled using message-based commands prefixed with a dot (`.`)

| Command             | Description                                                                          |
|---------------------|--------------------------------------------------------------------------------------|
| `.play <query>`     | Adds one or more YouTube videos by link, playlist link, or search query to the queue |
| `.skip`             | Skips to next media item in the queue                                                |
| `.stop or .leave`   | Stops playback, leaves voice channel, and clears queue                               |
| `.move <from> <to>` | Moves an item in the playback queue                                                  |
| `.clear`            | Clears the playback queue                                                            |
| `.remove <item>`    | Removes an item from the playback queue                                              |
| `.queue <page>`     | Shows a page of the playback queue. Shows 10 items per page.                         |