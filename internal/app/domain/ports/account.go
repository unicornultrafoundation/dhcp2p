package ports

type AccountService interface {
	GetAddress() (string, error)
	Sign(payload []byte) ([]byte, error)
}
