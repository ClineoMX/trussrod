package canonizer

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	DocumentType string
	SigninMethod string
	option       func(*Body)
)

const (
	MedicalNote     DocumentType = "MedicalNote"
	SimpleSignature SigninMethod = "SimpleSignature"
)

type Metadata struct {
	AppointmentId string       `bson:"appointment_id"`
	CreatedAt     time.Time    `bson:"created_at"`
	DoctorId      string       `bson:"doctor_id"`
	DocumentType  DocumentType `bson:"document_type"`
	NoteId        string       `bson:"note_id"`
	PatientId     string       `bson:"patient_id"`
	SignedAt      time.Time    `bson:"signed_at"`
	SignedWith    SigninMethod `bson:"signed_with"`
}

type Attachment struct {
	SHA256Hash string `bson:"sha256_hash"`
	Filename   string `bson:"filename"`
}

type Body struct {
	Content     string       `bson:"content"`
	Metadata    Metadata     `bson:"metadata"`
	Attachments []Attachment `bson:"attachments"`
}

func WithAppointment(id string) option {
	return func(b *Body) {
		b.Metadata.AppointmentId = id
	}
}

func WithDoctor(id string) option {
	return func(b *Body) {
		b.Metadata.DoctorId = id
	}
}

func WithType(docType DocumentType) option {
	return func(b *Body) {
		b.Metadata.DocumentType = docType
	}
}

func WithId(id string) option {
	return func(b *Body) {
		b.Metadata.NoteId = id
	}
}

func WithPatient(id string) option {
	return func(b *Body) {
		b.Metadata.PatientId = id
	}
}

func WithMethod(method SigninMethod) option {
	return func(b *Body) {
		b.Metadata.SignedWith = method
	}
}

func WithContent(content string) option {
	return func(b *Body) {
		b.Content = content
	}
}

func New(opts ...option) *Body {
	b := &Body{
		Metadata: Metadata{
			CreatedAt: time.Now(),
			SignedAt:  time.Now(),
		},
		Attachments: []Attachment{},
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Body) Canonicalize() ([]byte, error) {
	bytes, err := bson.Marshal(b)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
