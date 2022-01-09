package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	pg "github.com/lib/pq"
	"grader/internal/app/panel/model"
	"grader/internal/app/panel/storage"
	"grader/pkg/apperr"
	"grader/pkg/logger"
)

// storage.AssessmentRepository interface implementation
var _ storage.AssessmentRepository = (*AssessmentRepository)(nil)

type AssessmentRepository struct {
	db *sql.DB
}

func NewAssessmentRepository(db *sql.DB) (*AssessmentRepository, error) {
	s := &AssessmentRepository{
		db: db,
	}

	return s, nil
}

// Create implementation of interface storage.AssessmentRepository
func (r *AssessmentRepository) Create(ctx context.Context, m *model.Assessment) (*model.Assessment, error) {
	const SQL = `
		INSERT INTO assessments (part_id, container_image, summary, file_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id
`

	err := r.db.QueryRowContext(
		ctx,
		SQL,
		m.PartID,
		m.ContainerImage,
		m.Summary,
		m.FileName,
	).Scan(&m.ID)
	if err != nil {
		if pgErr, ok := err.(*pg.Error); ok {
			if pgerrcode.IsIntegrityConstraintViolation(string(pgErr.Code)) {
				return nil, apperr.ErrConflict
			}
		}

		return nil, fmt.Errorf("insert: %w", err)
	}

	return m, nil
}

// Read implementation of interface storage.AssessmentRepository
func (r *AssessmentRepository) Read(ctx context.Context, id uuid.UUID) (*model.Assessment, error) {
	const SQL = `
		SELECT id, created_at, part_id, container_image, summary, file_name
		FROM assessments 
		WHERE id=$1
`
	m := &model.Assessment{}

	err := r.db.QueryRowContext(ctx, SQL, id).Scan(
		&m.ID,
		&m.CreatedAt,
		&m.PartID,
		&m.ContainerImage,
		&m.Summary,
		&m.FileName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("select: %w", err)
	}

	return m, nil
}

func (r *AssessmentRepository) All(ctx context.Context) ([]*model.Assessment, error) {
	l := logger.Ctx(ctx).With().Str("method", "All").Logger()

	const SQL = `
		SELECT id, created_at, part_id, container_image, summary, file_name
		FROM assessments
		ORDER BY created_at
`
	rows, err := r.db.QueryContext(ctx, SQL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Debug().Err(err).Send()
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("select: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	res := make([]*model.Assessment, 0)

	for rows.Next() {
		if err := rows.Err(); err != nil {
			l.Debug().Err(err).Send()
			return nil, fmt.Errorf("rows next: %w", err)
		}
		m := &model.Assessment{}
		if err := rows.Scan(
			&m.ID,
			&m.CreatedAt,
			&m.PartID,
			&m.ContainerImage,
			&m.Summary,
			&m.FileName,
		); err != nil {
			l.Debug().Err(err).Send()
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, m)
	}

	l.Debug().Msgf("Found: %#v", res)

	return res, nil
}
