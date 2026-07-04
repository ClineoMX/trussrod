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

type LogLevel string

const (
	Success LogLevel = "success"
	Info    LogLevel = "info"
	Warning LogLevel = "warning"
	Error   LogLevel = "error"
)

type Log struct {
	EventType    string   `json:"event_type"`
	Level        LogLevel `json:"level"`
	ActorID      string   `json:"actor_id"`
	ActorRole    string   `json:"actor_role"`
	ResourcePath *string  `json:"resource_path"`
	IPAddress    *string  `json:"ip_address"`
	UserAgent    *string  `json:"user_agent"`
	RequestID    *string  `json:"request_id"`
	SessionID    *string  `json:"session_id"`
	Metadata     any      `json:"metadata"`
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
	if _, ok := fields["session_id"]; ok {
		l.SessionID = stringPtrFromField(fields, "session_id")
	}
	if _, ok := fields["user_agent"]; ok {
		l.UserAgent = stringPtrFromField(fields, "user_agent")
	}
	if _, ok := fields["request_id"]; ok {
		l.RequestID = stringPtrFromField(fields, "request_id")
	}
	if l.IPAddress == nil {
		l.IPAddress = stringPtrFromField(fields, "ip_address")
	}
	if l.SessionID == nil {
		l.SessionID = stringPtrFromField(fields, "session_id")
	}
	if l.UserAgent == nil {
		l.UserAgent = stringPtrFromField(fields, "user_agent")
	}
	if l.RequestID == nil {
		l.RequestID = stringPtrFromField(fields, "request_id")
	}
	if l.ResourcePath == nil {
		l.ResourcePath = stringPtrFromField(fields, "resource_path")
	}
	if l.Metadata == nil {
		l.Metadata = fields["metadata"]
	}
}

func (a *DatabaseAuditor) Write(ctx context.Context, log *Log) error {
	log.UpdateFromFields(a.fields)
	_, err := a.db.Exec(ctx, `
		INSERT INTO audit_log (event_type, actor_id, actor_role, resource_path, metadata, ip_address, session_id, user_agent, request_id, level)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`, log.EventType, log.ActorID, log.ActorRole, log.ResourcePath, log.Metadata, log.IPAddress, log.SessionID, log.UserAgent, log.RequestID, log.Level)
	return err
}

func (a *DatabaseAuditor) Log(ctx context.Context, log *Log) {
	log.UpdateFromFields(a.fields)
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
