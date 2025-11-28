package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

const schemaInitQuery = `
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	username TEXT NOT NULL,
	normalized_username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
)

// User 表示一个存储的账号信息（密码使用散列值）
type User struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserStore 负责将用户账号信息持久化到 PostgreSQL
type UserStore struct {
	db *sql.DB
}

// NewUserStore 创建一个用户存储实例，并确保数据库中存在所需表
func NewUserStore(dsn string) (*UserStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("数据库连接串不能为空")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库不可用: %w", err)
	}

	store := &UserStore{db: db}
	if err := store.ensureSchema(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *UserStore) ensureSchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schemaInitQuery); err != nil {
		return fmt.Errorf("初始化用户表失败: %w", err)
	}
	return nil
}

// SaveUser 保存或更新一个账号及其密码散列
func (s *UserStore) SaveUser(username, password string) error {
	return s.saveUserInternal(username, password, false)
}

// RegisterUser 新注册账号，若账号已存在则返回错误
func (s *UserStore) RegisterUser(username, password string) error {
	return s.saveUserInternal(username, password, true)
}

// GetUser 返回指定用户名的账号信息
func (s *UserStore) GetUser(username string) (User, bool) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if normalized == "" {
		return User{}, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err := s.db.QueryRowContext(ctx, `
		SELECT username, password_hash, created_at, updated_at
		FROM users
		WHERE normalized_username = $1
	`, normalized).Scan(&user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false
		}
		return User{}, false
	}

	return user, true
}

// VerifyUser 校验用户名与密码是否匹配
func (s *UserStore) VerifyUser(username, password string) error {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if normalized == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(password) == "" {
		return ErrEmptyPassword
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var hash string
	err := s.db.QueryRowContext(ctx, `
		SELECT password_hash
		FROM users
		WHERE normalized_username = $1
	`, normalized).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

func (s *UserStore) saveUserInternal(username, password string, failIfExists bool) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return ErrEmptyUsername
	}
	if strings.TrimSpace(password) == "" {
		return ErrEmptyPassword
	}

	normalized := strings.ToLower(username)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("加密密码失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if failIfExists {
		result, err := s.db.ExecContext(ctx, `
			INSERT INTO users (username, normalized_username, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (normalized_username) DO NOTHING
		`, username, normalized, string(hash))
		if err != nil {
			return fmt.Errorf("注册用户失败: %w", err)
		}
		rows, err := result.RowsAffected()
		if err == nil && rows == 0 {
			return ErrUserExists
		}
		return nil
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, normalized_username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (normalized_username)
		DO UPDATE SET
			username = EXCLUDED.username,
			password_hash = EXCLUDED.password_hash,
			updated_at = NOW()
	`, username, normalized, string(hash)); err != nil {
		return fmt.Errorf("保存用户失败: %w", err)
	}
	return nil
}
