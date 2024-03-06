package webexbot

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pavelzagorodnyuk/webexbot/internal/webexapi/v1"
)

// A webexEventProducer produces events which are based on Webex webhook callbacks
type webexEventProducer struct {
	authenticationKey string
	webexClient       webexapi.Client
	webhookServer     *http.Server
}

func newWebexEventProducer(
	addr string,
	webexClient webexapi.Client,
	outgoingEvents chan<- Event,
	authenticationKey string,
	tlsConfig *tls.Config,
	filters []EventFilter,
) webexEventProducer {
	mux := http.NewServeMux()
	webhookHandler := newWebhookHandler(authenticationKey, webexClient, filters, outgoingEvents)
	mux.Handle("POST "+webhookHandlerPath, webhookHandler)

	// TODO: should we validate addr here?

	return webexEventProducer{
		authenticationKey: authenticationKey,
		webexClient:       webexClient,
		webhookServer: &http.Server{
			Addr:      addr,
			Handler:   mux,
			TLSConfig: tlsConfig,
		},
	}
}

func (p webexEventProducer) run(ctx context.Context) error {
	err := p.createWebhooks(ctx)
	if err != nil {
		return err
	}

	// schedule the server to close when the context is done
	context.AfterFunc(ctx, func() {
		_ = p.webhookServer.Close()
	})

	return p.webhookServer.ListenAndServe()
}

func (p webexEventProducer) createWebhooks(ctx context.Context) error {
	err := p.createMessageWebhook(ctx)
	if err != nil {
		return fmt.Errorf("unable to create a webhook for messages : %w", err)
	}

	err = p.createAttachmentActionWebhook(ctx)
	if err != nil {
		return fmt.Errorf("unable to create a webhook for attachment actions : %w", err)
	}
	return nil
}

func (p webexEventProducer) createMessageWebhook(ctx context.Context) error {
	request := webexapi.CreateWebhookRequest{
		Name:      "Message webhook [webexbot]",
		TargetUrl: p.webhookHandlerURL(),
		Resource:  webexapi.Messages,
		Event:     webexapi.ResourceCreated,
		Secret:    p.authenticationKey,
	}

	_, webexErr, err := p.webexClient.CreateWebhook(ctx, request)
	if err != nil {
		return err
	}
	// the HTTP status conflict means that the webhook already exists
	if webexErr != nil && webexErr.StatusCode != http.StatusConflict {
		return webexErr
	}
	return nil
}

func (p webexEventProducer) createAttachmentActionWebhook(ctx context.Context) error {
	request := webexapi.CreateWebhookRequest{
		Name:      "Attachment action webhook [webexbot]",
		TargetUrl: p.webhookHandlerURL(),
		Resource:  webexapi.AttachmentActions,
		Event:     webexapi.ResourceCreated,
		Secret:    p.authenticationKey,
	}

	_, webexErr, err := p.webexClient.CreateWebhook(ctx, request)
	if err != nil {
		return err
	}
	// the HTTP status conflict means that the webhook already exists
	if webexErr != nil && webexErr.StatusCode != http.StatusConflict {
		return webexErr
	}
	return nil
}

const webhookHandlerPath = "/webhooks"

func (p webexEventProducer) webhookHandlerURL() string {
	return fmt.Sprintf("https://%s%s", p.webhookServer.Addr, webhookHandlerPath)
}

// webhookHandler is an HTTP handler which serves webhook callbacks and creates a new Event for each of them. Those
// events which match the handler filters are sent into the channel for outgoing events.
type webhookHandler struct {
	authenticationKey string
	webexClient       webexapi.Client
	filters           []EventFilter
	outgoingEvents    chan<- Event
}

// EventFilter is a function that decides which events must be processed and which must be skipped. An event matches
// the filter and must be processed if the filter call for this event returns true.
type EventFilter func(Event) bool

func newWebhookHandler(
	authenticationKey string,
	webexClient webexapi.Client,
	filters []EventFilter,
	outgoingEvents chan<- Event,
) webhookHandler {
	return webhookHandler{
		authenticationKey: authenticationKey,
		webexClient:       webexClient,
		filters:           filters,
		outgoingEvents:    outgoingEvents,
	}
}

func (h webhookHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	err := h.authenticate(request)
	if err != nil {
		response.WriteHeader(http.StatusUnauthorized)
		slog.Error("the request cannot be processed because its sender has not been authenticated : %w", err)
		return
	}

	callback := new(webexapi.WebhookCallback)
	err = json.NewDecoder(request.Body).Decode(callback)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		slog.Error("unable to decode the request body as a webhook callback : %w", err)
		return
	}

	ctx := request.Context()

	event, statusCode, err := h.prepareEvent(ctx, *callback)
	if err != nil {
		response.WriteHeader(statusCode)
		slog.Error("unable to prepare an event based on the webhook callback : %w", err)
		return
	}

	matches := h.matchesFilters(event)
	if !matches {
		response.WriteHeader(http.StatusOK)
		slog.Info("skipping the event because it does not match the filters")
		return
	}

	isSuccessful := h.enqueue(ctx, event)
	if isSuccessful {
		response.WriteHeader(http.StatusAccepted)
	} else {
		response.WriteHeader(http.StatusTooManyRequests)
		slog.Error("unable to enqueue the event because the queue is full")
	}
}

