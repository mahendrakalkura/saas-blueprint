package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (id, user_id, key, filename, mime_type, size, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	file.ID = uuid.New().String()
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		file.ID,
		file.UserID,
		file.Key,
		file.Filename,
		file.MimeType,
		file.Size,
		file.Metadata,
		file.CreatedAt,
		file.UpdatedAt,
	).Scan(&file.ID, &file.CreatedAt, &file.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id string) (*models.File, error) {
	query := `
		SELECT id, user_id, key, filename, mime_type, size, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`

	file := &models.File{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID,
		&file.UserID,
		&file.Key,
		&file.Filename,
		&file.MimeType,
		&file.Size,
		&file.Metadata,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

func (r *FileRepository) GetByKey(ctx context.Context, key string) (*models.File, error) {
	query := `
		SELECT id, user_id, key, filename, mime_type, size, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE key = $1 AND deleted_at IS NULL
	`

	file := &models.File{}
	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&file.ID,
		&file.UserID,
		&file.Key,
		&file.Filename,
		&file.MimeType,
		&file.Size,
		&file.Metadata,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

func (r *FileRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.File, error) {
	query := `
		SELECT id, user_id, key, filename, mime_type, size, metadata, created_at, updated_at, deleted_at
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer rows.Close()

	files := []*models.File{}
	for rows.Next() {
		file := &models.File{}
		err := rows.Scan(
			&file.ID,
			&file.UserID,
			&file.Key,
			&file.Filename,
			&file.MimeType,
			&file.Size,
			&file.Metadata,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

func (r *FileRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE files
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}

func (r *FileRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count files: %w", err)
	}

	return count, nil
}
