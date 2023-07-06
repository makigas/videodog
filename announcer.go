package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
)

// AnnounceVideo will send a Discord notification using the given webhookUrl
// as a webhook endpoint, notifying about the video given as parameter. The
// roleID parameter indicates the ID of the role that should be additionally
// be included.
func AnnounceVideo(webhookUrl string, roleID string, video *Video) error {
	log.Printf("Announcing video %s over Discord...", video.VideoId)

	client, err := webhook.NewWithURL(webhookUrl)
	if err != nil {
		return err
	}
	msg, err := client.CreateMessage(discord.WebhookMessageCreate{
		Content: webhookMessage(video, roleID),
		Embeds:  []discord.Embed{webhookEmbed(video)},
	})

	log.Printf("Announced video %s as message %s-%s", video.VideoId, msg.ChannelID, msg.ID)
	return err
}

func webhookMessage(video *Video, roles ...string) string {
	var (
		builder strings.Builder
		base    = fmt.Sprintf("**%s**\n<%s>", video.Title, video.VideoUrl)
	)
	builder.WriteString(base)

	// Append roles into the string.
	for _, role := range roles {
		mention := fmt.Sprintf(" <@&%s>", role)
		builder.WriteString(mention)
	}

	return builder.String()
}

func webhookEmbed(video *Video) discord.Embed {
	return discord.NewEmbedBuilder().
		SetAuthorName(video.ChannelName).
		SetTitle(video.Title).
		SetDescription(cleanupDescription(video.Description)).
		SetURL(video.VideoUrl).
		SetImage(video.ThumbnailUrl).
		Build()
}

func cleanupDescription(description string) string {
	paragraphs := strings.Split(description, "\n")
	paragraph := paragraphs[0]
	if len(paragraph) > 240 {
		return paragraph[:237] + "..."
	}
	return paragraph
}
