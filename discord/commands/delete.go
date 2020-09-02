package commands

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bwmarrin/discordgo"
	"time"
	"turtl/db"
	"turtl/storage"
	"turtl/utils"
)

func deleteCommand(s *discordgo.Session, m *discordgo.Message) {
	args := UseArgs(m)
	if len(args) < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I need a file to delete! Please use the file URL as the first argument.")
		return
	}

	file, ok := db.GetFileFromURL(args[0])
	if !ok || file.Bucket == "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "I couldn't find that file.")
		return
	}

	if file.DeletedAt != 0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "This file has already been deleted.")
		return
	}

	isAdmin, ok := db.CheckAdmin(m)
	if !ok {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	if file.Uploader != m.Author.ID && !isAdmin {
		_, _ = s.ChannelMessageSend(m.ChannelID, "That file doesn't belong to you!")
		return
	}

	_, err := storage.S3Service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(file.Bucket),
		Key:    aws.String(file.FileName),
	})
	if utils.HandleError(err, "delete file from bucket") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}

	_, err = db.DB.Exec("update objects set deletedat=$1 where bucket=$2 and wildcard=$3 and filename=$4 and uploader=$5 and createdat=$6 and md5=$7 and sha256=$8", time.Now().Unix(), file.Bucket, file.Wildcard, file.FileName, file.Uploader, file.CreatedAt, file.MD5, file.SHA256)
	if utils.HandleError(err, "delete file from db") {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Error! Please try again later.")
		return
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "File has been successfully deleted!")
}

func init() {
	RegisterCommand(&Command{
		Exec:     deleteCommand,
		Trigger:  "delete",
		Aliases:  []string{"deletefile", "remove", "removefile"},
		Usage:    "delete <file URL>",
		Desc:     "Delete a file from our servers",
		Disabled: false,
	})
}
