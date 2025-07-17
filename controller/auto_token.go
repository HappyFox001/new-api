package controller

import (
	"net/http"
	"one-api/common"
	"one-api/model"

	"github.com/gin-gonic/gin"
)

// AutoCreateTokenRequest defines the request structure for auto token creation
type AutoCreateTokenRequest struct {
	Username    string `json:"username"`      // Username for authentication
	Password    string `json:"password"`      // Password for authentication
	TokenName   string `json:"token_name"`    // Name for the token
	RemainQuota int    `json:"remain_quota"`  // Initial quota for the token
	ExpiredTime int64  `json:"expired_time"`  // Expiration time (-1 for never expire)
	Group       string `json:"group"`         // Group for the token (optional)
}

// AutoCreateTokenResponse defines the response structure
type AutoCreateTokenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		TokenID int    `json:"token_id"`
		Key     string `json:"key"`
		UserID  int    `json:"user_id"`
	} `json:"data,omitempty"`
}

// AutoCreateToken creates a new API token for a user using username/password
// This endpoint is designed for external systems to automatically create tokens
func AutoCreateToken(c *gin.Context) {
	var req AutoCreateTokenRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, AutoCreateTokenResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, AutoCreateTokenResponse{
			Success: false,
			Message: "Username is required",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, AutoCreateTokenResponse{
			Success: false,
			Message: "Password is required",
		})
		return
	}

	if req.TokenName == "" {
		req.TokenName = "Auto-generated token"
	}

	if len(req.TokenName) > 30 {
		c.JSON(http.StatusBadRequest, AutoCreateTokenResponse{
			Success: false,
			Message: "Token name too long (max 30 characters)",
		})
		return
	}

	// Set default values
	if req.RemainQuota <= 0 {
		req.RemainQuota = 100000 // Default quota: 100,000
	}

	if req.ExpiredTime == 0 {
		req.ExpiredTime = -1 // Never expire by default
	}

	if req.Group == "" {
		req.Group = "default"
	}

	// Authenticate user by username and password
	user := model.ValidateUserCredentials(req.Username, req.Password)
	if user == nil {
		c.JSON(http.StatusUnauthorized, AutoCreateTokenResponse{
			Success: false,
			Message: "Invalid username or password",
		})
		return
	}

	// Check if user is enabled
	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusForbidden, AutoCreateTokenResponse{
			Success: false,
			Message: "User account is disabled",
		})
		return
	}

	// Generate API key
	key, err := common.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, AutoCreateTokenResponse{
			Success: false,
			Message: "Failed to generate API key: " + err.Error(),
		})
		common.SysError("failed to generate token key: " + err.Error())
		return
	}

	// Create token object
	token := model.Token{
		UserId:             user.Id,
		Name:               req.TokenName,
		Key:                key,
		CreatedTime:        common.GetTimestamp(),
		AccessedTime:       common.GetTimestamp(),
		ExpiredTime:        req.ExpiredTime,
		RemainQuota:        req.RemainQuota,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: false,
		ModelLimits:        "",
		AllowIps:           nil,
		Group:              req.Group,
	}

	// Save token to database
	err = token.Insert()
	if err != nil {
		c.JSON(http.StatusInternalServerError, AutoCreateTokenResponse{
			Success: false,
			Message: "Failed to create token: " + err.Error(),
		})
		return
	}

	// Return success response with token info
	response := AutoCreateTokenResponse{
		Success: true,
		Message: "Token created successfully",
	}
	response.Data.TokenID = token.Id
	response.Data.Key = key
	response.Data.UserID = user.Id

	c.JSON(http.StatusOK, response)
}

// UpdateTokenQuotaRequest defines the request structure for updating token quota
type UpdateTokenQuotaRequest struct {
	TokenID     int `json:"token_id"`     // Token ID to update
	RemainQuota int `json:"remain_quota"` // New quota amount
}

// UpdateTokenQuotaResponse defines the response structure
type UpdateTokenQuotaResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UpdateTokenQuota updates the remaining quota for a specific token
func UpdateTokenQuota(c *gin.Context) {
	var req UpdateTokenQuotaRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.TokenID <= 0 {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid token_id: must be greater than 0",
		})
		return
	}

	if req.RemainQuota < 0 {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid remain_quota: must be >= 0",
		})
		return
	}

	// Get token from database
	token, err := model.GetTokenById(req.TokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Token not found: " + err.Error(),
		})
		return
	}

	// Update quota
	token.RemainQuota = req.RemainQuota
	err = token.Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Failed to update token quota: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UpdateTokenQuotaResponse{
		Success: true,
		Message: "Token quota updated successfully",
	})
}

