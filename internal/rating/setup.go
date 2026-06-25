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
}
