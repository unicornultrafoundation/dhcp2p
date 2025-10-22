package models

type AuthRequest struct {
	Pubkey []byte
}

type AuthResponse struct {
	NonceID string
}

type AuthVerifyRequest struct {
	NonceID   string
	Signature []byte
	Pubkey    []byte
}

type AuthVerifyResponse struct {
	Pubkey []byte
}
