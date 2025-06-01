package service

import (
	"context"
	"log/slog"

	"harajuku/backend/internal/adapter/storage/postgres"
	"harajuku/backend/internal/adapter/storage/postgres/repository"
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"harajuku/backend/internal/core/util"

	"github.com/google/uuid"
)

type PaymentProofService struct {
	repo      port.PaymentProofRepository
	file      port.FileRepository
	quoteRepo port.QuoteRepository
	email     port.EmailRepository
	db        postgres.DB
	cache     port.CacheRepository
}

func NewPaymentProofService(
	repo port.PaymentProofRepository,
	file port.FileRepository,
	quoteRepo port.QuoteRepository,
	email port.EmailRepository,
	db postgres.DB,
	cache port.CacheRepository,
) *PaymentProofService {
	return &PaymentProofService{
		repo:      repo,
		file:      file,
		quoteRepo: quoteRepo,
		email:     email,
		db:        db,
		cache:     cache,
	}
}

// CreatePaymentProof carga la imagen y crea el registro asociado a una cotización
func (ps *PaymentProofService) CreatePaymentProof(ctx context.Context, proof *domain.PaymentProof, file []byte, fileName string) (*domain.PaymentProof, error) {
	// Validar que la cotización exista
	_, err := ps.quoteRepo.GetQuoteByID(ctx, proof.QuoteID)
	if err != nil {
		return nil, domain.ErrDataNotFound
	}

	// Verificar que no exista ya un comprobante para la cotización
	existing, err := ps.repo.GetPaymentProofByQuoteID(ctx, proof.QuoteID)
	if err != nil && err != domain.ErrDataNotFound {
		slog.Error("failed to get existing payment proof", "error", err)
		return nil, domain.ErrInternal
	}
	if existing != nil {
		slog.Warn("payment proof already exists for quote", "quoteID", proof.QuoteID)
		return nil, domain.ErrConflictingData
	}

	// Upload the image first. Fail fast if this errors.
	path, err := ps.file.Save(ctx, file, fileName)
	if err != nil {
		slog.Error("file save failed", "error", err)
		return nil, domain.ErrInternal
	}

	// Atomic DB transaction: create quote + image row
	var created *domain.PaymentProof

	err = ps.db.WithTx(ctx, func(txDB *postgres.DB) error {
		// build repo on txDB
		txRepo := repository.NewPaymentProofRepository(txDB)

		// insert image in *same* tx
		proof.ID = uuid.New()
		proof.URL = path
		proof.IsReviewed = false // default

		p, err := txRepo.CreatePaymentProof(ctx, proof)
		if err != nil {
			return err
		}
		created = p
		return nil
	})

	if err != nil {
		slog.Error("transaction failed", "error", err)
		_ = ps.file.Delete(ctx, path) // limpiar archivo
		return nil, domain.ErrInternal
	}

	// Cachear
	cacheKey := util.GenerateCacheKey("paymentProof", created.ID)
	data, _ := util.Serialize(created)
	if err := ps.cache.Set(ctx, cacheKey, data, 0); err != nil {
		slog.Warn("cache set failed", "error", err)
	}
	_ = ps.cache.DeleteByPrefix(ctx, "paymentProofs:*")

	return created, nil
}

// GetPaymentProofByID obtiene comprobante por ID con cache y archivo desde S3
func (ps *PaymentProofService) GetPaymentProofByID(ctx context.Context, id uuid.UUID) (*domain.PaymentProof, []byte, error) {
	var proof *domain.PaymentProof
	cacheKey := util.GenerateCacheKey("paymentProof", id)

	cached, err := ps.cache.Get(ctx, cacheKey)
	if err == nil {
		err = util.Deserialize(cached, &proof)
		if err == nil {
			// obtener archivo desde S3
			file, err := ps.file.Get(ctx, proof.URL)
			if err != nil {
				return nil, nil, domain.ErrInternal
			}
			return proof, file, nil
		}
	}

	proof, err = ps.repo.GetPaymentProofByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, nil, err
		}
		return nil, nil, domain.ErrInternal
	}

	data, err := util.Serialize(proof)
	if err == nil {
		_ = ps.cache.Set(ctx, cacheKey, data, 0)
	}

	// obtener archivo desde S3
	file, err := ps.file.Get(ctx, proof.URL)
	if err != nil {
		return nil, nil, domain.ErrInternal
	}

	return proof, file, nil
}

// GetPaymentProofs lista comprobantes con filtro y cache
func (ps *PaymentProofService) GetPaymentProofs(ctx context.Context, filter port.PaymentProofFilter) ([]domain.PaymentProof, error) {
	var proofs []domain.PaymentProof

	params := util.GenerateCacheKeyParams(filter)
	cacheKey := util.GenerateCacheKey("paymentProofs", params)

	cached, err := ps.cache.Get(ctx, cacheKey)
	if err == nil {
		err = util.Deserialize(cached, &proofs)
		if err == nil {
			return proofs, nil
		}
	}

	proofs, err = ps.repo.GetPaymentProofs(ctx, filter)
	if err != nil {
		return nil, domain.ErrInternal
	}

	data, err := util.Serialize(proofs)
	if err == nil {
		_ = ps.cache.Set(ctx, cacheKey, data, 0)
	}

	return proofs, nil
}

// UpdatePaymentProof permite actualizar el campo IsReviewed (por ejemplo)
func (ps *PaymentProofService) UpdatePaymentProof(ctx context.Context, proof *domain.PaymentProof) (*domain.PaymentProof, error) {
	existing, err := ps.repo.GetPaymentProofByID(ctx, proof.ID)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	if existing.IsReviewed == proof.IsReviewed {
		return nil, domain.ErrNoUpdatedData
	}

	updated, err := ps.repo.UpdatePaymentProof(ctx, proof)
	if err != nil {
		if err == domain.ErrConflictingData {
			return nil, err
		}
		return nil, domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("paymentProof", updated.ID)
	_ = ps.cache.Delete(ctx, cacheKey)

	data, err := util.Serialize(updated)
	if err == nil {
		_ = ps.cache.Set(ctx, cacheKey, data, 0)
	}

	_ = ps.cache.DeleteByPrefix(ctx, "paymentProofs:*")

	return updated, nil
}

// DeletePaymentProof elimina comprobante y archivo asociado
func (ps *PaymentProofService) DeletePaymentProof(ctx context.Context, id uuid.UUID) error {
	proof, err := ps.repo.GetPaymentProofByID(ctx, id)
	if err != nil {
		if err == domain.ErrDataNotFound {
			return err
		}
		return domain.ErrInternal
	}

	err = ps.db.WithTx(ctx, func(txDB *postgres.DB) error {
		txRepo := repository.NewPaymentProofRepository(txDB)

		if err := ps.file.Delete(ctx, proof.URL); err != nil {
			return err
		}

		return txRepo.DeletePaymentProof(ctx, id)
	})

	if err != nil {
		slog.Error("transaction failed", "error", err)
		return domain.ErrInternal
	}

	cacheKey := util.GenerateCacheKey("paymentProof", id)
	_ = ps.cache.Delete(ctx, cacheKey)
	_ = ps.cache.DeleteByPrefix(ctx, "paymentProofs:*")

	return nil
}
