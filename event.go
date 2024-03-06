package webexbot

import "github.com/pavelzagorodnyuk/webexbot/internal/webexapi/v1"

// Event describes an event which happened on the Webex side and is directly or indirectly related to a bot
type Event struct {
	// The identifier of the person who initiated the event
	InitiatorId string

	// The email of the person who initiated the event
	InitiatorEmail string

	// The identifier of the room where the event occurred
	RoomId string

	// The resource instance which the event occurred with
	Resource any

	// The kind of the resource
	ResourceKind webexapi.ResourceKind

	// The kind of the event which occurred with the resource
	ResourceEvent webexapi.ResourceEvent
}

// ResourceKind represents the kind of a resource (for example, messages, rooms, meetings, etc.)
type ResourceKind webexapi.ResourceKind

const (
	AttachmentActions   = webexapi.AttachmentActions
	Memberships         = webexapi.Memberships
	Messages            = webexapi.Messages
	Rooms               = webexapi.Rooms
	Meetings            = webexapi.Meetings
	Recordings          = webexapi.Recordings
	MeetingParticipants = webexapi.MeetingParticipants
	MeetingTranscripts  = webexapi.MeetingTranscripts
)

// ResourceEvent is an event which can happen to a resource (a resource can be created, updated, deleted, etc.)
type ResourceEvent webexapi.ResourceEvent

const (
	ResourceCreated  = webexapi.ResourceCreated
	ResourceUpdated  = webexapi.ResourceUpdated
	ResourceDeleted  = webexapi.ResourceDeleted
	ResourceStarted  = webexapi.ResourceStarted
	ResourceEnded    = webexapi.ResourceEnded
	ResourceJoined   = webexapi.ResourceJoined
	ResourceLeft     = webexapi.ResourceLeft
	ResourceMigrated = webexapi.ResourceMigrated
	AllEvents        = webexapi.AllEvents
)
