package main

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// A Channel is a particular notofication configuration. It is a specific block
// designed to forward new videos uploaded into a particular YouTube channel
// to the given Discord webhook, notifying the specific members part of a role
// by issuing a ping in the message.
type Channel struct {
	Feed    string `json:"channel_id"`  // The channel ID whose feed will be read from YouTube
	Webhook string `json:"webhook_url"` // The webhook URL where new videos should be posted
	Role    string `json:"role_id"`     // The role to ping on announcements

	lastCheck     time.Time  // The last time the channel was polled for new content
	lastCheckLock sync.Mutex // To prevent multiple threads from doing weird things
}

// TryCheckStale will return true if the feed is stale and has to be checked again.
// When checking if the feed is stale, it will mark it as new at the same time,
// so this function will only return true every couple of minutes.
func (c *Channel) TryCheckStale() bool {
	c.lastCheckLock.Lock()
	defer c.lastCheckLock.Unlock()

	staleLimit := time.Now().Add(-time.Duration(flagInterval) * time.Minute)
	if c.lastCheck.Before(staleLimit) {
		// It has been ten minutes since the last time we checked the feed.
		c.lastCheck = time.Now()
		return true
	}
	return false
}

// A ChannelList contains the list of channels. While it would be great to have
// them as an array, it is important to treat the spool for each channel as
// separate, thus we need a way to identify each spool. Changing the name
// of a key will reset the spool.
type ChannelList map[string]*Channel

// A Config struct is the complete settings of the application.
type Config struct {
	Spoolfile string      `json:"spool"`       // The spoolfile to use to remember notified videos
	Autospool bool        `json:"autodiscard"` // Whether to mark all videos in feed as notified on startup
	Channels  ChannelList `json:"channels"`    // The list of channels that will be watched by the agent
}

// ReadConfigObject processes the given Reader structure to configure the
// behaviour of the agent based on the JSON contents there. The Reader should
// provide a valid JSON document that is conformant with the schema. The
// outcome will be given as a result.
func ReadConfigObject(r io.Reader) (*Config, error) {
	var cfg Config
	cfg.Autospool = true

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
