-- ============================================================
-- events
-- ============================================================
CREATE TABLE IF NOT EXISTS events (
    id               BIGSERIAL PRIMARY KEY,
    company_id       BIGINT NOT NULL REFERENCES companies(id),
    host_id          BIGINT NOT NULL REFERENCES users(id),
    name             TEXT NOT NULL,
    description      TEXT,
    visibility       TEXT NOT NULL CHECK (visibility IN ('public', 'private')),
    status           TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'ended')),
    current_stage    TEXT NOT NULL DEFAULT 'idle' CHECK (current_stage IN ('idle', 'waiting', 'form_open')),
    active_cycle_id  BIGINT,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_company ON events (company_id);
CREATE INDEX IF NOT EXISTS idx_events_host    ON events (host_id);

-- ============================================================
-- cycles
-- ============================================================
CREATE TABLE IF NOT EXISTS cycles (
    id           BIGSERIAL PRIMARY KEY,
    event_id     BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    order_index  INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cycles_event ON cycles (event_id);

-- ============================================================
-- forms
-- ============================================================
CREATE TABLE IF NOT EXISTS forms (
    id           BIGSERIAL PRIMARY KEY,
    event_id     BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    type         TEXT NOT NULL CHECK (type IN ('rating', 'mood', 'free_text')),
    label        TEXT NOT NULL,
    order_index  INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_forms_event ON forms (event_id);

-- ============================================================
-- event_members
-- ============================================================
CREATE TABLE IF NOT EXISTS event_members (
    event_id  BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id   BIGINT NOT NULL REFERENCES users(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_event_members ON event_members (event_id, user_id);

-- ============================================================
-- participants
-- ============================================================
CREATE TABLE IF NOT EXISTS participants (
    id        BIGSERIAL PRIMARY KEY,
    event_id  BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id   BIGINT NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_participants_event_user ON participants (event_id, user_id);
CREATE INDEX IF NOT EXISTS idx_participants_event ON participants (event_id);

-- ============================================================
-- responses
-- ============================================================
CREATE TABLE IF NOT EXISTS responses (
    id              BIGSERIAL PRIMARY KEY,
    cycle_id        BIGINT NOT NULL REFERENCES cycles(id) ON DELETE CASCADE,
    participant_id  BIGINT NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    submitted_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_responses_cycle_participant ON responses (cycle_id, participant_id);
CREATE INDEX IF NOT EXISTS idx_responses_cycle ON responses (cycle_id);

-- ============================================================
-- response_items
-- ============================================================
CREATE TABLE IF NOT EXISTS response_items (
    id           BIGSERIAL PRIMARY KEY,
    response_id  BIGINT NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
    form_id      BIGINT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    value_number INT,
    value_text   TEXT
);

CREATE INDEX IF NOT EXISTS idx_response_items_response ON response_items (response_id);

-- ============================================================
-- events.active_cycle_id FK (after cycles exists)
-- ============================================================
ALTER TABLE events
    ADD CONSTRAINT fk_events_active_cycle
    FOREIGN KEY (active_cycle_id) REFERENCES cycles(id);
