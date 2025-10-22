-- name: GetNonce :one
SELECT id, peer_id, issued_at, expires_at, used, used_at FROM nonces 
WHERE id = $1 AND expires_at > now() AND used = false;

-- name: CreateNonce :one
INSERT INTO nonces (peer_id, issued_at, expires_at) 
VALUES ($1, now(), now() + (sqlc.arg(ttl)::int * interval '1 minute')) 
RETURNING id, peer_id, issued_at, expires_at, used, used_at;

-- name: ConsumeNonce :one
UPDATE nonces
SET used = true, used_at = now()
WHERE id = $1 AND peer_id = $2 AND used = false AND expires_at > now()
RETURNING id, peer_id, issued_at, expires_at, used, used_at;

-- name: DeleteExpiredNonces :exec
DELETE FROM nonces WHERE expires_at < now();

-- name: GetLeaseByTokenID :one
SELECT token_id, peer_id, expires_at, created_at, updated_at, EXTRACT(EPOCH FROM (expires_at - now()))::int AS ttl
FROM leases
WHERE token_id = $1 AND expires_at > now();

-- name: GetLeaseByPeerID :one
SELECT token_id, peer_id, expires_at, created_at, updated_at, EXTRACT(EPOCH FROM (expires_at - now()))::int AS ttl
FROM leases
WHERE peer_id = $1 AND expires_at > now();

-- name: FindExpiredLeaseForReuse :one
SELECT token_id, peer_id, expires_at, created_at, updated_at, EXTRACT(EPOCH FROM (expires_at - now()))::int AS ttl
FROM leases
WHERE expires_at < now()
ORDER BY expires_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: ReuseLease :one
UPDATE leases
SET peer_id = $1,
    expires_at = now() + (sqlc.arg(ttl)::int * interval '1 minute'),
    updated_at = now()
WHERE token_id = $2
RETURNING token_id, peer_id, expires_at, created_at, updated_at, EXTRACT(EPOCH FROM (expires_at - now()))::int AS ttl;

-- name: RenewLease :one
UPDATE leases
SET expires_at = now() + (sqlc.arg(ttl)::int * interval '1 minute'),
    updated_at = now()
WHERE token_id = $1 AND peer_id = $2 AND expires_at > now()
RETURNING token_id, peer_id, expires_at, created_at, updated_at, EXTRACT(EPOCH FROM (expires_at - now()))::int AS ttl;

-- name: InsertLease :one
INSERT INTO leases (token_id, peer_id, expires_at, created_at, updated_at)
VALUES ($1, $2, now() + (sqlc.arg(ttl)::int * interval '1 minute'), now(), now())
RETURNING token_id, peer_id, expires_at, created_at, updated_at, EXTRACT(EPOCH FROM (expires_at - now()))::int AS ttl;

-- name: AllocateNextTokenID :one
UPDATE alloc_state
SET last_token_id = (last_token_id - 1)
WHERE id = 1
RETURNING last_token_id;

-- name: ReleaseLease :exec
UPDATE leases
SET expires_at = now()
WHERE token_id = $1 AND peer_id = $2;