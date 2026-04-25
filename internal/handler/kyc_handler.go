package handler

import (
	"net/http"
	"strings"

	"pocketful/internal/middleware"
	"pocketful/internal/models"
	"pocketful/internal/service"
	"pocketful/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// KYCHandler manages KYC-related HTTP endpoints.
type KYCHandler struct {
	kycService *service.KYCService
}

// NewKYCHandler creates a new KYCHandler.
func NewKYCHandler(kycService *service.KYCService) *KYCHandler {
	return &KYCHandler{kycService: kycService}
}

// getUserIDFromContext extracts the authenticated user's ObjectID from the Gin context.
func getUserIDFromContext(c *gin.Context) (primitive.ObjectID, bool) {
	userIDStr, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return primitive.NilObjectID, false
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
		return primitive.NilObjectID, false
	}
	return userID, true
}

// Initiate godoc
// @Summary      Initiate KYC
// @Description  Creates a new KYC session for the authenticated user (or returns the existing one)
// @Tags         kyc
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  models.KYC
// @Failure      401  {object}  map[string]string
// @Router       /kyc/initiate [post]
func (h *KYCHandler) Initiate(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	kyc, err := h.kycService.Initiate(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KYC session ready",
		"kyc":     kyc,
	})
}

// Upload godoc
// @Summary      Upload KYC Document
// @Description  Uploads a document (PAN, AADHAAR, or SELFIE) for KYC verification
// @Tags         kyc
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        doc_type  formData  string  true  "Document type: PAN, AADHAAR, SELFIE"
// @Param        file      formData  file    true  "Document file"
// @Success      200       {object}  models.Document
// @Failure      400       {object}  map[string]string
// @Router       /kyc/upload [post]
func (h *KYCHandler) Upload(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	docTypeStr := strings.ToUpper(strings.TrimSpace(c.PostForm("doc_type")))
	if docTypeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doc_type is required (PAN, AADHAAR, or SELFIE)"})
		return
	}

	// Validate PAN number if doc_type is PAN
	if docTypeStr == "PAN" {
		panNumber := c.PostForm("pan_number")
		if panNumber != "" && !utils.IsValidPAN(panNumber) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PAN format. Must be like: ABCDE1234F"})
			return
		}
	}

	// Validate Aadhaar number if doc_type is AADHAAR
	if docTypeStr == "AADHAAR" {
		aadhaarNumber := c.PostForm("aadhaar_number")
		if aadhaarNumber != "" && !utils.IsValidAadhaar(aadhaarNumber) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Aadhaar format. Must be 12 digits"})
			return
		}
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// Validate file size (max 5MB)
	const maxFileSize = 5 * 1024 * 1024
	if fileHeader.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size must not exceed 5MB"})
		return
	}

	doc, err := h.kycService.UploadDocument(c.Request.Context(), userID, models.DocumentType(docTypeStr), fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Document uploaded successfully",
		"document": doc,
	})
}

// Status godoc
// @Summary      Get KYC Status
// @Description  Returns the current KYC status and uploaded documents for the authenticated user
// @Tags         kyc
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  service.KYCStatusResponse
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /kyc/status [get]
func (h *KYCHandler) Status(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		return
	}

	resp, err := h.kycService.GetStatus(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "no KYC session found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
