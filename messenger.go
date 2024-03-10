package webexbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"

	"github.com/pavelzagorodnyuk/webexbot/internal/webexapi/v1"
)

type Messenger interface {
	DialogInfo() DialogInfo
	Send(context.Context, Message) (messageId string, err error)
	Listen(context.Context, Listener) (response any, err error)
	OfferChoice(context.Context, ChoiceMessage) (optionId string, err error)
}

type DialogInfo struct {
	PersonId     string
	RoomId       string
	InitialEvent Event
}

type Message struct {
	Text         string
	Format       TextFormat
	PlainText    string
	File         *File
	AdaptiveCard json.RawMessage
	ParentId     string
}

type TextFormat int

const (
	TextFormatMarkdown TextFormat = iota
	TextFormatHTML
)

type File webexapi.File

type Listener func(Event) (response any, clarification *Message, err error)

type ChoiceMessage struct {
	Text     string
	Options  []Option
	ParentId string
}

type Option struct {
	Id    string
	Label string
}

type messengerProvider struct {
	webexClient webexapi.Client
}

func newMessengerProvider(webexClient webexapi.Client) messengerProvider {
	return messengerProvider{
		webexClient: webexClient,
	}
}

func (p messengerProvider) provide(initialEvent Event, eventChan <-chan Event) Messenger {
	return messenger{
		personId:       initialEvent.InitiatorId,
		roomId:         initialEvent.RoomId,
		initialEvent:   initialEvent,
		incomingEvents: eventChan,
		webexClient:    p.webexClient,
	}
}

type messenger struct {
	personId, roomId string
	initialEvent     Event
	incomingEvents   <-chan Event
	webexClient      webexapi.Client
}

func (m messenger) DialogInfo() DialogInfo {
	return DialogInfo{
		PersonId:     m.personId,
		RoomId:       m.roomId,
		InitialEvent: m.initialEvent,
	}
}

func (m messenger) Send(ctx context.Context, message Message) (messageId string, err error) {
	messageCreationRequest, err := m.newMessageCreationRequest(message)
	if err != nil {
		return "", fmt.Errorf("unable to prepare a message creation request : %w", err)
	}

	webexMessage, webexErr, err := m.webexClient.CreateMessage(ctx, messageCreationRequest)
	const unableToSendMessage = "unable to send the message : %w"
	if err != nil {
		return "", fmt.Errorf(unableToSendMessage, err)
	}
	if webexErr != nil {
		return "", fmt.Errorf(unableToSendMessage, webexErr)
	}

	return webexMessage.Id, nil
}

func (m messenger) newMessageCreationRequest(message Message) (webexapi.CreateMessageRequest, error) {
	request := webexapi.CreateMessageRequest{
		ToPersonId: m.personId,
		RoomId:     m.roomId,
		Text:       message.PlainText,
		File:       (*webexapi.File)(message.File),
		ParentId:   message.ParentId,
	}

	switch message.Format {
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

func (m messenger) Listen(ctx context.Context, listener Listener) (response any, err error) {
	for {
		select {
		case event := <-m.incomingEvents:
			response, clarification, err := listener(event)
			if err != nil {
				return nil, err
			}

			if response != nil {
				return response, nil
			}

			if clarification == nil {
				continue
			}

			_, err = m.Send(ctx, *clarification)
			if err != nil {
				return nil, fmt.Errorf("unable to send the clarification : %w", err)
			}

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (m messenger) OfferChoice(ctx context.Context, message ChoiceMessage) (optionId string, err error) {
	_, err = m.sendChoiceMessage(ctx, message)
	if err != nil {
		return "", fmt.Errorf("unable to send the choice message : %w", err)
	}

	return m.listenForChoice(ctx, message.Options)
}

func (m messenger) sendChoiceMessage(ctx context.Context, message ChoiceMessage) (messageId string, err error) {
	messageCreationRequest, err := m.newChoiceMessageCreationRequest(message)
	if err != nil {
		return "", fmt.Errorf("unable to prepare a message creation request : %w", err)
	}

	webexMessage, webexErr, err := m.webexClient.CreateMessage(ctx, messageCreationRequest)
	const unableToSendMessage = "unable to send the message : %w"
	if err != nil {
		return "", fmt.Errorf(unableToSendMessage, err)
	}
	if webexErr != nil {
		return "", fmt.Errorf(unableToSendMessage, webexErr)
	}

	return webexMessage.Id, nil
}

var choiceMessageAdaptiveCardTemplate = template.Must(template.New("ChoiceMessageAdaptiveCard").Parse(`{
    "type": "AdaptiveCard",
    "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
    "version": "1.3",
    "body": [
		{
            "type": "TextBlock",
            "text": "{{.Message}}",
			"wrap": true
        },
	    {
	        "type": "ActionSet",
            "actions": [
				{{- range $index, $option := .Options}}{{if $index }},{{end}}
				{
	                "type": "Action.Submit",
	                "title": "{{$option.Label}}",
					"data": "{{$option.Id}}"
                }
				{{- end}}
			]
	    }
	]
}`))

func (m messenger) newChoiceMessageCreationRequest(message ChoiceMessage) (webexapi.CreateMessageRequest, error) {
	adaptiveCardBuffer := new(bytes.Buffer)
	err := choiceMessageAdaptiveCardTemplate.Execute(adaptiveCardBuffer, message)
	if err != nil {
		return webexapi.CreateMessageRequest{}, fmt.Errorf("unable to build an adaptive card : %w", err)
	}

	return webexapi.CreateMessageRequest{
		ToPersonId: m.personId,
		RoomId:     m.roomId,
		Markdown:   message.Text,
		ParentId:   message.ParentId,
		Attachment: &webexapi.Attachment{
			ContentType: "application/vnd.microsoft.card.adaptive",
			Content:     adaptiveCardBuffer.Bytes(),
		},
	}, nil
}

func (m messenger) listenForChoice(ctx context.Context, options []Option) (string, error) {
	listener := m.newOptionsListener(options)
	userChoice, err := m.Listen(ctx, listener)
	return userChoice.(string), err
}

func (m messenger) newOptionsListener(options []Option) Listener {
	return func(event Event) (response any, clarification *Message, err error) {
		attachmentAction, isSuccessful := fetchAttachmentAction(event)
		if !isSuccessful {
			return
		}

		userChoice, err := fetchUserChoice(attachmentAction)
		if err != nil {
			return nil, nil, err
		}

		if !isOneOf(userChoice, options) {
			return
		}

		return userChoice, nil, nil
	}
}

func fetchAttachmentAction(event Event) (attachmentAction webexapi.AttachmentAction, isSuccessful bool) {
	if event.ResourceKind != AttachmentActions {
		return
	}
	attachmentAction, isSuccessful = event.Resource.(webexapi.AttachmentAction)
	return
}

func fetchUserChoice(attachmentAction webexapi.AttachmentAction) (string, error) {
	type attachmentActionInputs struct {
		Data string `json:"data"`
	}

	inputs := new(attachmentActionInputs)
	err := json.Unmarshal(attachmentAction.Inputs, inputs)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal the attachment action inputs : %w", err)
	}
	return inputs.Data, nil
}

func isOneOf(optionId string, options []Option) bool {
	for _, option := range options {
		if option.Id == optionId {
			return true
		}
	}
	return false
}
