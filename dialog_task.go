package webexbot

// DialogTask defines the behavior of a bot in a dialog with a user
type DialogTask interface {
	Talk(Dialog) error
}

// Provides dialog tasks for the event processing
type DialogTaskProvider interface {
	// Provides a dialog task which is triggered by the event or returns nil if the event does not trigger any provider
	// dialog task
	ProvideFor(Event) DialogTask
}
