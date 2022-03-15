package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	pg "github.com/lib/pq"
	"grader/internal/app/panel/storage"
	"grader/internal/pkg/model"
	"grader/pkg/apperr"
	"grader/pkg/logger"
)

// storage.SubmissionRepository interface implementation
var _ storage.SubmissionRepository = (*SubmissionRepository)(nil)

type SubmissionRepository struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) (*SubmissionRepository, error) {
	s := &SubmissionRepository{
		db: db,
	}

	return s, nil
}

// Create implementation of interface storage.SubmissionRepository
func (r *SubmissionRepository) Create(ctx context.Context, m *model.Submission) (*model.Submission, error) {
	const SQL = `
		INSERT INTO Submissions (
			user_id,
			assessment_id,
			file_name,
			file_url
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id
`
	err := r.db.QueryRowContext(
		ctx,
		SQL,
		m.UserID,
		m.AssessmentID,
		m.FileName,
		m.FileURL,
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

// Read implementation of interface storage.SubmissionRepository
func (r *SubmissionRepository) Read(ctx context.Context, id uuid.UUID) (*model.Submission, error) {
	const SQL = `
		SELECT
			id,
			created_at,
			user_id,
			assessment_id,
			file_name,
			file_url,
			external_id,
			result_date,
			result_pass,
			result_text
		FROM Submissions 
		WHERE id=$1
`
	m := &model.Submission{}

	err := r.db.QueryRowContext(ctx, SQL, id).Scan(
		&m.ID,
		&m.CreatedAt,
		m.UserID,
		m.AssessmentID,
		m.FileName,
		m.FileURL,
		m.ExternalID,
		m.ResultDate,
		m.ResultPass,
		m.ResultText,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("select: %w", err)
	}

	return m, nil
}

func (r *SubmissionRepository) All(ctx context.Context) ([]*model.Submission, error) {
	l := logger.Ctx(ctx).With().Str("method", "All").Logger()

	const SQL = `
		SELECT
			id,
			created_at,
			user_id,
			assessment_id,
			file_name,
			file_url,
			external_id,
			result_date,
			result_pass,
			result_text
		FROM Submissions
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

	res := make([]*model.Submission, 0)

	for rows.Next() {
		if err := rows.Err(); err != nil {
			l.Debug().Err(err).Send()
			return nil, fmt.Errorf("rows next: %w", err)
		}
		m := &model.Submission{}
		if err := rows.Scan(
			&m.ID,
			&m.CreatedAt,
			m.UserID,
			m.AssessmentID,
			m.FileName,
			m.FileURL,
			m.ExternalID,
			m.ResultDate,
			m.ResultPass,
			m.ResultText,
		); err != nil {
			l.Debug().Err(err).Send()
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, m)
	}

	l.Debug().Msgf("Found: %#v", res)

	return res, nil
}

func (r *SubmissionRepository) AllByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Submission, error) {
	l := logger.Ctx(ctx).With().Str("method", "All").Logger()

	const SQL = `
		SELECT
			id,
			created_at,
			user_id,
			assessment_id,
			file_name,
			file_url,
			external_id,
			result_date,
			result_pass,
			result_text
		FROM Submissions
		WHERE user_id=$1
		ORDER BY created_at
`
	rows, err := r.db.QueryContext(ctx, SQL, userID)
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

	res := make([]*model.Submission, 0)

	for rows.Next() {
		if err := rows.Err(); err != nil {
			l.Debug().Err(err).Send()
			return nil, fmt.Errorf("rows next: %w", err)
		}
		m := &model.Submission{}
		if err := rows.Scan(
			&m.ID,
			&m.CreatedAt,
			m.UserID,
			m.AssessmentID,
			m.FileName,
			m.FileURL,
			m.ExternalID,
			m.ResultDate,
			m.ResultPass,
			m.ResultText,
		); err != nil {
			l.Debug().Err(err).Send()
			return nil, fmt.Errorf("scan: %w", err)
		}
		res = append(res, m)
	}

	l.Debug().Msgf("Found: %#v", res)

	return res, nil
}
