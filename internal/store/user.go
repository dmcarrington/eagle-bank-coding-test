package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/domain"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, u *domain.User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users
			(id, name, email, password_hash, phone_number,
			 address_line1, address_line2, address_line3,
			 address_town, address_county, address_postcode,
			 created_timestamp, updated_timestamp)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		u.ID, u.Name, u.Email, u.PasswordHash, u.PhoneNumber,
		u.Address.Line1, nullString(u.Address.Line2), nullString(u.Address.Line3),
		u.Address.Town, u.Address.County, u.Address.Postcode,
		u.CreatedAt.UTC().Format(time.RFC3339),
		u.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrDuplicate
	}
	return err
}

func (s *UserStore) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u, err := s.scan(s.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, phone_number,
		        address_line1, address_line2, address_line3,
		        address_town, address_county, address_postcode,
		        created_timestamp, updated_timestamp
		 FROM users WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := s.scan(s.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, phone_number,
		        address_line1, address_line2, address_line3,
		        address_town, address_county, address_postcode,
		        created_timestamp, updated_timestamp
		 FROM users WHERE email = ?`, email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (s *UserStore) Update(ctx context.Context, u *domain.User) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE users SET
			name = ?, email = ?, phone_number = ?,
			address_line1 = ?, address_line2 = ?, address_line3 = ?,
			address_town = ?, address_county = ?, address_postcode = ?,
			updated_timestamp = ?
		WHERE id = ?`,
		u.Name, u.Email, u.PhoneNumber,
		u.Address.Line1, nullString(u.Address.Line2), nullString(u.Address.Line3),
		u.Address.Town, u.Address.County, u.Address.Postcode,
		u.UpdatedAt.UTC().Format(time.RFC3339),
		u.ID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("user %s not found", u.ID)
	}
	return nil
}

func (s *UserStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}

func (s *UserStore) scan(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var createdStr, updatedStr string
	var line2, line3 sql.NullString

	err := row.Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.PhoneNumber,
		&u.Address.Line1, &line2, &line3,
		&u.Address.Town, &u.Address.County, &u.Address.Postcode,
		&createdStr, &updatedStr,
	)
	if err != nil {
		return nil, err
	}
	u.Address.Line2 = line2.String
	u.Address.Line3 = line3.String
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &u, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
