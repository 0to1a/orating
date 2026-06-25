package rating

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
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

func (h *Handler) handleActivate(ctx context.Context, input *EventIDInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	_, err = h.deps.Queries.RatingActivateEvent(ctx, compiled.RatingActivateEventParams{
		ID:        input.ID,
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("event not found or already active")
		}
		return nil, err
	}

	return nil, nil
}

func (h *Handler) handleStartCycle(ctx context.Context, input *StartCycleInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	_, err = h.deps.Queries.RatingGetCycleByEvent(ctx, compiled.RatingGetCycleByEventParams{
		ID:      input.Body.CycleID,
		EventID: input.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("cycle does not belong to this event")
		}
		return nil, err
	}

	_, err = h.deps.Queries.RatingStartCycle(ctx, compiled.RatingStartCycleParams{
		ID:            input.ID,
		CompanyID:     p.SelectedCompanyID,
		HostID:        p.UserID,
		ActiveCycleID: pgtype.Int8{Int64: input.Body.CycleID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("event not active")
		}
		return nil, err
	}

	return nil, nil
}

func (h *Handler) handleShowForm(ctx context.Context, input *EventIDInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	_, err = h.deps.Queries.RatingShowForm(ctx, compiled.RatingShowFormParams{
		ID:        input.ID,
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("event not in waiting stage")
		}
		return nil, err
	}

	return nil, nil
}

func (h *Handler) handleNextCycle(ctx context.Context, input *NextCycleInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	_, err = h.deps.Queries.RatingGetCycleByEvent(ctx, compiled.RatingGetCycleByEventParams{
		ID:      input.Body.CycleID,
		EventID: input.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("cycle does not belong to this event")
		}
		return nil, err
	}

	_, err = h.deps.Queries.RatingNextCycle(ctx, compiled.RatingNextCycleParams{
		ID:            input.ID,
		CompanyID:     p.SelectedCompanyID,
		HostID:        p.UserID,
		ActiveCycleID: pgtype.Int8{Int64: input.Body.CycleID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("event not active")
		}
		return nil, err
	}

	return nil, nil
}

func (h *Handler) handleEndEvent(ctx context.Context, input *EventIDInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	_, err = h.deps.Queries.RatingEndEvent(ctx, compiled.RatingEndEventParams{
		ID:        input.ID,
		CompanyID: p.SelectedCompanyID,
		HostID:    p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("event not active")
		}
		return nil, err
	}

	return nil, nil
}

func (h *Handler) handleJoin(ctx context.Context, input *JoinEventInput) (*JoinEventOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEventByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.CompanyID != p.SelectedCompanyID {
		return nil, humax.NotFound("event not found")
	}
	if eventRow.Status != "active" {
		return nil, humax.Unprocessable("event is not active")
	}
	if eventRow.Visibility == "private" {
		_, err := h.deps.Queries.RatingGetEventMemberCheck(ctx, compiled.RatingGetEventMemberCheckParams{
			EventID: input.ID,
			UserID:  p.UserID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, humax.Forbidden("not invited to this private event")
			}
			return nil, err
		}
	}

	_, err = h.deps.Queries.RatingJoinEvent(ctx, compiled.RatingJoinEventParams{
		EventID: input.ID,
		UserID:  p.UserID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	participant, err := h.deps.Queries.RatingGetParticipant(ctx, compiled.RatingGetParticipantParams{
		EventID: input.ID,
		UserID:  p.UserID,
	})
	if err != nil {
		return nil, err
	}

	return &JoinEventOutput{
		Body: ParticipantInfo{
			ID:       participant.ID,
			EventID:  participant.EventID,
			UserID:   participant.UserID,
			JoinedAt: participant.JoinedAt.Time.UTC(),
		},
	}, nil
}

func (h *Handler) handleGetSession(ctx context.Context, input *GetSessionInput) (*GetSessionOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEventByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.CompanyID != p.SelectedCompanyID {
		return nil, humax.NotFound("event not found")
	}

	resp := SessionResponse{
		CurrentStage: eventRow.CurrentStage,
		Forms:        []FormInfo{},
	}

	if eventRow.ActiveCycleID.Valid {
		cycleID := eventRow.ActiveCycleID.Int64
		resp.ActiveCycleID = &cycleID

		cycleRow, err := h.deps.Queries.RatingGetCycleByID(ctx, cycleID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		if err == nil {
			resp.ActiveCycleName = cycleRow.Name
		}
	}

	if eventRow.CurrentStage == "form_open" {
		formRows, err := h.deps.Queries.RatingGetFormsForEvent(ctx, input.ID)
		if err != nil {
			return nil, err
		}
		forms := make([]FormInfo, len(formRows))
		for i, f := range formRows {
			forms[i] = FormInfo{
				ID:         f.ID,
				EventID:    f.EventID,
				Type:       f.Type,
				Label:      f.Label,
				OrderIndex: f.OrderIndex,
			}
		}
		resp.Forms = forms
	}

	participant, participantErr := h.deps.Queries.RatingGetParticipant(ctx, compiled.RatingGetParticipantParams{
		EventID: input.ID,
		UserID:  p.UserID,
	})
	if participantErr == nil {
		resp.IsParticipant = true
		if eventRow.ActiveCycleID.Valid {
			_, err2 := h.deps.Queries.RatingGetResponseForParticipantCycle(ctx, compiled.RatingGetResponseForParticipantCycleParams{
				CycleID:       eventRow.ActiveCycleID.Int64,
				ParticipantID: participant.ID,
			})
			if err2 == nil {
				resp.MyResponseSubmitted = true
			}
		}
	}

	return &GetSessionOutput{Body: resp}, nil
}

func (h *Handler) handleRespond(ctx context.Context, input *RespondInput) (*struct{}, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEventByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.CompanyID != p.SelectedCompanyID {
		return nil, humax.NotFound("event not found")
	}
	if eventRow.Status != "active" {
		return nil, humax.Unprocessable("event is not active")
	}
	if eventRow.CurrentStage != "form_open" {
		return nil, humax.Unprocessable("form is not open")
	}
	if !eventRow.ActiveCycleID.Valid {
		return nil, humax.Unprocessable("no active cycle")
	}

	participant, err := h.deps.Queries.RatingGetParticipant(ctx, compiled.RatingGetParticipantParams{
		EventID: input.ID,
		UserID:  p.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.Unprocessable("must join event first")
		}
		return nil, err
	}

	_, err = h.deps.Queries.RatingGetResponseForParticipantCycle(ctx, compiled.RatingGetResponseForParticipantCycleParams{
		CycleID:       eventRow.ActiveCycleID.Int64,
		ParticipantID: participant.ID,
	})
	if err == nil {
		return nil, humax.Conflict("already responded for this cycle")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	for _, item := range input.Body.Items {
		formRow, err := h.deps.Queries.RatingGetFormByID(ctx, item.FormID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, humax.BadRequest("form not found")
			}
			return nil, err
		}
		if formRow.EventID != input.ID {
			return nil, humax.BadRequest("form does not belong to this event")
		}
		switch formRow.Type {
		case "rating":
			if item.ValueNumber == nil || *item.ValueNumber < 1 || *item.ValueNumber > 5 {
				return nil, humax.BadRequest("rating value must be between 1 and 5")
			}
		case "mood":
			if item.ValueNumber == nil || *item.ValueNumber < 1 || *item.ValueNumber > 4 {
				return nil, humax.BadRequest("mood value must be between 1 and 4")
			}
		case "free_text":
			if item.ValueText == nil || *item.ValueText == "" {
				return nil, humax.BadRequest("free text value is required")
			}
		}
	}

	tx, err := h.deps.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := compiled.New(tx)

	responseRow, err := q.RatingInsertResponse(ctx, compiled.RatingInsertResponseParams{
		CycleID:       eventRow.ActiveCycleID.Int64,
		ParticipantID: participant.ID,
	})
	if err != nil {
		return nil, err
	}

	for _, item := range input.Body.Items {
		var valNum pgtype.Int4
		if item.ValueNumber != nil {
			valNum = pgtype.Int4{Int32: *item.ValueNumber, Valid: true}
		}
		var valText pgtype.Text
		if item.ValueText != nil {
			valText = pgtype.Text{String: *item.ValueText, Valid: true}
		}
		if err := q.RatingInsertResponseItem(ctx, compiled.RatingInsertResponseItemParams{
			ResponseID:  responseRow.ID,
			FormID:      item.FormID,
			ValueNumber: valNum,
			ValueText:   valText,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return nil, nil
}

func (h *Handler) handleGetMonitor(ctx context.Context, input *GetMonitorInput) (*GetMonitorOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEventByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.CompanyID != p.SelectedCompanyID {
		return nil, humax.NotFound("event not found")
	}
	if eventRow.HostID != p.UserID {
		return nil, humax.Forbidden("only the event host can view monitor")
	}

	row, err := h.deps.Queries.RatingGetMonitorCounts(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	resp := MonitorResponse{
		ParticipantCount: row.ParticipantCount,
		RespondedCount:   row.RespondedCount,
	}
	if row.ActiveCycleID.Valid {
		cycleID := row.ActiveCycleID.Int64
		resp.ActiveCycleID = &cycleID
	}

	return &GetMonitorOutput{Body: resp}, nil
}

func (h *Handler) getResultsData(ctx context.Context, eventID int64) (*ResultsResponse, error) {
	cycleRows, err := h.deps.Queries.RatingGetEventCycles(ctx, eventID)
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

	formRows, err := h.deps.Queries.RatingGetFormsForEvent(ctx, eventID)
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

	avgRows, err := h.deps.Queries.RatingGetResultAverages(ctx, eventID)
	if err != nil {
		return nil, err
	}
	avgTable := make([]CycleAverageResult, len(avgRows))
	for i, r := range avgRows {
		avgTable[i] = CycleAverageResult{
			CycleID: r.CycleID,
			FormID:  r.FormID,
			Average: r.Average,
		}
	}

	ftRows, err := h.deps.Queries.RatingGetResultFreeTexts(ctx, eventID)
	if err != nil {
		return nil, err
	}
	type ftKey struct{ cycleID, formID int64 }
	ftMap := make(map[ftKey][]string)
	ftOrder := []ftKey{}
	for _, r := range ftRows {
		key := ftKey{r.CycleID, r.FormID}
		if _, exists := ftMap[key]; !exists {
			ftOrder = append(ftOrder, key)
		}
		if r.ValueText.Valid {
			ftMap[key] = append(ftMap[key], r.ValueText.String)
		}
	}
	freeTexts := make([]FreeTextResult, 0, len(ftOrder))
	for _, key := range ftOrder {
		freeTexts = append(freeTexts, FreeTextResult{
			CycleID: key.cycleID,
			FormID:  key.formID,
			Texts:   ftMap[key],
		})
	}

	return &ResultsResponse{
		Cycles:    cycles,
		Forms:     forms,
		AvgTable:  avgTable,
		FreeTexts: freeTexts,
	}, nil
}

func (h *Handler) handleGetResults(ctx context.Context, input *GetResultsInput) (*GetResultsOutput, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEventByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.CompanyID != p.SelectedCompanyID {
		return nil, humax.NotFound("event not found")
	}
	if eventRow.HostID != p.UserID {
		return nil, humax.Forbidden("only the event host can view results")
	}
	if eventRow.Status != "ended" {
		return nil, humax.Unprocessable("event has not ended")
	}

	data, err := h.getResultsData(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetResultsOutput{Body: *data}, nil
}

func (h *Handler) handleExportCSV(ctx context.Context, input *ExportCSVInput) (*huma.StreamResponse, error) {
	p, err := humax.RequireSelectedCompany(ctx)
	if err != nil {
		return nil, err
	}

	eventRow, err := h.deps.Queries.RatingGetEventByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, humax.NotFound("event not found")
		}
		return nil, err
	}
	if eventRow.CompanyID != p.SelectedCompanyID {
		return nil, humax.NotFound("event not found")
	}
	if eventRow.HostID != p.UserID {
		return nil, humax.Forbidden("only the event host can export results")
	}
	if eventRow.Status != "ended" {
		return nil, humax.Unprocessable("event has not ended")
	}

	data, err := h.getResultsData(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &huma.StreamResponse{
		Body: func(hctx huma.Context) {
			hctx.SetHeader("Content-Type", "text/csv")
			hctx.SetHeader("Content-Disposition", `attachment; filename="results.csv"`)
			w := hctx.BodyWriter()
			cw := csv.NewWriter(w)

			header := make([]string, 0, 1+len(data.Forms))
			header = append(header, "Cycle")
			for _, f := range data.Forms {
				header = append(header, f.Label)
			}
			cw.Write(header) //nolint:errcheck

			avgIndex := make(map[int64]map[int64]float64)
			for _, a := range data.AvgTable {
				if avgIndex[a.CycleID] == nil {
					avgIndex[a.CycleID] = make(map[int64]float64)
				}
				avgIndex[a.CycleID][a.FormID] = a.Average
			}

			ftIndex := make(map[int64]map[int64][]string)
			for _, ft := range data.FreeTexts {
				if ftIndex[ft.CycleID] == nil {
					ftIndex[ft.CycleID] = make(map[int64][]string)
				}
				ftIndex[ft.CycleID][ft.FormID] = ft.Texts
			}

			for _, cycle := range data.Cycles {
				row := make([]string, 0, 1+len(data.Forms))
				row = append(row, cycle.Name)
				for _, f := range data.Forms {
					switch f.Type {
					case "rating", "mood":
						if byForm, ok := avgIndex[cycle.ID]; ok {
							if avg, ok2 := byForm[f.ID]; ok2 {
								row = append(row, fmt.Sprintf("%.2f", avg))
								continue
							}
						}
						row = append(row, "")
					case "free_text":
						if byForm, ok := ftIndex[cycle.ID]; ok {
							if texts, ok2 := byForm[f.ID]; ok2 {
								row = append(row, strings.Join(texts, " | "))
								continue
							}
						}
						row = append(row, "")
					default:
						row = append(row, "")
					}
				}
				cw.Write(row) //nolint:errcheck
			}

			cw.Flush()
		},
	}, nil
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

