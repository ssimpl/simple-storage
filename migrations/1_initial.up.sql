CREATE TABLE IF NOT EXISTS objects_metadata (
    name TEXT PRIMARY KEY,
    fragments JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS servers (
    id UUID NOT NULL DEFAULT gen_random_uuid (),
    addr TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO
    servers (addr)
VALUES ('storage1:51051'),
    ('storage2:52051'),
    ('storage3:53051'),
    ('storage4:54051'),
    ('storage5:55051'),
    ('storage6:56051');
