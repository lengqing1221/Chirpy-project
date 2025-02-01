INSERT INTO users (id, email, created_at, updated_at)
VALUES (gen_random_uuid(), NOW(), NOW(), $1
)
RETURNING *;
