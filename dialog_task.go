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

// DialogTaskProviderFunc allows to use ordinary functions which implements DialogTaskProvider.ProvideFor method as
// dialog task providers
type DialogTaskProviderFunc func(Event) DialogTask

func (p DialogTaskProviderFunc) ProvideFor(event Event) DialogTask {
	return p(event)
}

type combinedDialogTaskProvider struct {
	subproviders []DialogTaskProvider
}

// NewCombinedDialogTaskProvider creates a new dialog task provider which combines the specified dialog task providers.
// It calls each subprovider and returns the first received dialog task. If all subproviders return nil, it returns nil
// too.
func NewCombinedDialogTaskProvider(subproviders ...DialogTaskProvider) DialogTaskProvider {
	return &combinedDialogTaskProvider{
		subproviders: subproviders,
	}
}

func (p combinedDialogTaskProvider) ProvideFor(event Event) DialogTask {
	for _, subprovider := range p.subproviders {
		triggeredTask := subprovider.ProvideFor(event)
		if triggeredTask != nil {
			return triggeredTask
		}
	}
	return nil
}
