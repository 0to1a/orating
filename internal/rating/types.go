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
