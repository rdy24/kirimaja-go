package auth

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"kirimaja-go/models"
)

type fakeRepo struct {
	users       map[string]*models.User
	role        *models.Role
	createdUser *models.User
}

func (f *fakeRepo) FindUserByEmail(_ context.Context, email string) (*models.User, error) {
	return f.users[email], nil
}
func (f *fakeRepo) FindRoleByKey(context.Context, string) (*models.Role, error) {
	return f.role, nil
}
func (f *fakeRepo) CreateUser(_ context.Context, u *models.User) error {
	u.ID = 1
	if f.role != nil {
		u.Role = *f.role
	}
	f.createdUser = u
	if f.users == nil {
		f.users = map[string]*models.User{}
	}
	f.users[u.Email] = u // so the post-create reload finds it
	return nil
}

func newAuthSvc(r Repository) Service { return NewService(r, "test-secret", "24h") }

func TestLogin(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-horse"), 10)
	repo := &fakeRepo{users: map[string]*models.User{
		"a@b.c": {ID: 1, Email: "a@b.c", Password: string(hash), Role: models.Role{ID: 2}},
	}}
	s := newAuthSvc(repo)
	ctx := context.Background()

	t.Run("unknown email", func(t *testing.T) {
		if _, err := s.Login(ctx, LoginRequest{Email: "no@one.com", Password: "correct-horse"}); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})
	t.Run("wrong password", func(t *testing.T) {
		if _, err := s.Login(ctx, LoginRequest{Email: "a@b.c", Password: "nope"}); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("want ErrInvalidCredentials, got %v", err)
		}
	})
	t.Run("success", func(t *testing.T) {
		res, err := s.Login(ctx, LoginRequest{Email: "a@b.c", Password: "correct-horse"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AccessToken == "" {
			t.Fatal("expected a token")
		}
		if res.User.Email != "a@b.c" {
			t.Fatalf("wrong user echoed: %+v", res.User)
		}
	})
}

func TestRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("email already registered", func(t *testing.T) {
		repo := &fakeRepo{users: map[string]*models.User{"a@b.c": {Email: "a@b.c"}}}
		_, err := newAuthSvc(repo).Register(ctx, RegisterRequest{Email: "a@b.c", Password: "longenough"})
		if !errors.Is(err, ErrEmailRegistered) {
			t.Fatalf("want ErrEmailRegistered, got %v", err)
		}
	})

	t.Run("customer role missing", func(t *testing.T) {
		repo := &fakeRepo{users: map[string]*models.User{}, role: nil}
		_, err := newAuthSvc(repo).Register(ctx, RegisterRequest{Email: "new@b.c", Password: "longenough"})
		if err == nil {
			t.Fatal("expected error when customer role not found")
		}
	})

	t.Run("success hashes password", func(t *testing.T) {
		repo := &fakeRepo{users: map[string]*models.User{}, role: &models.Role{ID: 3, Key: "customer"}}
		res, err := newAuthSvc(repo).Register(ctx, RegisterRequest{
			Name: "X", Email: "new@b.c", Password: "plaintext-pw", PhoneNumber: "0812345678",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.AccessToken == "" {
			t.Fatal("expected a token")
		}
		if repo.createdUser.Password == "plaintext-pw" {
			t.Fatal("password stored in plaintext")
		}
		if bcrypt.CompareHashAndPassword([]byte(repo.createdUser.Password), []byte("plaintext-pw")) != nil {
			t.Fatal("stored password is not a valid bcrypt hash of the input")
		}
	})
}
