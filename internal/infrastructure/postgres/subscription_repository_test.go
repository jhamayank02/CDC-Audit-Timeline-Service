package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/domain/subscription"
	"github.com/jhamayank02/CDC-Audit-Timeline-Service/internal/infrastructure/logger"
)

func TestSubscriptionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepository(db, logger.New())
	autoRenew := true

	t.Run("success", func(t *testing.T) {
		input := subscription.CreateInput{
			UserID:    "11111111-1111-1111-1111-111111111111",
			PlanName:  "test-plan-" + uuid.NewString(),
			Status:    "active",
			StartDate: "2026-06-01T00:00:00Z",
			EndDate:   "2026-07-01T00:00:00Z",
			AutoRenew: &autoRenew,
			CreatedBy: uuid.NewString(),
		}
		var createdID string
		t.Cleanup(func() {
			if createdID != "" {
				_, _ = db.Exec("DELETE FROM subscriptions WHERE id = $1", createdID)
			}
		})

		created, err := repo.Create(context.Background(), input)
		if err != nil {
			t.Fatalf("create subscription: %v", err)
		}
		createdID = created.ID.String()
		if created.ID == uuid.Nil {
			t.Fatal("expected a generated subscription ID")
		}
		if created.UserID.String() != input.UserID || created.PlanName != input.PlanName || created.Status != input.Status {
			t.Errorf("created subscription = %+v, want input %+v", created, input)
		}
		if created.AutoRenew != *input.AutoRenew {
			t.Errorf("auto renew = %t, want %t", created.AutoRenew, *input.AutoRenew)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		_, err := repo.Create(context.Background(), subscription.CreateInput{
			UserID:    uuid.NewString(),
			PlanName:  "test-plan",
			Status:    "active",
			StartDate: "2026-06-01T00:00:00Z",
			EndDate:   "2026-07-01T00:00:00Z",
		})
		if err == nil {
			t.Fatal("expected foreign-key error for an unknown user")
		}
	})
}

func TestSubscriptionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepository(db, logger.New())
	autoRenew := true

	t.Run("success", func(t *testing.T) {
		created, err := repo.Create(context.Background(), subscription.CreateInput{
			UserID: "11111111-1111-1111-1111-111111111111", PlanName: "test-plan-" + uuid.NewString(), Status: "active",
			StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-01T00:00:00Z", AutoRenew: &autoRenew,
		})
		if err != nil {
			t.Fatalf("create subscription: %v", err)
		}
		t.Cleanup(func() { _, _ = db.Exec("DELETE FROM subscriptions WHERE id = $1", created.ID) })

		updatedAutoRenew := false
		input := subscription.UpdateInput{ID: created.ID.String(), Status: "cancelled", AutoRenew: &updatedAutoRenew, UpdatedBy: uuid.NewString()}
		updated, err := repo.Update(context.Background(), input)
		if err != nil {
			t.Fatalf("update subscription: %v", err)
		}
		if updated.ID != created.ID || updated.Status != input.Status || updated.AutoRenew != *input.AutoRenew {
			t.Errorf("updated subscription = %+v, want status %q and auto renew %t", updated, input.Status, *input.AutoRenew)
		}
		if updated.UpdatedBy != input.UpdatedBy {
			t.Errorf("updated by = %q, want %q", updated.UpdatedBy, input.UpdatedBy)
		}
	})

	t.Run("subscription not found", func(t *testing.T) {
		_, err := repo.Update(context.Background(), subscription.UpdateInput{ID: uuid.NewString(), Status: "cancelled"})
		if err != subscription.ErrSubscriptionNotFound {
			t.Fatalf("update error = %v, want %v", err, subscription.ErrSubscriptionNotFound)
		}
	})
}

func TestSubscriptionRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepository(db, logger.New())

	t.Run("success", func(t *testing.T) {
		created, err := repo.Create(context.Background(), subscription.CreateInput{
			UserID: "11111111-1111-1111-1111-111111111111", PlanName: "test-plan-" + uuid.NewString(), Status: "active",
			StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-01T00:00:00Z",
		})
		if err != nil {
			t.Fatalf("create subscription: %v", err)
		}
		t.Cleanup(func() { _, _ = db.Exec("DELETE FROM subscriptions WHERE id = $1", created.ID) })

		got, err := repo.GetByID(context.Background(), created.ID.String())
		if err != nil {
			t.Fatalf("get subscription: %v", err)
		}
		if got.ID != created.ID || got.PlanName != created.PlanName || got.Status != created.Status {
			t.Errorf("got subscription = %+v, want %+v", got, created)
		}
	})

	t.Run("subscription not found", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), uuid.NewString())
		if err != subscription.ErrSubscriptionNotFound {
			t.Fatalf("get error = %v, want %v", err, subscription.ErrSubscriptionNotFound)
		}
	})
}

func TestSubscriptionRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepository(db, logger.New())

	for range 2 {
		created, err := repo.Create(context.Background(), subscription.CreateInput{
			UserID: "11111111-1111-1111-1111-111111111111", PlanName: "test-plan-" + uuid.NewString(), Status: "active",
			StartDate: "2026-06-01T00:00:00Z", EndDate: "2026-07-01T00:00:00Z",
		})
		if err != nil {
			t.Fatalf("create subscription: %v", err)
		}
		t.Cleanup(func() { _, _ = db.Exec("DELETE FROM subscriptions WHERE id = $1", created.ID) })
	}

	subscriptions, total, err := repo.List(context.Background(), 10, 0, "created_at", "desc")
	if err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	if len(subscriptions) == 0 || total < len(subscriptions) {
		t.Errorf("list returned %d subscriptions with total %d", len(subscriptions), total)
	}
}
