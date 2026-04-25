package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"pocketful/internal/config"
	"pocketful/internal/models"
	"pocketful/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// KYCStatusResponse combines a KYC session with its uploaded documents.
type KYCStatusResponse struct {
	KYC       *models.KYC        `json:"kyc"`
	Documents []models.Document  `json:"documents"`
}

// KYCService handles KYC business logic.
type KYCService struct {
	kycRepo  *repository.KYCRepository
	docRepo  *repository.DocumentRepository
}

// NewKYCService creates a new KYCService.
func NewKYCService(kycRepo *repository.KYCRepository, docRepo *repository.DocumentRepository) *KYCService {
	return &KYCService{kycRepo: kycRepo, docRepo: docRepo}
}

// Initiate creates a new KYC session for a user.
// Returns an error if the user already has an active session.
func (s *KYCService) Initiate(ctx context.Context, userID primitive.ObjectID) (*models.KYC, error) {
	exists, err := s.kycRepo.ExistsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		// Return existing session
		return s.kycRepo.FindByUserID(ctx, userID)
	}

	kyc := &models.KYC{
		UserID:    userID,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.kycRepo.Create(ctx, kyc)
}

// UploadDocument saves a document file locally and records its metadata in the database.
// After each upload, it asynchronously checks if all required documents are present.
func (s *KYCService) UploadDocument(
	ctx context.Context,
	userID primitive.ObjectID,
	docType models.DocumentType,
	fileHeader *multipart.FileHeader,
) (*models.Document, error) {
	// Validate document type
	validType := false
	for _, t := range models.RequiredDocumentTypes {
		if t == docType {
			validType = true
			break
		}
	}
	if !validType {
		return nil, fmt.Errorf("invalid document type: %s (must be PAN, AADHAAR, or SELFIE)", docType)
	}

	// Get existing KYC session
	kyc, err := s.kycRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("KYC session not found; please initiate KYC first")
		}
		return nil, err
	}

	// Reject uploads if KYC is already verified or under final review
	if kyc.Status == models.StatusVerified {
		return nil, errors.New("KYC is already verified")
	}

	// Save file to local storage
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Build: uploads/<userID>/<kycID>/
	uploadDir := filepath.Join(config.AppConfig.UploadDir, userID.Hex(), kyc.ID.Hex())
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Filename: PAN_<timestamp>.<ext>
	ext := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s_%d%s", string(docType), time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Upsert document metadata in MongoDB
	doc := &models.Document{
		KycID:       kyc.ID,
		UserID:      userID,
		Type:        docType,
		FileName:    fileName,
		FilePath:    filePath,
		ContentType: fileHeader.Header.Get("Content-Type"),
		SizeBytes:   fileHeader.Size,
	}

	savedDoc, err := s.docRepo.UpsertByKycIDAndType(ctx, doc)
	if err != nil {
		return nil, err
	}

	// Asynchronously check if all required docs are uploaded and advance status
	go s.checkAndAdvanceStatus(kyc.ID, userID)

	return savedDoc, nil
}

// checkAndAdvanceStatus is run as a goroutine after each document upload.
// If all required document types are present, it moves KYC to UNDER_REVIEW.
func (s *KYCService) checkAndAdvanceStatus(kycID, userID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kyc, err := s.kycRepo.FindByID(ctx, kycID)
	if err != nil || kyc.Status != models.StatusPending {
		return
	}

	docs, err := s.docRepo.FindByKycID(ctx, kycID)
	if err != nil {
		log.Printf("[KYC Worker] Error fetching documents for kycID=%s: %v", kycID.Hex(), err)
		return
	}

	uploadedTypes := make(map[models.DocumentType]bool)
	for _, d := range docs {
		uploadedTypes[d.Type] = true
	}

	allPresent := true
	for _, required := range models.RequiredDocumentTypes {
		if !uploadedTypes[required] {
			allPresent = false
			break
		}
	}

	if allPresent {
		log.Printf("[KYC Worker] All documents received for kycID=%s. Moving to UNDER_REVIEW.", kycID.Hex())
		if err := s.kycRepo.UpdateStatusOnly(ctx, kycID, models.StatusUnderReview); err != nil {
			log.Printf("[KYC Worker] Failed to update status for kycID=%s: %v", kycID.Hex(), err)
		}
	}
}

// GetStatus retrieves the KYC session and uploaded documents for a user.
func (s *KYCService) GetStatus(ctx context.Context, userID primitive.ObjectID) (*KYCStatusResponse, error) {
	kyc, err := s.kycRepo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("no KYC session found")
		}
		return nil, err
	}

	docs, err := s.docRepo.FindByKycID(ctx, kyc.ID)
	if err != nil {
		return nil, err
	}

	return &KYCStatusResponse{KYC: kyc, Documents: docs}, nil
}

// AdminVerify allows an admin to approve or reject a KYC session.
func (s *KYCService) AdminVerify(ctx context.Context, req models.AdminVerifyRequest, adminID primitive.ObjectID) error {
	kycID, err := primitive.ObjectIDFromHex(req.KYCID)
	if err != nil {
		return errors.New("invalid KYC ID")
	}

	kyc, err := s.kycRepo.FindByID(ctx, kycID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("KYC session not found")
		}
		return err
	}

	if kyc.Status == models.StatusVerified || kyc.Status == models.StatusRejected {
		return fmt.Errorf("KYC is already %s and cannot be updated", kyc.Status)
	}

	var newStatus models.KYCStatus
	switch req.Action {
	case "APPROVE":
		newStatus = models.StatusVerified
	case "REJECT":
		if req.RejectionNote == "" {
			return errors.New("rejection_note is required when rejecting KYC")
		}
		newStatus = models.StatusRejected
	default:
		return errors.New("invalid action: must be APPROVE or REJECT")
	}

	return s.kycRepo.UpdateStatus(ctx, kycID, newStatus, adminID, req.RejectionNote)
}
