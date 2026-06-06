package audit

import (
	"context"
	"errors"
	"maps"

	"github.com/clineomx/trussrod/apperr"
	"github.com/clineomx/trussrod/database"
	"github.com/clineomx/trussrod/logging"
)

type Auditor interface {
	Log(ctx context.Context, log *Log)
	WithFields(fields map[string]any) Auditor
}

type DatabaseAuditor struct {
	db     database.DB
	queue  chan *Log
	logger *logging.Logger
	fields map[string]any
}

func NewDatabaseAuditor(db database.DB, logger *logging.Logger) *DatabaseAuditor {
	a := &DatabaseAuditor{db: db, queue: make(chan *Log, 1000), logger: logger}
	go a.worker()
	return a
}

type Log struct {
	EventType    string  `json:"event_type"`
	ActorID      string  `json:"actor_id"`
	ActorRole    string  `json:"actor_role"`
	PatientID    *string `json:"patient_id"`
	ResourceType *string `json:"resource_type"`
	ResourceID   *string `json:"resource_id"`
	Action       string  `json:"action"`
	BeforeState  any     `json:"before_state"`
	AfterState   any     `json:"after_state"`
	IPAddress    *string `json:"ip_address"`
	SessionID    *string `json:"session_id"`
	Reason       *string `json:"reason"`
}

func stringPtrFromField(fields map[string]any, key string) *string {
	v, ok := fields[key]
	if !ok || v == nil {
		return nil
	}
	switch s := v.(type) {
	case *string:
		return s
	case string:
		return &s
	default:
		return nil
	}
}

func (l *Log) UpdateFromFields(fields map[string]any) {
	if fields == nil {
		return
	}
	if actorID, ok := fields["actor_id"]; ok {
		l.ActorID = actorID.(string)
	}
	if actorRole, ok := fields["actor_role"]; ok {
		l.ActorRole = actorRole.(string)
	}
	if eventType, ok := fields["event_type"]; ok {
		l.EventType = eventType.(string)
	}
	if action, ok := fields["action"]; ok {
		l.Action = action.(string)
	}
	if _, ok := fields["session_id"]; ok {
		l.SessionID = stringPtrFromField(fields, "session_id")
	}
	if _, ok := fields["reason"]; ok {
		l.Reason = stringPtrFromField(fields, "reason")
	}
	if _, ok := fields["patient_id"]; ok {
		l.PatientID = stringPtrFromField(fields, "patient_id")
	}
	if l.IPAddress == nil {
		l.IPAddress = stringPtrFromField(fields, "ip_address")
	}
	if l.SessionID == nil {
		l.SessionID = stringPtrFromField(fields, "session_id")
	}
	if l.Reason == nil {
		l.Reason = stringPtrFromField(fields, "reason")
	}
	if l.PatientID == nil {
		l.PatientID = stringPtrFromField(fields, "patient_id")
	}
	if l.ResourceType == nil {
		l.ResourceType = stringPtrFromField(fields, "resource_type")
	}
	if l.ResourceID == nil {
		l.ResourceID = stringPtrFromField(fields, "resource_id")
	}
	if l.BeforeState == nil {
		if v, ok := fields["before_state"]; ok {
			l.BeforeState = v
		}
	}
	if l.AfterState == nil {
		if v, ok := fields["after_state"]; ok {
			l.AfterState = v
		}
	}
}

func (a *DatabaseAuditor) Write(ctx context.Context, log *Log) error {
	log.UpdateFromFields(a.fields)
	_, err := a.db.Exec(ctx, `
		INSERT INTO audit_log (event_type, actor_id, actor_role, patient_id, resource_type, resource_id, action, before_state, after_state, ip_address, session_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`, log.EventType, log.ActorID, log.ActorRole, log.PatientID, log.ResourceType, log.ResourceID, log.Action, log.BeforeState, log.AfterState, log.IPAddress, log.SessionID, log.Reason)
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

func (a *DatabaseAuditor) WithFields(fields map[string]any) Auditor {
	newFields := make(map[string]any)
	maps.Copy(newFields, a.fields)
	maps.Copy(newFields, fields)
	return &DatabaseAuditor{
		db:     a.db,
		queue:  a.queue,
		logger: a.logger,
		fields: newFields,
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

const AuditorKey string = "CLINEO_AUDITOR"

func FromContext(ctx context.Context) Auditor {
	auditor, ok := ctx.Value(AuditorKey).(Auditor)
	if !ok {
		return nil
	}
	return auditor
}
