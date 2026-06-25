-- name: RatingCreateEvent :one
INSERT INTO events (company_id, host_id, name, description, visibility, status, current_stage)
VALUES ($1, $2, $3, $4, $5, 'draft', 'idle')
RETURNING *;

-- name: RatingBulkInsertCycles :many
INSERT INTO cycles (event_id, name, order_index)
SELECT $1, unnest($2::text[]), generate_series(0, array_length($2::text[], 1) - 1)
RETURNING *;

-- name: RatingBulkInsertForms :many
INSERT INTO forms (event_id, type, label, order_index)
SELECT $1, unnest($2::text[]), unnest($3::text[]), generate_series(0, array_length($2::text[], 1) - 1)
RETURNING *;

-- name: RatingListEvents :many
SELECT * FROM events
WHERE company_id = $1
  AND (visibility = 'public' OR host_id = $2 OR EXISTS (
    SELECT 1 FROM event_members WHERE event_id = events.id AND user_id = $2
  ))
ORDER BY created_at DESC;

-- name: RatingGetEvent :one
SELECT * FROM events
WHERE id = $1 AND company_id = $2
  AND (visibility = 'public' OR host_id = $3 OR EXISTS (
    SELECT 1 FROM event_members WHERE event_id = events.id AND user_id = $3
  ));

-- name: RatingGetEventCycles :many
SELECT * FROM cycles WHERE event_id = $1 ORDER BY order_index;

-- name: RatingGetEventForms :many
SELECT * FROM forms WHERE event_id = $1 ORDER BY order_index;

-- name: RatingAddEventMember :exec
INSERT INTO event_members (event_id, user_id) VALUES ($1, $2);

-- name: RatingRemoveEventMember :exec
DELETE FROM event_members WHERE event_id = $1 AND user_id = $2;

-- name: RatingGetEventMember :one
SELECT event_id, user_id FROM event_members WHERE event_id = $1 AND user_id = $2;

-- name: RatingFindUserByEmail :one
SELECT id, email, name FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: RatingCreateUser :one
INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id, email, name;

-- name: RatingActivateEvent :one
UPDATE events
SET status = 'active', current_stage = 'idle', updated_at = NOW()
WHERE id = $1 AND company_id = $2 AND host_id = $3 AND status = 'draft'
RETURNING id;

-- name: RatingStartCycle :one
UPDATE events
SET active_cycle_id = $2, current_stage = 'waiting', updated_at = NOW()
WHERE id = $1 AND company_id = $3 AND host_id = $4 AND status = 'active'
RETURNING id;

-- name: RatingShowForm :one
UPDATE events
SET current_stage = 'form_open', updated_at = NOW()
WHERE id = $1 AND company_id = $2 AND host_id = $3 AND status = 'active' AND current_stage = 'waiting'
RETURNING id;

-- name: RatingNextCycle :one
UPDATE events
SET active_cycle_id = $2, current_stage = 'waiting', updated_at = NOW()
WHERE id = $1 AND company_id = $3 AND host_id = $4 AND status = 'active'
RETURNING id;

-- name: RatingEndEvent :one
UPDATE events
SET status = 'ended', updated_at = NOW()
WHERE id = $1 AND company_id = $2 AND host_id = $3 AND status = 'active'
RETURNING id;

-- name: RatingGetCycleByEvent :one
SELECT * FROM cycles WHERE id = $1 AND event_id = $2;

-- name: RatingGetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: RatingJoinEvent :one
INSERT INTO participants (event_id, user_id)
VALUES ($1, $2)
ON CONFLICT (event_id, user_id) DO NOTHING
RETURNING *;

-- name: RatingGetParticipant :one
SELECT * FROM participants WHERE event_id = $1 AND user_id = $2;

-- name: RatingGetEventForSession :one
SELECT id, status, current_stage, active_cycle_id FROM events WHERE id = $1 AND company_id = $2;

-- name: RatingGetCycleByID :one
SELECT id, name, order_index FROM cycles WHERE id = $1;

-- name: RatingGetFormsForEvent :many
SELECT id, event_id, type, label, order_index FROM forms WHERE event_id = $1 ORDER BY order_index;

-- name: RatingGetResponseForParticipantCycle :one
SELECT id FROM responses WHERE cycle_id = $1 AND participant_id = $2;

-- name: RatingInsertResponse :one
INSERT INTO responses (cycle_id, participant_id)
VALUES ($1, $2)
RETURNING id, submitted_at;

-- name: RatingInsertResponseItem :exec
INSERT INTO response_items (response_id, form_id, value_number, value_text)
VALUES ($1, $2, $3, $4);

-- name: RatingGetEventMemberCheck :one
SELECT event_id FROM event_members WHERE event_id = $1 AND user_id = $2;

-- name: RatingGetFormByID :one
SELECT id, event_id, type FROM forms WHERE id = $1;
