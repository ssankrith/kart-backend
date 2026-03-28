-- Option A: flat product row with optional image JSON (Supabase / Postgres).

CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    price NUMERIC(12, 2) NOT NULL,
    category TEXT NOT NULL,
    image_json JSONB
);
