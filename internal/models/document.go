package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DocumentType represents the type of an uploaded KYC document.
type DocumentType string

const (
	DocTypePAN     DocumentType = "PAN"
	DocTypeAadhaar DocumentType = "AADHAAR"
	DocTypeSelfie  DocumentType = "SELFIE"
)

// RequiredDocumentTypes is the set of documents required to complete KYC.
var RequiredDocumentTypes = []DocumentType{DocTypePAN, DocTypeAadhaar, DocTypeSelfie}

// Document represents a single uploaded document associated with a KYC session.
type Document struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"  json:"id"`
	KycID       primitive.ObjectID `bson:"kyc_id"         json:"kyc_id"`
	UserID      primitive.ObjectID `bson:"user_id"        json:"user_id"`
	Type        DocumentType       `bson:"type"           json:"type"`
	FileName    string             `bson:"file_name"      json:"file_name"`
	FilePath    string             `bson:"file_path"      json:"file_path"`
	ContentType string             `bson:"content_type"   json:"content_type"`
	SizeBytes   int64              `bson:"size_bytes"     json:"size_bytes"`
	UploadedAt  time.Time          `bson:"uploaded_at"    json:"uploaded_at"`
}