func (h webhookHandler) authenticate(request *http.Request) error {
	// MAC is a message authentication code which is used to verify both the data integrity and authenticity of a
	// message. A message is considered verified if the both provided and locally computed MACs are equal.
	providedMAC := h.fetchMAC(request)
	if providedMAC == nil {
		return errors.New("the request does not include a message authentication code")
	}

	actualMAC, err := h.computeMAC(request)
	if err != nil {
		return fmt.Errorf("unable to compute the actual message authentication code : %w", err)
	}

	if !hmac.Equal(actualMAC, providedMAC) {
		return fmt.Errorf("actual and provided message authentication codes are different")
	}
	return nil
}

func (h webhookHandler) fetchMAC(request *http.Request) []byte {
	// X-Spark-Signature is an HTTP header which contains a hash-based message authentication code
	mac := request.Header.Get("X-Spark-Signature")
	if len(mac) == 0 {
		return nil
	}
	return []byte(mac)
}

func (h webhookHandler) computeMAC(request *http.Request) ([]byte, error) {
	body, err := request.GetBody()
	if err != nil {
		return nil, fmt.Errorf("unable to get the request body : %w", err)
	}

	hmacEncoder := hmac.New(sha1.New, []byte(h.authenticationKey))
	_, err = io.Copy(hmacEncoder, body)
	if err != nil {
		return nil, err
	}

	return hmacEncoder.Sum(nil), nil
}

// prepareEvent prepares a new Event based on the passed webhook callback
func (h webhookHandler) prepareEvent(
	ctx context.Context,
	callback webexapi.WebhookCallback,
) (
	event Event,
	statusCode int,
	err error,
) {
	resource, statusCode, err := h.prepareEventResource(ctx, callback)
	if err != nil {
		return Event{}, statusCode, fmt.Errorf("unable to prepare the resource : %w", err)
	}

	eventId, err := uuid.NewRandom()
	if err != nil {
		return Event{}, http.StatusInternalServerError,
			fmt.Errorf("unable to generate an identifier for the event : %w", err)
	}

	initiatorEmail := h.fetchInitiatorEmail(resource)
	roomId := h.fetchRoomId(resource)

	return Event{
		Id:             eventId.String(),
		InitiatorId:    callback.ActorId,
		InitiatorEmail: initiatorEmail,
		RoomId:         roomId,
		Resource:       resource,
		ResourceKind:   callback.Resource,
		ResourceEvent:  callback.Event,
	}, 0, nil
}

func (h webhookHandler) prepareEventResource(
	ctx context.Context,
	callback webexapi.WebhookCallback,
) (
	resource any,
	statusCode int,
	err error,
) {
	resourceId, err := h.encodeResourceId(callback.Data)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("unable to get the resource identifier : %w", err)
	}

	switch callback.Resource {
	case webexapi.Messages:
		message, webexErr, err := h.webexClient.GetMessage(ctx, resourceId)
		const unableToGetMessageById = "unable to get a message by the identifier '%s' : %w"
		if err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf(unableToGetMessageById, resourceId, err)
		}
		if webexErr != nil {
			return nil, http.StatusFailedDependency,
				fmt.Errorf(unableToGetMessageById, resourceId, webexErr)
		}

		return message, 0, nil

	case webexapi.AttachmentActions:
		attachmentAction, webexErr, err := h.webexClient.GetAttachmentAction(ctx, resourceId)
		const unableToGetAttachmentActionById = "unable to get an attachment action by the identifier '%s' : %w"
		if err != nil {
			return nil, http.StatusInternalServerError,
				fmt.Errorf(unableToGetAttachmentActionById, resourceId, err)
		}
		if webexErr != nil {
			return nil, http.StatusFailedDependency,
				fmt.Errorf(unableToGetAttachmentActionById, resourceId, webexErr)
		}

		return attachmentAction, 0, nil

	default:
		return nil, 0, nil
	}
}

func (h webhookHandler) encodeResourceId(rawResource json.RawMessage) (string, error) {
	type resourceWithId struct {
		Id string `json:"id"`
	}

	resource := new(resourceWithId)
	err := json.Unmarshal(rawResource, resource)
	return resource.Id, err
}

func (h webhookHandler) fetchInitiatorEmail(resource any) string {
	switch v := resource.(type) {
	case webexapi.Message:
		return v.PersonEmail

	// TODO: implement fetching for attachment actions

	default:
		return ""
	}
}

func (h webhookHandler) fetchRoomId(resource any) string {
	switch v := resource.(type) {
	case webexapi.Message:
		return v.RoomId

	case webexapi.AttachmentAction:
		return v.RoomId

	default:
		return ""
	}
}

func (h webhookHandler) matchesFilters(event Event) bool {
	for _, filter := range h.filters {
		matches := filter(event)
		if !matches {
			return false
		}
	}
	return true
}

func (h webhookHandler) enqueue(ctx context.Context, event Event) (isSuccessful bool) {
	for {
		select {
		case h.outgoingEvents <- event:
			return true

		case <-ctx.Done():
			return false

		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}
