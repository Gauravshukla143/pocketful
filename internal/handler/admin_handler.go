package handler

import (
	"net/http"

	"pocketful/internal/middleware"
	"pocketful/internal/models"
	"pocketful/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminHandler manages admin-only HTTP endpoints.
type AdminHandler struct {
	kycService *service.KYCService
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(kycService *service.KYCService) *AdminHandler {
	return &AdminHandler{kycService: kycService}
}

// Verify godoc
// @Summary      Verify or Reject KYC
// @Description  Admin endpoint to approve or reject a KYC session
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.AdminVerifyRequest  true  "Verification payload"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /admin/verify [post]
func (h *AdminHandler) Verify(c *gin.Context) {
	adminIDStr, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminID, err := primitive.ObjectIDFromHex(adminIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin ID"})
		return
	}

	var req models.AdminVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.kycService.AdminVerify(c.Request.Context(), req, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action := "verified"
	if req.Action == "REJECT" {
		action = "rejected"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KYC session " + action + " successfully",
		"kyc_id":  req.KYCID,
		"action":  req.Action,
	})
}
