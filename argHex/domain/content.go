package domain

// Content lifecycle status. Archived is schema-only for now — no transition
// wires to it yet, but published/read filtering already understands it.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// Entity types tag revisions and activity-log entries so the generic stores
// stay content-agnostic.
const (
	EntityProject    = "project"
	EntityNote       = "note"
	EntityHobby      = "hobby"
	EntitySuggestion = "suggestion"
	EntityCopy       = "copy"
	EntityLantern    = "lantern"
	EntityMedia      = "media"
	EntityFigurehead = "figurehead"
)