// UpdateTokenQuotaByKeyRequest defines the request structure for updating token quota by API key
type UpdateTokenQuotaByKeyRequest struct {
	APIKey      string `json:"api_key"`      // API key to identify the token
	RemainQuota int    `json:"remain_quota"` // New quota amount
}

// UpdateTokenQuotaByKey updates the remaining quota for a token using its API key
func UpdateTokenQuotaByKey(c *gin.Context) {
	var req UpdateTokenQuotaByKeyRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "API key is required",
		})
		return
	}

	if req.RemainQuota < 0 {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid remain_quota: must be >= 0",
		})
		return
	}

	// Get token from database by key
	token, err := model.GetTokenByKey(req.APIKey, true)
	if err != nil {
		c.JSON(http.StatusNotFound, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Token not found: " + err.Error(),
		})
		return
	}

	// Update quota
	token.RemainQuota = req.RemainQuota
	err = token.Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Failed to update token quota: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UpdateTokenQuotaResponse{
		Success: true,
		Message: "Token quota updated successfully",
	})
}

// GetTokenInfoRequest defines the request structure for getting token info
type GetTokenInfoRequest struct {
	APIKey string `json:"api_key"` // API key to query
}

// GetTokenInfoResponse defines the response structure for token info
type GetTokenInfoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		TokenID     int    `json:"token_id"`
		Name        string `json:"name"`
		RemainQuota int    `json:"remain_quota"`
		UsedQuota   int    `json:"used_quota"`
		CreatedTime int64  `json:"created_time"`
		ExpiredTime int64  `json:"expired_time"`
		Group       string `json:"group"`
		Status      int    `json:"status"`
	} `json:"data,omitempty"`
}

// GetTokenInfo retrieves token information by API key
func GetTokenInfo(c *gin.Context) {
	var req GetTokenInfoRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, GetTokenInfoResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, GetTokenInfoResponse{
			Success: false,
			Message: "API key is required",
		})
		return
	}

	// Get token from database by key
	token, err := model.GetTokenByKey(req.APIKey, true)
	if err != nil {
		c.JSON(http.StatusNotFound, GetTokenInfoResponse{
			Success: false,
			Message: "Token not found: " + err.Error(),
		})
		return
	}

	// Return token information
	response := GetTokenInfoResponse{
		Success: true,
		Message: "Token information retrieved successfully",
	}
	response.Data.TokenID = token.Id
	response.Data.Name = token.Name
	response.Data.RemainQuota = token.RemainQuota
	response.Data.UsedQuota = token.UsedQuota
	response.Data.CreatedTime = token.CreatedTime
	response.Data.ExpiredTime = token.ExpiredTime
	response.Data.Group = token.Group
	response.Data.Status = token.Status

	c.JSON(http.StatusOK, response)
}

// AddTokenQuotaRequest defines the request structure for adding quota to token
type AddTokenQuotaRequest struct {
	APIKey    string `json:"api_key"`    // API key to identify the token
	AddQuota  int    `json:"add_quota"`  // Amount of quota to add
}

// AddTokenQuota adds quota to an existing token
func AddTokenQuota(c *gin.Context) {
	var req AddTokenQuotaRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "API key is required",
		})
		return
	}

	if req.AddQuota <= 0 {
		c.JSON(http.StatusBadRequest, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Invalid add_quota: must be greater than 0",
		})
		return
	}

	// Get token from database by key
	token, err := model.GetTokenByKey(req.APIKey, true)
	if err != nil {
		c.JSON(http.StatusNotFound, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Token not found: " + err.Error(),
		})
		return
	}

	// Add quota to existing quota
	token.RemainQuota += req.AddQuota
	err = token.Update()
	if err != nil {
		c.JSON(http.StatusInternalServerError, UpdateTokenQuotaResponse{
			Success: false,
			Message: "Failed to add token quota: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UpdateTokenQuotaResponse{
		Success: true,
		Message: "Token quota added successfully",
	})
} 