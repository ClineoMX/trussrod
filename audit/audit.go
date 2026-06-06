package audit

import (
	"context"
	"errors"
	"time"

	"github.com/clineomx/trussrod/apperr"
	"github.com/clineomx/trussrod/database"
	"github.com/clineomx/trussrod/logging"
)

type Auditor interface {
	Log(ctx context.Context, log *Log)
}

type DatabaseAuditor struct {
	db     database.DB
	queue  chan *Log
	logger *logging.Logger
}

func NewDatabaseAuditor(db database.DB, logger *logging.Logger) *DatabaseAuditor {
	a := &DatabaseAuditor{db: db, queue: make(chan *Log, 1000), logger: logger}
	go a.worker()
	return a
}

type Log struct {
	EventType    string    `json:"event_type"`
	ActorID      string    `json:"actor_id"`
	ActorRole    string    `json:"actor_role"`
	PatientID    *string   `json:"patient_id"`
	ResourceType *string   `json:"resource_type"`
	ResourceID   *string   `json:"resource_id"`
	Action       string    `json:"action"`
	BeforeState  any       `json:"before_state"`
	AfterState   any       `json:"after_state"`
	IPAddress    *string   `json:"ip_address"`
	SessionID    *string   `json:"session_id"`
	Reason       *string   `json:"reason"`
	Timestamp    time.Time `json:"timestamp"`
}

func (a *DatabaseAuditor) Write(ctx context.Context, log *Log) error {
	_, err := a.db.Exec(ctx, `
		INSERT INTO audit_log (timestamp, event_type, actor_id, actor_role, patient_id, resource_type, resource_id, action, before_state, after_state, ip_address, session_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);
	`, log.Timestamp, log.EventType, log.ActorID, log.ActorRole, log.PatientID, log.ResourceType, log.ResourceID, log.Action, log.BeforeState, log.AfterState, log.IPAddress, log.SessionID, log.Reason)
	if err != nil {
		return err
	}
	return nil
}

func (a *DatabaseAuditor) Log(ctx context.Context, log *Log) {
	select {
	case a.queue <- log:
	default:
		err := apperr.Internal(errors.New("AUDITOR:QUEUE_FULL"))
		a.logger.ErrorFields("audit queue full, logging to stderr", err, map[string]any{
			"event": log,
		})
	}
}

func (a *DatabaseAuditor) worker() {
	for event := range a.queue {
		if err := a.Write(context.Background(), event); err != nil {
			a.logger.ErrorFields("failed to persist audit event", err, map[string]any{
				"event": event,
			})
		}
	}
}
