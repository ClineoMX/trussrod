package request

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/Domedik/trussrod/identity"
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
	DomedikUser        key = "DOMEDIK_SESSION"
	DomedikIdentity    key = "DOMEDIK_IDENTITY"
	DomedikPatient     key = "DOMEDIK_PATIENT"
	DomedikDek         key = "DOMEDIK_DEK"
	DomedikCredentials key = "DOMEDIK_CREDENTIALS"
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
)

func GetIdentity(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(DomedikIdentity).(string)
	return id, ok
}

func GetUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(DomedikUser).(*User)
	return user, ok
}

func GetPatient(ctx context.Context) (string, bool) {
	patient, ok := ctx.Value(DomedikPatient).(string)
	return patient, ok
}

func GetCredentials(ctx context.Context) (*identity.Credentials, bool) {
	creds, ok := ctx.Value(DomedikCredentials).(*identity.Credentials)
	return creds, ok
}

func MustGetDek(ctx context.Context) []byte {
	key, ok := GetDek(ctx)
	if !ok {
		panic("could not retrieve dek from context")
	}
	return key
}

func MustGetIdentity(ctx context.Context) string {
	token, ok := GetIdentity(ctx)
	if !ok {
		panic("could not identity token from context")
	}
	return token
}

func MustGetUser(ctx context.Context) *User {
	user, ok := GetUser(ctx)
	if !ok {
		panic("could not user from context")
	}
	return user
}

func MustGetPatient(ctx context.Context) string {
	patient, ok := GetPatient(ctx)
	if !ok {
		panic("could not patient token from context")
	}
	return patient
}

func MustGetCredentials(ctx context.Context) *identity.Credentials {
	creds, ok := GetCredentials(ctx)
	if !ok {
		panic("could not get credentials from context")
	}
	return creds
}

func GetDek(ctx context.Context) ([]byte, bool) {
	key, ok := ctx.Value(DomedikDek).([]byte)
	return key, ok
}

func WithUser(r *http.Request, user *User) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, DomedikUser, user)
	return r.WithContext(ctx)
}

func WithIdentity(r *http.Request, identity string) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, DomedikIdentity, identity)
	return r.WithContext(ctx)
}

func WithPatient(r *http.Request, patient string) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, DomedikPatient, patient)
	return r.WithContext(ctx)
}

func WithDEK(r *http.Request, d []byte) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, DomedikDek, d)
	return r.WithContext(ctx)
}

func WithCredentials(r *http.Request, creds *identity.Credentials) *http.Request {
	parent := r.Context()
	ctx := context.WithValue(parent, DomedikCredentials, creds)
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
