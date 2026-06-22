package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/user"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5434/cdc_audit_timeline_service?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db, logger.New())

	t.Run("success", func(t *testing.T) {
		input := user.CreateInput{
			FirstName: "John",
			LastName:  "Doe",
			Email:     uuid.NewString() + "@example.com",
			CreatedBy: uuid.NewString(),
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE email = $1", input.Email)
		})

		created, err := repo.Create(context.Background(), input)
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected a generated user ID")
		}
		if created.FirstName != input.FirstName || created.LastName != input.LastName || created.Email != input.Email {
			t.Errorf("created user = %+v, want input %+v", created, input)
		}
		if created.CreatedBy != input.CreatedBy {
			t.Errorf("created by = %q, want %q", created.CreatedBy, input.CreatedBy)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		input := user.CreateInput{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE email = $1", input.Email)
		})

		if _, err := repo.Create(context.Background(), input); err != nil {
			t.Fatalf("create existing user: %v", err)
		}
		if _, err := repo.Create(context.Background(), input); err == nil {
			t.Fatal("expected duplicate email error")
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db, logger.New())

	t.Run("success", func(t *testing.T) {
		createInput := user.CreateInput{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			CreatedBy: uuid.NewString(),
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE email = $1", createInput.Email)
		})

		created, err := repo.Create(context.Background(), createInput)
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected a generated user ID")
		}

		updateInput := user.UpdateInput{
			ID:        created.ID.String(),
			FirstName: "John Updated",
			LastName:  "Doe Updated",
			Email:     createInput.Email,
			UpdatedBy: uuid.NewString(),
		}
		updated, err := repo.Update(context.Background(), updateInput)
		if err != nil {
			t.Fatalf("update user: %v", err)
		}
		if updated.ID == uuid.Nil {
			t.Fatal("expected a generated user ID")
		}
		if updated.FirstName != updateInput.FirstName || updated.LastName != updateInput.LastName || updated.Email != updateInput.Email {
			t.Errorf("updated user = %+v, want input %+v", updated, updateInput)
		}
		if updated.UpdatedBy != updateInput.UpdatedBy {
			t.Errorf("updated by = %q, want %q", updated.UpdatedBy, updateInput.UpdatedBy)
		}
	})

	t.Run("user not exists", func(t *testing.T) {
		updateInput := user.UpdateInput{
			ID:        uuid.NewString(),
			FirstName: "John Updated",
			LastName:  "Doe Updated",
			Email:     "john@example.com",
			UpdatedBy: uuid.NewString(),
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE email = $1", updateInput.Email)
		})
		_, err := repo.Update(context.Background(), updateInput)
		if err == nil {
			t.Fatalf("expected error in update user: %v", err)
		}
	})
}

func TestUserRepository_GetById(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db, logger.New())

	t.Run("success", func(t *testing.T) {
		createInput := user.CreateInput{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
			CreatedBy: uuid.NewString(),
		}
		t.Cleanup(func() {
			_, _ = db.Exec("DELETE FROM users WHERE email = $1", createInput.Email)
		})

		created, err := repo.Create(context.Background(), createInput)
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected a generated user ID")
		}

		existingUser, err := repo.GetByID(context.Background(), created.ID.String())
		if err != nil {
			t.Fatalf("get by id user: %v", err)
		}
		if existingUser.ID == uuid.Nil {
			t.Fatal("expected a generated user ID")
		}
		if existingUser.FirstName != createInput.FirstName || existingUser.LastName != createInput.LastName || existingUser.Email != createInput.Email {
			t.Errorf("get by id user = %+v, want input %+v", existingUser, createInput)
		}
		if existingUser.CreatedBy != createInput.CreatedBy {
			t.Errorf("createdBy by = %q, want %q", existingUser.CreatedBy, createInput.CreatedBy)
		}
	})

	t.Run("user not exists", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), uuid.NewString())
		if err == nil {
			t.Fatalf("expected error in get by id user: %v", err)
		}
	})
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db, logger.New())

	t.Run("success", func(t *testing.T) {

		users := []user.CreateInput{
			{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				CreatedBy: uuid.NewString(),
			},
			{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john2@example.com",
				CreatedBy: uuid.NewString(),
			},
		}

		t.Cleanup(func() {
			for _, u := range users {
				_, _ = db.Exec("DELETE FROM users WHERE email = $1", u.Email)
			}
		})

		for _, u := range users {
			_, err := repo.Create(context.Background(), u)
			if err != nil {
				t.Fatalf("failed to create test user %s: %v", u.Email, err)
			}
		}

		_, userCount, err := repo.List(context.Background(), 10, 0, "created_at", "desc")
		if err != nil {
			t.Fatalf("list user: %v", err)
		}
		if userCount == 0 {
			t.Errorf("got 0 user count")
		}
	})
}
