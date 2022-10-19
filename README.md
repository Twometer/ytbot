# ytbot

ytbot is a self-hosted YouTube Streaming Bot for Discord written in Go

## About

Currently, most major music bots are stopping their YouTube streaming functionality or can no longer
be added to new servers due to having their verification status revoked.

This project aims to be an open-source and self-hosted replacement for these kinds of bots. It
uses [yt-dlp](https://github.com/yt-dlp/yt-dlp)
as its YouTube backend, and features a custom Discord Client library for API v10.

## Usage

The ytbot requires Windows or Linux, Go >= 1.18, and an [ffmpeg](https://ffmpeg.org/) installation. It can be built like
any other Go application. On the first run, it will download a copy of yt-dlp into the working directory.

For the bot to start, the following environment variables have to be set

| Variable name         | Description                                                                                                                                                                 |
|-----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `YTB_AUTH_TOKEN`      | A Discord Bot authentication token. [Register an application](https://discord.com/developers/applications) at Discord, create a bot for it, and you will get your own token |
| `YTB_FFMPEG_LOCATION` | The path to the ffmpeg executable (not the installation directory)                                                                                                          |