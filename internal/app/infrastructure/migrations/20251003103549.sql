-- Ensure alloc_state has the initial row
INSERT INTO alloc_state (id, last_token_id)
VALUES (1, 167902209)
ON CONFLICT (id) DO NOTHING;