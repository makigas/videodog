package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
)

// A Video is the public-facing data structure that you get once you get videos from the feed.
type Video struct {
	ChannelName  string // The channel name as presented in the embed.
	VideoId      string // The internal YouTube ID of the video.
	Title        string // The public title of the video.
	Description  string // The channel description as fetched from the feed.
	VideoUrl     string // The URL where the video can be viewed.
	ThumbnailUrl string // The URL for the thumbnail of this video.
}

// FetchVideos will query the remote for the recent videos in the channel whose ID is given
// as parameter and return the list. The feed is unmarshaled and demangled so that the
// function can return Video entities.
func FetchYouTubeFeed(channelId string) ([]Video, error) {
	bytes, err := fetchFeed(channelId)
	if err != nil {
		return nil, err
	}
	unmarshal, err := unmarshalFeed(bytes)
	if err != nil {
		return nil, err
	}
	return demangleFeed(unmarshal), nil
}

func fetchFeed(channelId string) ([]byte, error) {
	url := fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelId)
	log.Printf("Downloading feed from %s", url)
	res, err := http.Get(url)
	log.Printf("Status code from request: %d", res.StatusCode)
	if err != nil {
		return nil, err
	}
	if res.StatusCode > 299 {
		err := fmt.Errorf("invalid status code: %d", res.StatusCode)
		return nil, err
	}
	return io.ReadAll(res.Body)
}

func unmarshalFeed(data []byte) (xmlNodeFeed, error) {
	var feed xmlNodeFeed
	err := xml.Unmarshal(data, &feed)
	log.Printf("Unmarshaled %d entries from the feed", len(feed.Entries))
	return feed, err
}

func demangleFeed(feed xmlNodeFeed) []Video {
	videos := make([]Video, len(feed.Entries))
	for i, entry := range feed.Entries {
		videos[i].VideoId = entry.VideoId
		videos[i].ChannelName = feed.Title
		videos[i].Title = entry.Title
		videos[i].Description = entry.Group.Description
		videos[i].VideoUrl = entry.Link.Href
		// Note that we ignore the original thumbnail URL carried by the feed
		// and we upgrade automatically into the maxresdefault rendition so
		// that we can have HD thumbnails.
		videos[i].ThumbnailUrl = fmt.Sprintf(`https://i1.ytimg.com/vi/%s/maxresdefault.jpg`, entry.VideoId)
	}
	return videos
}

type xmlNodeLink struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
}

type xmlNodeGroup struct {
	XMLName     xml.Name `xml:"group"`
	Description string   `xml:"http://search.yahoo.com/mrss/ description"`
}

type xmlNodeEntry struct {
	XMLName xml.Name     `xml:"entry"`
	VideoId string       `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
	Title   string       `xml:"title"`
	Link    xmlNodeLink  `xml:"link"`
	Group   xmlNodeGroup `xml:"http://search.yahoo.com/mrss/ group"`
}

type xmlNodeFeed struct {
	XMLName xml.Name       `xml:"feed"`
	Title   string         `xml:"title"`
	Entries []xmlNodeEntry `xml:"entry"`
}
