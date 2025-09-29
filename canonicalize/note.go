package canonicalize

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	DocumentType  string
	SigningMethod string
)

type Metadata struct {
	AppointmentId string    `bson:"appointment_id"`
	CreatedAt     time.Time `bson:"created_at"`
	DoctorId      string    `bson:"doctor_id"`
	DocumentType  string    `bson:"document_type"`
	NoteId        string    `bson:"note_id"`
	PatientId     string    `bson:"patient_id"`
	SignedAt      time.Time `bson:"signed_at"`
	SignedWith    string    `bson:"signed_with"`
}

type Attachment struct {
	SHA256Hash string `bson:"sha256_hash"`
	Filename   string `bson:"filename"`
}

type CanonicalNote struct {
	Content     string       `bson:"content"`
	Metadata    Metadata     `bson:"metadata"`
	Attachments []Attachment `bson:"attachments"`
}

type NoteInput struct {
	DoctorId     string
	DocumentType string
	NoteId       string
	PatientId    string
	Content      string
	SignedWith   string
	Attachments  []Attachment
}

func NewCanonicalNote(i *NoteInput) *CanonicalNote {
	n := &CanonicalNote{
		Metadata: Metadata{
			CreatedAt:    time.Now(),
			SignedAt:     time.Now(),
			DoctorId:     i.DoctorId,
			DocumentType: i.DocumentType,
			NoteId:       i.NoteId,
			PatientId:    i.PatientId,
			SignedWith:   i.SignedWith,
		},
		Attachments: []Attachment{},
	}
	return n
}

func (n *CanonicalNote) Canonicalize() ([]byte, error) {
	bytes, err := bson.Marshal(n)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
