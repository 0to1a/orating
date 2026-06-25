package rating

import "time"

type CycleInput struct {
	Name string `json:"name"`
}

type FormInput struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type CreateEventRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Visibility  string       `json:"visibility"`
	Cycles      []CycleInput `json:"cycles"`
	Forms       []FormInput  `json:"forms"`
	Members     []string     `json:"members,omitempty"`
}

type EventInfo struct {
	ID           int64     `json:"id"`
	CompanyID    int64     `json:"companyId"`
	HostID       int64     `json:"hostId"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Visibility   string    `json:"visibility"`
	Status       string    `json:"status"`
	CurrentStage string    `json:"currentStage"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CycleInfo struct {
	ID         int64  `json:"id"`
	EventID    int64  `json:"eventId"`
	Name       string `json:"name"`
	OrderIndex int32  `json:"orderIndex"`
}

type FormInfo struct {
	ID         int64  `json:"id"`
	EventID    int64  `json:"eventId"`
	Type       string `json:"type"`
	Label      string `json:"label"`
	OrderIndex int32  `json:"orderIndex"`
}

type EventDetail struct {
	EventInfo
	Cycles []CycleInfo `json:"cycles"`
	Forms  []FormInfo  `json:"forms"`
}

type AddMemberRequest struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Huma input/output wrappers.

type CreateEventInput struct {
	Body CreateEventRequest
}

type CreateEventOutput struct {
	Body EventDetail
}

type ListEventsOutput struct {
	Body struct {
		Events []EventInfo `json:"events"`
	}
}

type GetEventInput struct {
	ID int64 `path:"id"`
}

type GetEventOutput struct {
	Body EventDetail
}

type AddMemberInput struct {
	EventID int64 `path:"id"`
	Body    AddMemberRequest
}

type AddMemberOutput struct{}

type RemoveMemberInput struct {
	EventID int64 `path:"id"`
	UserID  int64 `path:"userId"`
}

type StartCycleRequest struct {
	CycleID int64 `json:"cycleId"`
}

type NextCycleRequest struct {
	CycleID int64 `json:"cycleId"`
}

type EventIDInput struct {
	ID int64 `path:"id"`
}

type StartCycleInput struct {
	ID   int64 `path:"id"`
	Body StartCycleRequest
}

type NextCycleInput struct {
	ID   int64 `path:"id"`
	Body NextCycleRequest
}

type SessionResponse struct {
	CurrentStage        string     `json:"currentStage"`
	ActiveCycleID       *int64     `json:"activeCycleId,omitempty"`
	ActiveCycleName     string     `json:"activeCycleName,omitempty"`
	MyResponseSubmitted bool       `json:"myResponseSubmitted"`
	Forms               []FormInfo `json:"forms"`
}

type RespondItemInput struct {
	FormID      int64   `json:"formId"`
	ValueNumber *int32  `json:"valueNumber,omitempty"`
	ValueText   *string `json:"valueText,omitempty"`
}

type RespondRequest struct {
	Items []RespondItemInput `json:"items"`
}

type ParticipantInfo struct {
	ID       int64     `json:"id"`
	EventID  int64     `json:"eventId"`
	UserID   int64     `json:"userId"`
	JoinedAt time.Time `json:"joinedAt"`
}

type JoinEventInput struct {
	ID int64 `path:"id"`
}

type JoinEventOutput struct {
	Body ParticipantInfo
}

type GetSessionInput struct {
	ID int64 `path:"id"`
}

type GetSessionOutput struct {
	Body SessionResponse
}

type RespondInput struct {
	ID   int64 `path:"id"`
	Body RespondRequest
}

type MonitorResponse struct {
	ParticipantCount int64  `json:"participantCount"`
	RespondedCount   int64  `json:"respondedCount"`
	ActiveCycleID    *int64 `json:"activeCycleId,omitempty"`
}

type GetMonitorInput struct {
	ID int64 `path:"id"`
}

type GetMonitorOutput struct {
	Body MonitorResponse
}

type CycleAverageResult struct {
	CycleID int64   `json:"cycleId"`
	FormID  int64   `json:"formId"`
	Average float64 `json:"average"`
}

type FreeTextResult struct {
	CycleID int64    `json:"cycleId"`
	FormID  int64    `json:"formId"`
	Texts   []string `json:"texts"`
}

type ResultsResponse struct {
	Cycles    []CycleInfo          `json:"cycles"`
	Forms     []FormInfo           `json:"forms"`
	AvgTable  []CycleAverageResult `json:"avgTable"`
	FreeTexts []FreeTextResult     `json:"freeTexts"`
}

type GetResultsInput struct {
	ID int64 `path:"id"`
}

type GetResultsOutput struct {
	Body ResultsResponse
}

type ExportCSVInput struct {
	ID int64 `path:"id"`
}
