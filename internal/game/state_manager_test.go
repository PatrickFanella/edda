package game

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// mockQuerier implements statedb.Querier for testing.
type mockQuerier struct {
	users       map[string]statedb.User
	nextUserID  pgtype.UUID
	createCount int
}

func newMockQuerier() *mockQuerier {
	return &mockQuerier{
		users:      make(map[string]statedb.User),
		nextUserID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
	}
}

func (m *mockQuerier) GetUserByName(_ context.Context, name string) (statedb.User, error) {
	u, ok := m.users[name]
	if !ok {
		return statedb.User{}, pgx.ErrNoRows
	}
	return u, nil
}

func (m *mockQuerier) CreateUser(_ context.Context, name string) (statedb.User, error) {
	m.createCount++
	u := statedb.User{ID: m.nextUserID, Name: name}
	m.users[name] = u
	return u, nil
}

func (m *mockQuerier) GetUserByID(_ context.Context, id pgtype.UUID) (statedb.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return statedb.User{}, pgx.ErrNoRows
}

func (m *mockQuerier) ListUsers(_ context.Context) ([]statedb.User, error) {
	var out []statedb.User
	for _, u := range m.users {
		out = append(out, u)
	}
	return out, nil
}

func (m *mockQuerier) UpdateUser(_ context.Context, arg statedb.UpdateUserParams) (statedb.User, error) {
	return statedb.User{}, pgx.ErrNoRows
}

func (m *mockQuerier) DeleteUser(_ context.Context, _ pgtype.UUID) error {
	return nil
}

func (m *mockQuerier) Ping(_ context.Context) (int32, error) {
	return 1, nil
}

func TestGetOrCreateDefaultUser_Creates(t *testing.T) {
	mq := newMockQuerier()
	sm := newStateManagerWithQuerier(mq)

	u, err := sm.GetOrCreateDefaultUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Player" {
		t.Fatalf("expected name Player, got %q", u.Name)
	}
	if mq.createCount != 1 {
		t.Fatalf("expected 1 create call, got %d", mq.createCount)
	}
}

func TestGetOrCreateDefaultUser_ReturnsExisting(t *testing.T) {
	mq := newMockQuerier()
	mq.users["Player"] = statedb.User{
		ID:   pgtype.UUID{Bytes: [16]byte{42}, Valid: true},
		Name: "Player",
	}
	sm := newStateManagerWithQuerier(mq)

	u, err := sm.GetOrCreateDefaultUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Player" {
		t.Fatalf("expected name Player, got %q", u.Name)
	}
	if mq.createCount != 0 {
		t.Fatalf("should not create when user exists, got %d create calls", mq.createCount)
	}
}

func TestGetOrCreateDefaultUser_CalledTwice(t *testing.T) {
	mq := newMockQuerier()
	sm := newStateManagerWithQuerier(mq)

	u1, err := sm.GetOrCreateDefaultUser(context.Background())
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	u2, err := sm.GetOrCreateDefaultUser(context.Background())
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if u1.Name != u2.Name {
		t.Fatalf("expected same user, got %q and %q", u1.Name, u2.Name)
	}
	if mq.createCount != 1 {
		t.Fatalf("should create only once, got %d", mq.createCount)
	}
}
