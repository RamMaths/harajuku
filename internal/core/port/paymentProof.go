package port

import (
	"context"
	"harajuku/backend/internal/core/domain"

	"github.com/google/uuid"
)

//go:generate mockgen -source=payment_proof.go -destination=mock/payment_proof.go -package=mock

type PaymentProofFilter struct {
	QuoteID    *uuid.UUID // Optional filter by QuoteID
	IsReviewed *bool      // Optional filter by IsReviewed status
	Skip       uint64     // Number of records to skip for pagination
	Limit      uint64     // Maximum number of records to return

}

// PaymentProofRepository is an interface for interacting with payment-proof-related data
type PaymentProofRepository interface {
	// CreatePaymentProof inserts a new payment proof into the database
	CreatePaymentProof(ctx context.Context, proof *domain.PaymentProof) (*domain.PaymentProof, error)
	// UpdatePaymentProof updates an existing payment proof
	UpdatePaymentProof(ctx context.Context, proof *domain.PaymentProof) (*domain.PaymentProof, error)
	// DeletePaymentProof deletes a payment proof by ID
	DeletePaymentProof(ctx context.Context, id uuid.UUID) error
	// GetPaymentProofByID selects a payment proof by its ID
	GetPaymentProofByID(ctx context.Context, id uuid.UUID) (*domain.PaymentProof, error)
	GetPaymentProofByQuoteID(ctx context.Context, quoteID uuid.UUID) (*domain.PaymentProof, error)
	// GetPaymentProofs selects all payment proofs with optional filtering by QuoteID
	GetPaymentProofs(ctx context.Context, filter PaymentProofFilter) ([]domain.PaymentProof, error)
	// Wrap a function in a DB transaction; if fn returns an error, rollback
	WithTx(ctx context.Context, fn func(repo PaymentProofRepository) error) error
}

// PaymentProofService is an interface for interacting with payment-proof-related business logic
type PaymentProofService interface {
	// CreatePaymentProof creates a new payment proof
	CreatePaymentProof(ctx context.Context, proof *domain.PaymentProof, file []byte, fileName string) (*domain.PaymentProof, error)
	// UpdatePaymentProof updates an existing payment proof
	UpdatePaymentProof(ctx context.Context, proof *domain.PaymentProof) (*domain.PaymentProof, error)
	// DeletePaymentProof deletes a payment proof by its ID
	DeletePaymentProof(ctx context.Context, id uuid.UUID) error
	// GetPaymentProofByID returns a payment proof and its file content by its ID
	GetPaymentProofByID(ctx context.Context, id uuid.UUID) (*domain.PaymentProof, []byte, error)
	// GetPaymentProofs returns a list of payment proofs, with optional filtering by quoteId
	GetPaymentProofs(ctx context.Context, filter PaymentProofFilter) ([]domain.PaymentProof, error)
}
