package webexbot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pavelzagorodnyuk/webexbot/internal/webexapi/v1"
)

type Dialog interface {
	context.Context

	PersonId() string
	RoomId() string

	Inform(InformationMessage) (messageId string, err error)
}

type InformationMessage struct {
	ParentId     string
	PlainText    string
	Text         string
	TextFormat   TextFormat
	File         *File
	AdaptiveCard json.RawMessage
}

type TextFormat int

const (
	TextFormatMarkdown TextFormat = iota
	TextFormatHTML
)

type File webexapi.File

type dialog struct {
	context.Context

	personId, roomId string
	incomingEvents   <-chan Event
	webexClient      webexapi.Client
}

func (d *dialog) Inform(message InformationMessage) (messageId string, err error) {
	createMessageRequest, err := d.newCreationRequestForInformationMessage(message)
	if err != nil {
		return "", fmt.Errorf("unable to prepare a message creation request : %w", err)
	}

	webexMessage, webexErr, err := d.webexClient.CreateMessage(d.Context, createMessageRequest)
	const unableToSendMessage = "unable to send the message : %w"
	if err != nil {
		return "", fmt.Errorf(unableToSendMessage, err)
	}
	if webexErr != nil {
		return "", fmt.Errorf(unableToSendMessage, webexErr)
	}

	return webexMessage.Id, nil
}

func (d *dialog) newCreationRequestForInformationMessage(
	message InformationMessage,
) (
	webexapi.CreateMessageRequest,
	error,
) {
	request := webexapi.CreateMessageRequest{
		ToPersonId: d.personId,
		RoomId:     d.roomId,
		Text:       message.PlainText,
		File:       (*webexapi.File)(message.File),
	}

	switch message.TextFormat {
	case TextFormatMarkdown:
		request.Markdown = message.Text

	case TextFormatHTML:
		request.Html = message.Text

	default:
		return webexapi.CreateMessageRequest{}, errors.New("unknown text format")
	}

	if message.AdaptiveCard != nil {
		request.Attachment = &webexapi.Attachment{
			ContentType: "application/vnd.microsoft.card.adaptive",
			Content:     message.AdaptiveCard,
		}
	}

	return request, nil
}
