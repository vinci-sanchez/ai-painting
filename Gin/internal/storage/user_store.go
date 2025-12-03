package storage

import (
	"context"
	"database/sql"
	"encoding/json"
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

CREATE TABLE IF NOT EXISTS user_comics (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	page_number INT NOT NULL DEFAULT 1,
	image_base64 TEXT NOT NULL,
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_comics_user_id ON user_comics(user_id);
`

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrEmptyUsername      = errors.New("username cannot be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrUserNotFound       = errors.New("user does not exist")
)

// User 表示一个存储的账号信息（密码使用散列值）
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ComicRecord 描述用户已生成并保存的漫画页
type ComicRecord struct {
	ID          int64           `json:"id"`
	UserID      int64           `json:"user_id"`
	Title       string          `json:"title"`
	PageNumber  int             `json:"page_number"`
	ImageBase64 string          `json:"image_base64"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
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
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE normalized_username = $1
	`, normalized).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
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

func (s *UserStore) lookupUserID(ctx context.Context, username string) (int64, error) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if normalized == "" {
		return 0, ErrEmptyUsername
	}

	var id int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM users
		WHERE normalized_username = $1
	`, normalized).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, fmt.Errorf("查询用户失败: %w", err)
	}
	return id, nil
}

// SaveComicForUser 将漫画页保存到用户名下
func (s *UserStore) SaveComicForUser(ctx context.Context, username string, comic ComicRecord) (ComicRecord, error) {
	if strings.TrimSpace(comic.Title) == "" {
		return ComicRecord{}, errors.New("title cannot be empty")
	}
	if strings.TrimSpace(comic.ImageBase64) == "" {
		return ComicRecord{}, errors.New("image cannot be empty")
	}
	if comic.PageNumber <= 0 {
		comic.PageNumber = 1
	}
	if len(comic.Metadata) == 0 {
		comic.Metadata = json.RawMessage(`{}`)
	}

	userID, err := s.lookupUserID(ctx, username)
	if err != nil {
		return ComicRecord{}, err
	}

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO user_comics (user_id, title, page_number, image_base64, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, userID, comic.Title, comic.PageNumber, comic.ImageBase64, comic.Metadata).Scan(&comic.ID, &comic.CreatedAt)
	if err != nil {
		return ComicRecord{}, fmt.Errorf("保存漫画失败: %w", err)
	}

	comic.UserID = userID
	return comic, nil
}

// ListComicsForUser 返回用户历史保存的漫画
func (s *UserStore) ListComicsForUser(ctx context.Context, username string) ([]ComicRecord, error) {
	userID, err := s.lookupUserID(ctx, username)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, page_number, image_base64, metadata, created_at
		FROM user_comics
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("查询漫画失败: %w", err)
	}
	defer rows.Close()

	var records []ComicRecord
	for rows.Next() {
		var record ComicRecord
		record.UserID = userID
		if err := rows.Scan(&record.ID, &record.Title, &record.PageNumber, &record.ImageBase64, &record.Metadata, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("读取漫画记录失败: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}
