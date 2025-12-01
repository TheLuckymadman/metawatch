CREATE TABLE metrics (
    agent_id text NOT NULL,
    id text NOT NULL,
    mtype text NOT NULL CHECK (mtype IN ('counter','gauge')),
    delta bigint,
    value double precision,
    hash text,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT metrics_valid_pair_chk CHECK (
        (mtype = 'counter' AND delta IS NOT NULL AND value IS NULL) OR
        (mtype = 'gauge' AND value IS NOT NULL AND delta IS NULL)
    ),
    CONSTRAINT metrics_pk PRIMARY KEY (agent_id, id, mtype)
);

CREATE INDEX idx_metrics_mtype ON metrics(mtype);
CREATE INDEX idx_metrics_agent ON metrics(agent_id);