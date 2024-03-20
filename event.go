package webexbot

import "github.com/pavelzagorodnyuk/webexbot/internal/webexapi/v1"

// Event describes an event which happened on the Webex side and is directly or indirectly related to a bot
type Event struct {
	// The identifier of the person who initiated the event
	InitiatorId string

	// The email of the person who initiated the event
	InitiatorEmail string

	// The identifier of the room where the event occurred
	// NOTE: this field is empty in events related to attachment actions
	RoomId string

	// The type of the room where the event occurred
	// NOTE: this field is empty in events related to attachment actions
	RoomType webexapi.RoomType

	// The resource instance which the event occurred with
	Resource any

	// The kind of the resource
	ResourceKind webexapi.ResourceKind

	// The kind of the event which occurred with the resource
	ResourceEvent webexapi.ResourceEvent
}

const (
	RoomTypeDirect = webexapi.RoomTypeDirect
	RoomTypeGroup  = webexapi.RoomTypeGroup
)

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
