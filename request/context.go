package request

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/clineomx/trussrod/keys"
	"github.com/clineomx/trussrod/logging"
)

type key string
type ApiAction string
type QueryParam string

type User struct {
	Role     string
	Name     string
	Lastname string
	ID       string
	CMKARN   string
}

const (
	ClineoUser          key = "CLINEO_SESSION"
	ClineoPatient       key = "CLINEO_PATIENT"
	ClineoDek           key = "CLINEO_DEK"
	ClineoSigner        key = "CLINEO_SIGNER"
	ClineoTraceID       key = "CLINEO_TRACE_ID"
	ClineoRequestLogger key = "CLINEO_REQUEST_LOGGER"
)

const (
	Archive ApiAction = "ARCHIVE"
	Recover ApiAction = "RECOVER"
)

var ValidApiActions = []ApiAction{Archive, Recover}

const (
	Action     QueryParam = "action"
	AttachedTo QueryParam = "attached_to"
	Attachment QueryParam = "attachment"
	Email      QueryParam = "email"
	SaveAnyway QueryParam = "save_anyway"
)

func GetUser(r *http.Request) (*User, bool) {
	user, ok := r.Context().Value(ClineoUser).(*User)
	return user, ok
}

func GetPatient(r *http.Request) (string, bool) {
	patient, ok := r.Context().Value(ClineoPatient).(string)
	return patient, ok
}

func GetSigner(r *http.Request) (keys.Signer, bool) {
	s, ok := r.Context().Value(ClineoSigner).(keys.Signer)
	return s, ok
}

func GetTraceID(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(ClineoTraceID).(string)
	return id, ok
}

func GetRequestLogger(r *http.Request) (*logging.Logger, bool) {
	log, ok := r.Context().Value("CLINEO_REQUEST_LOGGER").(*logging.Logger)
	return log, ok
}

func MustGetDek(r *http.Request) []byte {
	key, ok := GetDek(r)
	if !ok {
		panic("could not retrieve dek from context")
	}
	return key
}

func MustGetUser(r *http.Request) *User {
	user, ok := GetUser(r)
	if !ok {
		panic("could not user from context")
	}
	return user
}

func MustGetPatient(r *http.Request) string {
	patient, ok := GetPatient(r)
	if !ok {
		panic("could not patient token from context")
	}
	return patient
}

func MustGetSigner(r *http.Request) keys.Signer {
	s, ok := GetSigner(r)
	if !ok {
		panic("could not get signer from context")
	}
	return s
}

func MustGetTraceID(r *http.Request) string {
	id, ok := GetTraceID(r)
	if !ok {
		panic("could not get request ID from context")
	}
	return id
}

func MustGetRequestLogger(r *http.Request) *logging.Logger {
	log, ok := GetRequestLogger(r)
	if !ok {
		panic("could not get request logger from context")
	}
	return log
}

func GetDek(r *http.Request) ([]byte, bool) {
	key, ok := r.Context().Value(ClineoDek).([]byte)
	return key, ok
}

func WithUser(r *http.Request, user *User) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, ClineoUser, user)
	return r.WithContext(ctx)
}

func WithPatient(r *http.Request, patient string) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, ClineoPatient, patient)
	return r.WithContext(ctx)
}

func WithDEK(r *http.Request, d []byte) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, ClineoDek, d)
	return r.WithContext(ctx)
}

func WithSigner(r *http.Request, s keys.Signer) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, ClineoSigner, s)
	return r.WithContext(ctx)
}

func WithTraceID(r *http.Request, rid string) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, ClineoTraceID, rid)
	return r.WithContext(ctx)
}

func WithLogger(r *http.Request, logger *logging.Logger) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, logging.RequestLogger, logger)
	return r.WithContext(ctx)
}

func GetPage(r *http.Request) int {
	query := r.URL.Query()
	queryPage := query["page"]
	var p string
	if len(queryPage) > 0 {
		p = queryPage[0]
	}
	page, err := strconv.Atoi(p)
	if err != nil || page < 1 {
		page = 1
	}
	return page
}

func GetLimitAndOffset(r *http.Request) (int, int) {
	query := r.URL.Query()
	page := GetPage(r)

	querySize := query["size"]
	var size string
	if len(querySize) > 0 {
		size = querySize[0]
	}

	l, err := strconv.Atoi(size)
	if err != nil || l < 1 {
		l = 10
	}

	return l, (page - 1) * l
}

func GetActionParam(r *http.Request) (ApiAction, bool) {
	queryAction := r.URL.Query()[string(Action)]
	var action string
	if len(queryAction) > 0 {
		action = queryAction[0]
	}
	if action == "" {
		return "", false
	}
	normalized := ApiAction(strings.ToUpper(action))

	if !slices.Contains(ValidApiActions, normalized) {
		return "", false
	}
	return normalized, true
}

func GetFirstParam(r *http.Request, param QueryParam) (string, bool) {
	query := r.URL.Query()
	values := query[string(param)]

	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}

func GetAllParams(r *http.Request, param QueryParam) ([]string, bool) {
	query := r.URL.Query()
	values := query[string(param)]

	if len(values) == 0 {
		return nil, false
	}

	return values, true
}

func GetAttachedToParam(r *http.Request) string {
	query := r.URL.Query()
	values := query["attached_to"]

	return values[0]
}
