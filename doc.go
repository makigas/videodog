/*
Videodog is a daemon that will announce new YouTube videos. This application
connects two external integrations: the public RSS feed of a YouTube channel,
carrying the most recent videos from a channel, and a Discord webhook where
notifications about new videos should be sent.

# Synopsis

Usage:

	videodog -config <config> [-noannounce]

Parameters:

	-config <config>
	  Point to a valid configuration file to provide to videodog.
	-noannounce
	  Do not announce videos in the Discord webhook when uploaded.

# Spool

To persist data between executions, there is a spool file. The spool internally
is an SQLite database that lists the videos that have been announced in the
past. Whenever a YouTube feed is fetched, each video is tested and is skipped
if found in the spool. For videos that are not part of the spool, they are
considered NEW, and thus they are announced to Discord before being added
into the spool, so that no double notifications are sent in the future.

Because videodog may monitor multiple YouTube channels or forward to multiple
Discord webhooks, the spool also groups videos by channel identifier, so that
each announcement configuration is separate.

The spool is a standard SQLite database, so it can be modified from the outside
using any SQLite3 comaptible client. The schema is simple and only has a single
table called announcements, which has indexes to keep the performance tight.

# Configuration file

The configuration file is a JSON file containing the following keys:

  - spool (string): the path to the file to use as spool. Keep this value
    constant to allow the application state to be persisted between application
    executions.
  - autodiscard (bool): if this value is provided, it will set whether videodog
    should fetch all the feeds on application startup and mark every video in
    the feed in the spool. This way, only new videos uploaded after the
    application has started will be announced. If not given, the default value
    will be true.
  - channel (object): configuration for each channel (see below).

Each one of the keys of the channels object is the unique name of a channel
configuration. Multiple channels can be configured given the names are unique,
allowing a single application to monitor multiple channels part of the same
network and/or to announce a channel into multiple Discord webhooks.

For each channel entry, the following object is used:

  - channel_id (string): The Youtube channel ID to monitor. This is the internal
    ID, not the channel URL or handle, and you can get it from the YouTube
    settings.
  - webhook_url (string): The Discord webhook URL where the notification may
    be sent once a new video is uploaded. The system will send a message with
    some content and an embed, but it will not customize the channel name. This
    should be done in the Discord settings.
  - role_id (string): the Discord role to ping whenever a new video is added.
    This probably prevents pinging @everyone. Instead, a specific role has to be
    provided.

An example object is given below:

	{
	  "spool": "./data/spool.db",
	  "autodiscard": true,
	  "channels": {
	    "main_channel": {
	      "channel_id": "1_TheMainChannelID_1",
	      "webhook_url": "https://webhook.example.com",
	      "role_id": "12341234"
	    },
	    "secondary_channel": {
	      "channel_id": "2_TheSecondaryChannelID_2",
	      "webhook_url": "https://webhook2.example.com",
	      "role_id": "56785678"
	    }
	  }
	}

# Running in Docker

A Dockerfile is provided with the source code of the application. The daemon
may be dockerized to be run in a sandbox environment. Since the spoolfile is
an SQLite file, videodog is probably not a serverless application.

The Dockerfile is configured to run the application as an ENTRYPOINT, so you
should use the CMD to configure the application parameters. As an example,
here is a valid compose.yml / docker-compose.yml file:

	---
	version: '3.8'
	services:
	  daemon:
	    build: .
	    volumes:
	      - ./docker:/docker
	    command: ["-config", "/docker/config.json"]

The Docker directory may contain a configuration file at /docekr/config.json,
with the spoolfile configured to /docker/spoolfile.db. This way, the spool
will be persisted between sessions. However, running multiple instances of
the application pointing to the same spoolfile is not encouraged.

# Open Source Policy

This package has been made open source in the hope that it is useful for
people studying the behaviour of this software or the programming language or
library set.

However, this is not an open effort. Therefore, issues and pull requests may
be ignored. This program was designed to fulfill some specific requirements
that may not fit the requirements of other people. If other people is reading
this and considering that the application does not behave as expected, they
are free to write their own integrations.

# Tasks and enhancements

  - Allow to treat role_id as an optional key. If not given, the notification
    will not ping. Part of the code is already written but not everything is
    integratedx.
*/
package main
