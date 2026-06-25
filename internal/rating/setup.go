package rating

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"project/internal/platform"
	"project/internal/platform/humax"
)

func Setup(_ context.Context, deps platform.Deps) {
	h := newHandler(deps)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "create-event",
		Method:        http.MethodPost,
		Path:          "/api/events",
		Summary:       "Create a rating event",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleCreateEvent)

	huma.Register(deps.API, huma.Operation{
		OperationID: "list-events",
		Method:      http.MethodGet,
		Path:        "/api/events",
		Summary:     "List rating events",
		Tags:        []string{"rating"},
		Security:    humax.BearerAuth(),
	}, h.handleListEvents)

	huma.Register(deps.API, huma.Operation{
		OperationID: "get-event",
		Method:      http.MethodGet,
		Path:        "/api/events/{id}",
		Summary:     "Get a rating event",
		Tags:        []string{"rating"},
		Security:    humax.BearerAuth(),
	}, h.handleGetEvent)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "add-event-member",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/members",
		Summary:       "Add a member to a private event",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleAddMember)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "remove-event-member",
		Method:        http.MethodDelete,
		Path:          "/api/events/{id}/members/{userId}",
		Summary:       "Remove a member from a private event",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleRemoveMember)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "activate-event",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/activate",
		Summary:       "Activate an event",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleActivate)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "start-cycle",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/start-cycle",
		Summary:       "Start a cycle",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleStartCycle)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "show-form",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/show-form",
		Summary:       "Open the rating form for the current cycle",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleShowForm)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "next-cycle",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/next-cycle",
		Summary:       "Advance to the next cycle",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleNextCycle)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "end-event",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/end",
		Summary:       "End an event",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusNoContent,
	}, h.handleEndEvent)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "join-event",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/join",
		Summary:       "Join a rating event as a rater",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleJoin)

	huma.Register(deps.API, huma.Operation{
		OperationID: "get-event-session",
		Method:      http.MethodGet,
		Path:        "/api/events/{id}/session",
		Summary:     "Get current session state for rater polling",
		Tags:        []string{"rating"},
		Security:    humax.BearerAuth(),
	}, h.handleGetSession)

	huma.Register(deps.API, huma.Operation{
		OperationID:   "respond-to-cycle",
		Method:        http.MethodPost,
		Path:          "/api/events/{id}/respond",
		Summary:       "Submit rater responses for the current cycle",
		Tags:          []string{"rating"},
		Security:      humax.BearerAuth(),
		DefaultStatus: http.StatusCreated,
	}, h.handleRespond)
}
