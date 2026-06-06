package audit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/clineomx/trussrod/database"
	"github.com/clineomx/trussrod/logging"
)

type spyDB struct {
	mu   sync.Mutex
	logs []*Log
	done chan struct{}
}

func newSpyDB() *spyDB {
	return &spyDB{done: make(chan struct{}, 100)}
}

func (s *spyDB) Exec(_ context.Context, _ string, args ...any) (database.Result, error) {
	log := &Log{
		EventType: args[0].(string),
		ActorID:   args[1].(string),
		ActorRole: args[2].(string),
		Action:    args[6].(string),
	}
	if v := args[9]; v != nil {
		ip := v.(*string)
		log.IPAddress = ip
	}

	s.mu.Lock()
	s.logs = append(s.logs, log)
	s.mu.Unlock()
	s.done <- struct{}{}
	return spyResult{}, nil
}

func (s *spyDB) Query(context.Context, string, ...any) (database.Rows, error) {
	panic("not implemented")
}

func (s *spyDB) QueryRow(context.Context, string, ...any) database.Row {
	panic("not implemented")
}

func (s *spyDB) BeginTx(context.Context, any) (database.Tx, error) {
	panic("not implemented")
}

func (s *spyDB) Ping(context.Context) error { return nil }

type spyResult struct{}

func (spyResult) RowsAffected() int64 { return 1 }

func waitForPersistedLog(t *testing.T, db *spyDB) *Log {
	t.Helper()
	select {
	case <-db.done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for audit log to be persisted")
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	if len(db.logs) == 0 {
		t.Fatal("expected persisted audit log")
	}
	return db.logs[len(db.logs)-1]
}

func TestDatabaseAuditor_WithFieldsLogPersistsMetadata(t *testing.T) {
	db := newSpyDB()
	logger := logging.New(logging.Config{Level: "info", ServiceName: "test", Environment: "test"})
	auditor := NewDatabaseAuditor(db, logger)

	ip := "192.168.1.1"
	scoped := auditor.WithFields(map[string]any{
		"actor_id":   "user-123",
		"actor_role": "doctor",
		"ip_address": ip,
	})

	scoped.Log(context.Background(), &Log{
		EventType: "note.read",
		Action:    "read",
	})

	got := waitForPersistedLog(t, db)
	if got.ActorID != "user-123" {
		t.Fatalf("expected actor_id user-123, got %q", got.ActorID)
	}
	if got.ActorRole != "doctor" {
		t.Fatalf("expected actor_role doctor, got %q", got.ActorRole)
	}
	if got.IPAddress == nil || *got.IPAddress != ip {
		t.Fatalf("expected ip_address %q, got %v", ip, got.IPAddress)
	}
}

func TestDatabaseAuditor_RootWorkerAppliesChildWithFieldsMetadata(t *testing.T) {
	db := newSpyDB()
	logger := logging.New(logging.Config{Level: "info", ServiceName: "test", Environment: "test"})
	root := NewDatabaseAuditor(db, logger)

	child := root.WithFields(map[string]any{
		"actor_id":   "child-actor",
		"actor_role": "nurse",
	})

	child.Log(context.Background(), &Log{
		EventType: "patient.view",
		Action:    "read",
	})

	got := waitForPersistedLog(t, db)
	if got.ActorID != "child-actor" {
		t.Fatalf("expected worker to persist child actor_id, got %q", got.ActorID)
	}
	if got.ActorRole != "nurse" {
		t.Fatalf("expected worker to persist child actor_role, got %q", got.ActorRole)
	}
}
