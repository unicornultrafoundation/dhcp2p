package mocks

//go:generate mockgen -source=../../internal/app/domain/ports/lease.go -destination=lease_repository_mock.go -package=mocks
//go:generate mockgen -source=../../internal/app/domain/ports/nonce.go -destination=nonce_repository_mock.go -package=mocks  
//go:generate mockgen -source=../../internal/app/domain/ports/auth.go -destination=auth_repository_mock.go -package=mocks
//go:generate mockgen -source=../../internal/app/domain/ports/verifier.go -destination=verifier_mock.go -package=mocks

//go:generate echo "Mock generation completed. Run 'go generate' from tests/mocks directory."
