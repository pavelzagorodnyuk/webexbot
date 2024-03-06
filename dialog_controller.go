package webexbot

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type dialogController struct {
	incomingEvents     <-chan Event
	dialogProvider     dialogProvider
	dialogTaskProvider DialogTaskProvider
	activeDialogs      map[dialogKey]dialogReferences
	completionSignals  chan completionSignal
}

type dialogKey struct {
	personId, roomId string
}

type dialogReferences struct {
	context    context.Context
	eventChan  chan<- Event
	cancelFunc context.CancelFunc
}

type completionSignal struct {
	key            dialogKey
	executionError error
}

func newDialogController(
	incomingEvents <-chan Event,
	dialogProvider dialogProvider,
	dialogTaskProvider DialogTaskProvider,
) dialogController {
	return dialogController{
		incomingEvents:     incomingEvents,
		dialogProvider:     dialogProvider,
		dialogTaskProvider: dialogTaskProvider,
		activeDialogs:      make(map[dialogKey]dialogReferences),
		completionSignals:  make(chan completionSignal),
	}
}

func (c *dialogController) run(ctx context.Context) {
	defer c.stopAllDialogs()

	for {
		select {
		case nextEvent := <-c.incomingEvents:
			c.processEvent(ctx, nextEvent)

		case nextSignal := <-c.completionSignals:
			c.processCompletionSignal(ctx, nextSignal)

		case <-ctx.Done():
			return
		}
	}
}

func (c *dialogController) processEvent(ctx context.Context, event Event) {
	triggeredTask := c.dialogTaskProvider.ProvideFor(event)

	key := dialogKey{
		personId: event.InitiatorId,
		roomId:   event.RoomId,
	}

	isActive := c.isDialogActive(key)

	switch {
	case isActive && triggeredTask != nil:
		c.stopDialog(key)
		fallthrough

	case !isActive && triggeredTask != nil:
		c.startDialog(ctx, key, triggeredTask)
		fallthrough

	case isActive && triggeredTask == nil:
		c.pushEvent(ctx, key, event)
	}
}

func (c *dialogController) isDialogActive(key dialogKey) bool {
	_, isActive := c.activeDialogs[key]
	return isActive
}

func (c *dialogController) startDialog(ctx context.Context, key dialogKey, task DialogTask) {
	dialog, eventChan, cancelFunc := c.dialogProvider.provideFor(ctx, key.personId, key.roomId)

	go dialogTaskRoutine(dialog, task, c.completionSignals, cancelFunc)

	c.activeDialogs[key] = dialogReferences{
		context:    dialog,
		eventChan:  eventChan,
		cancelFunc: cancelFunc,
	}
}

func dialogTaskRoutine(
	dialog Dialog,
	task DialogTask,
	completionSignals chan<- completionSignal,
	cancelFunc context.CancelFunc,
) {
	var err error
	defer func() {
		select {
		// do not send a completion signal if the dialog has been stopped from the outside
		case <-dialog.Done():
			return

		default:
			cancelFunc()
		}

		if panicMessage := recover(); panicMessage != nil {
			err = fmt.Errorf("the dialog is recovered from panic : %v", panicMessage)
		}

		completionSignals <- completionSignal{
			key: dialogKey{
				personId: dialog.PersonId(),
				roomId:   dialog.RoomId(),
			},
			executionError: err,
		}
	}()

	err = task.Talk(dialog)
}

func (c *dialogController) pushEvent(ctx context.Context, key dialogKey, event Event) {
	const pushingTimeout = 10 * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, pushingTimeout)
	defer cancel()

	dialogContext := c.activeDialogs[key].context
	eventChan := c.activeDialogs[key].eventChan

	select {
	case eventChan <- event:
		return

	// do not push the event if the dialog is completed
	case <-dialogContext.Done():
		return

	case <-ctx.Done():
		slog.WarnContext(ctx, "unable to push the event to the dialog task because the pushing timeout has expired")
		return
	}
}

func (c *dialogController) stopDialog(key dialogKey) {
	dialog := c.activeDialogs[key]

	select {
	// the dialog is already completed
	case <-dialog.context.Done():
		return

	default:
		dialog.cancelFunc()
		close(dialog.eventChan)
	}
}

func (c *dialogController) stopAllDialogs() {
	for key := range c.activeDialogs {
		c.stopDialog(key)
	}
}

func (c *dialogController) processCompletionSignal(ctx context.Context, signal completionSignal) {
	if signal.executionError != nil {
		slog.ErrorContext(ctx, "the following error occurred during the execution of the dialog task : %w",
			signal.executionError)
	}

	dialog := c.activeDialogs[signal.key]
	close(dialog.eventChan)
	delete(c.activeDialogs, signal.key)
}
