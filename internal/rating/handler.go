package rating

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"project/internal/platform"
	"project/internal/platform/compiled"
	"project/internal/platform/humax"
)

const pgErrUniqueViolation = "23505"

type Handler struct {
	deps platform.Deps
}

func newHandler(deps platform.Deps) *Handler {
	return &Handler{deps: deps}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation
}

func (h *Handler) handleCreateEvent(ctx context.Context, input *CreateEventInput) (*CreateEventOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	req := input.Body
	if strings.TrimSpace(req.Name) == "" {
		return nil, humax.BadRequest("name is required")
	}
	if req.Visibility != "public" && req.Visibility != "private" {
		return nil, humax.BadRequest("visibility must be public or private")
	}

	tx, err := h.deps.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := compiled.New(tx)

	var desc pgtype.Text
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}

	eventRow, err := q.RatingCreateEvent(ctx, compiled.RatingCreateEventParams{
		CompanyID:   p.SelectedCompanyID,
		HostID:      p.UserID,
		Name:        req.Name,
		Description: desc,
		Visibility:  req.Visibility,
	})
	if err != nil {
		return nil, err
	}

	var cycles []CycleInfo
	if len(req.Cycles) > 0 {
		names := make([]string, len(req.Cycles))
		for i, c := range req.Cycles {
			names[i] = c.Name
		}
		rows, err := q.RatingBulkInsertCycles(ctx, compiled.RatingBulkInsertCyclesParams{
			EventID: eventRow.ID,
			Column2: names,
		})
		if err != nil {
			return nil, err
		}
		cycles = make([]CycleInfo, len(rows))
		for i, r := range rows {
			cycles[i] = CycleInfo{
				ID:         r.ID,
				EventID:    r.EventID,
				Name:       r.Name,
				OrderIndex: r.OrderIndex,
			}
		}
	}

	var forms []FormInfo
	if len(req.Forms) > 0 {
		types := make([]string, len(req.Forms))
		labels := make([]string, len(req.Forms))
		for i, f := range req.Forms {
			types[i] = f.Type
			labels[i] = f.Label
		}
		rows, err := q.RatingBulkInsertForms(ctx, compiled.RatingBulkInsertFormsParams{
			EventID: eventRow.ID,
			Column2: types,
			Column3: labels,
		})
		if err != nil {
			return nil, err
		}
		forms = make([]FormInfo, len(rows))
		for i, r := range rows {
			forms[i] = FormInfo{
				ID:         r.ID,
				EventID:    r.EventID,
				Type:       r.Type,
				Label:      r.Label,
				OrderIndex: r.OrderIndex,
			}
		}
	}

	if req.Visibility == "private" && len(req.Members) > 0 {
		for _, email := range req.Members {
			email = strings.ToLower(strings.TrimSpace(email))
			if email == "" {
				continue
			}
			userRow, err := q.RatingFindUserByEmail(ctx, email)
			var userID int64
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return nil, err
				}
				created, err := q.RatingCreateUser(ctx, compiled.RatingCreateUserParams{
					Email: email,
					Name:  email,
				})
				if err != nil {
					return nil, err
				}
				userID = created.ID
			} else {
				userID = userRow.ID
			}
			if err := q.RatingAddEventMember(ctx, compiled.RatingAddEventMemberParams{
				EventID: eventRow.ID,
				UserID:  userID,
			}); err != nil && !isUniqueViolation(err) {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &CreateEventOutput{
		Body: EventDetail{
			EventInfo: eventInfoFromRow(eventRow),
			Cycles:    cycles,
			Forms:     forms,
		},
	}, nil
}

func (h *Handler) handleListEvents(ctx context.Context, _ *struct{}) (*ListEventsOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := h.deps.Queries.RatingListEvents(ctx, compiled.RatingListEventsParams{
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		return nil, err
	}

	events := make([]EventInfo, 0, len(rows))
	for _, r := range rows {
		events = append(events, eventInfoFromRow(r))
	}

	out := &ListEventsOutput{}
	out.Body.Events = events
	return out, nil
}

func (h *Handler) handleGetEvent(ctx context.Context, input *GetEventInput) (*GetEventOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEvent(ctx, compiled.RatingGetEventParams{
		ID:        input.ID,
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}

	cycleRows, err := h.deps.Queries.RatingGetEventCycles(ctx, eventRow.ID)
	if err != nil {
		return nil, err
	}
	cycles := make([]CycleInfo, len(cycleRows))
	for i, r := range cycleRows {
		cycles[i] = CycleInfo{
			ID:         r.ID,
			EventID:    r.EventID,
			Name:       r.Name,
			OrderIndex: r.OrderIndex,
		}
	}

	formRows, err := h.deps.Queries.RatingGetEventForms(ctx, eventRow.ID)
	if err != nil {
		return nil, err
	}
	forms := make([]FormInfo, len(formRows))
	for i, r := range formRows {
		forms[i] = FormInfo{
			ID:         r.ID,
			EventID:    r.EventID,
			Type:       r.Type,
			Label:      r.Label,
			OrderIndex: r.OrderIndex,
		}
	}

	return &GetEventOutput{
		Body: EventDetail{
			EventInfo: eventInfoFromRow(eventRow),
			Cycles:    cycles,
			Forms:     forms,
		},
	}, nil
}

func (h *Handler) handleAddMember(ctx context.Context, input *AddMemberInput) (*AddMemberOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEvent(ctx, compiled.RatingGetEventParams{
		ID:        input.EventID,
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.HostID != p.UserID {
		return nil, humax.Forbidden("only the event host can manage members")
	}

	email := strings.ToLower(strings.TrimSpace(input.Body.Email))

	userRow, err := h.deps.Queries.RatingFindUserByEmail(ctx, email)
	var userID int64
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		name := input.Body.Name
		if name == "" {
			name = email
		}
		created, err := h.deps.Queries.RatingCreateUser(ctx, compiled.RatingCreateUserParams{
			Email: email,
			Name:  name,
		})
		if err != nil {
			return nil, err
		}
		userID = created.ID
	} else {
		userID = userRow.ID
	}

	if err := h.deps.Queries.RatingAddEventMember(ctx, compiled.RatingAddEventMemberParams{
		EventID: input.EventID,
		UserID:  userID,
	}); err != nil {
		if isUniqueViolation(err) {
			return nil, humax.Conflict("already a member")
		}
		return nil, err
	}

	return &AddMemberOutput{}, nil
}

func (h *Handler) handleRemoveMember(ctx context.Context, input *RemoveMemberInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEvent(ctx, compiled.RatingGetEventParams{
		ID:        input.EventID,
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.HostID != p.UserID {
		return nil, humax.Forbidden("only the event host can manage members")
	}

	if err := h.deps.Queries.RatingRemoveEventMember(ctx, compiled.RatingRemoveEventMemberParams{
		EventID: input.EventID,
		UserID:  input.UserID,
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

// eventInfoFromRow maps a compiled.Event to EventInfo.
func eventInfoFromRow(r compiled.Event) EventInfo {
	var desc string
	if r.Description.Valid {
		desc = r.Description.String
	}
	return EventInfo{
		ID:           r.ID,
		CompanyID:    r.CompanyID,
		HostID:       r.HostID,
		Name:         r.Name,
		Description:  desc,
		Visibility:   r.Visibility,
		Status:       r.Status,
		CurrentStage: r.CurrentStage,
		CreatedAt:    r.CreatedAt.Time.UTC(),
		UpdatedAt:    r.UpdatedAt.Time.UTC(),
	}
}

