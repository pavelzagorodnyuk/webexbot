package webexbot

import (
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/pavelzagorodnyuk/webexbot/internal/webexapi/v1"
)

type BotBuilder struct {
	hostname           string
	port               int
	webexToken         string
	dialogTaskProvider DialogTaskProvider
	tlsConfig          *tls.Config
	webhookSecret      string
	eventFilters       []EventFilter
}

func New(hostname string, port int, webexToken string) BotBuilder {
	return BotBuilder{
		hostname:   hostname,
		port:       port,
		webexToken: webexToken,
	}
}

func (b BotBuilder) SetDialogTaskProvider(provider DialogTaskProvider) BotBuilder {
	b.dialogTaskProvider = provider
	return b
}

func (b BotBuilder) SetTLSConfig(config *tls.Config) BotBuilder {
	b.tlsConfig = config
	return b
}

func (b BotBuilder) SetWebhookSecret(secret string) BotBuilder {
	b.webhookSecret = secret
	return b
}

func (b BotBuilder) SetEventFilters(filters ...EventFilter) BotBuilder {
	b.eventFilters = filters
	return b
}

const eventChannelBufferSize = 3

func (b BotBuilder) Build() (Bot, error) {
	err := b.validateConfiguration()
	if err != nil {
		return Bot{}, fmt.Errorf("the bot configuration is not valid : %w", err)
	}

	webexClient := b.createWebexClient()
	eventChan := make(chan Event, eventChannelBufferSize)

	webexEventProducer := newWebexEventProducer(
		b.hostname,
		b.port,
		webexClient,
		eventChan,
		b.webhookSecret,
		b.tlsConfig,
		b.eventFilters,
	)

	dialogController := newDialogController(
		eventChan,
		dialogProvider{},
		b.dialogTaskProvider,
	)

	return Bot{
		webexEventProducer: webexEventProducer,
		dialogController:   dialogController,
	}, nil
}

func (b BotBuilder) validateConfiguration() error {
	if len(b.hostname) == 0 {
		return errors.New("host name must be specified")
	}

	if len(b.webexToken) == 0 {
		return errors.New("webex token must be specified")
	}

	if b.dialogTaskProvider == nil {
		return errors.New("dialog task provider must be specified")
	}

	return nil
}

func (b BotBuilder) createWebexClient() webexapi.Client {
	var webexClientOptions []webexapi.Option

	if b.tlsConfig != nil {
		webexClientOptions = append(webexClientOptions, webexapi.WithTLSConfig(b.tlsConfig))
	}

	return webexapi.NewClient(b.webexToken, webexClientOptions...)
}
