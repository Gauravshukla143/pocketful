package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// KYCStatus represents the lifecycle state of a KYC application.
type KYCStatus string

const (
	StatusPending     KYCStatus = "PENDING"
	StatusUnderReview KYCStatus = "UNDER_REVIEW"
	StatusVerified    KYCStatus = "VERIFIED"
	StatusRejected    KYCStatus = "REJECTED"
)

// KYC represents a KYC session for a user.
type KYC struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"    json:"id"`
	UserID         primitive.ObjectID `bson:"user_id"          json:"user_id"`
	Status         KYCStatus          `bson:"status"           json:"status"`
	RejectionNote  string             `bson:"rejection_note"   json:"rejection_note,omitempty"`
	ReviewedBy     primitive.ObjectID `bson:"reviewed_by"      json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time         `bson:"reviewed_at"      json:"reviewed_at,omitempty"`
	CreatedAt      time.Time          `bson:"created_at"       json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"       json:"updated_at"`
}

// AdminVerifyRequest is the payload for an admin to verify or reject a KYC session.
type AdminVerifyRequest struct {
	KYCID         string `json:"kyc_id"         binding:"required"`
	Action        string `json:"action"         binding:"required,oneof=APPROVE REJECT"`
	RejectionNote string `json:"rejection_note"`
}
