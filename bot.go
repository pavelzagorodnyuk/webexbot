package webexbot

import "context"

type Bot struct {
	webexEventProducer webexEventProducer
	dialogController   dialogController
}

func (b Bot) ListenAndTalk(ctx context.Context) error {
	ctx, cancelWithCause := context.WithCancelCause(ctx)

	go func() {
		err := b.webexEventProducer.run(ctx)
		if err != nil {
			cancelWithCause(err)
		}
	}()

	b.dialogController.run(ctx)

	return context.Cause(ctx)
}
