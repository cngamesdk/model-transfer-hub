package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cngamesdk/model-transfer-hub/global"
	"github.com/cngamesdk/model-transfer-hub/model"
	"github.com/gin-gonic/gin"
)

// TokenManageHandler Token管理处理器
type TokenManageHandler struct{}

// TokenManageCreateReq 创建Token请求
type TokenManageCreateReq struct {
	Token         string     `json:"token" binding:"required"`
	Name          string     `json:"name" binding:"required"`
	Type          int8       `json:"type"`
	TokenLimit    int64      `json:"token_limit"`
	RequestLimit  int        `json:"request_limit"`
	ExpireAt      *time.Time `json:"expire_at"`
	AllowedModels string     `json:"allowed_models"`
	IPWhitelist   string     `json:"ip_whitelist"`
	Creator       string     `json:"creator"`
}

// TokenManageUpdateReq 更新Token请求
type TokenManageUpdateReq struct {
	Token         string     `json:"token" binding:"required"`
	Name          *string    `json:"name"`
	Type          *int8      `json:"type"`
	TokenLimit    *int64     `json:"token_limit"`
	RequestLimit  *int       `json:"request_limit"`
	ExpireAt      *time.Time `json:"expire_at"`
	Status        *int8      `json:"status"`
	AllowedModels *string    `json:"allowed_models"`
	IPWhitelist   *string    `json:"ip_whitelist"`
}

// TokenManageDetailReq 获取Token详情请求
type TokenManageDetailReq struct {
	Token string `json:"token" binding:"required"`
}

// TokenManageUsageReq 获取Token使用记录请求
type TokenManageUsageReq struct {
	Token     string `json:"token" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

// TokenManageUsageSummary 使用记录汇总
type TokenManageUsageSummary struct {
	TokenName   string `json:"token_name"`
	TotalReqs   int64  `json:"total_requests"`
	TotalInput  int64  `json:"total_input_tokens"`
	TotalOutput int64  `json:"total_output_tokens"`
	TotalTokens int64  `json:"total_tokens"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// Create 创建Token
func (h *TokenManageHandler) Create(c *gin.Context) {
	var req TokenManageCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: fmt.Sprintf("请求参数错误: %v", err),
				Type:    "invalid_request",
			},
		})
		return
	}

	// 检查Token是否已存在
	var existing model.AiToken
	if err := global.MTH_DB.Where("token = ?", req.Token).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "Token已存在",
				Type:    "token_already_exists",
			},
		})
		return
	}

	// 构建新Token
	aiToken := model.AiToken{
		Token:         req.Token,
		Name:          req.Name,
		Type:          req.Type,
		TokenLimit:    req.TokenLimit,
		UsedTokens:    0,
		RequestLimit:  req.RequestLimit,
		ExpireAt:      req.ExpireAt,
		Status:        1, // 默认启用
		AllowedModels: req.AllowedModels,
		IPWhitelist:   req.IPWhitelist,
		Creator:       req.Creator,
	}
	if aiToken.Type == 0 {
		aiToken.Type = 1 // 默认企业类型
	}
	if aiToken.RequestLimit == 0 {
		aiToken.RequestLimit = 60 // 默认RPM
	}

	if err := global.MTH_DB.Create(&aiToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: fmt.Sprintf("创建Token失败: %v", err),
				Type:    "api_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data": gin.H{
			"id": aiToken.ID,
		},
	})
}

// Update 更新Token
func (h *TokenManageHandler) Update(c *gin.Context) {
	var req TokenManageUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: fmt.Sprintf("请求参数错误: %v", err),
				Type:    "invalid_request",
			},
		})
		return
	}

	// 检查Token是否存在
	var existing model.AiToken
	if err := global.MTH_DB.Where("token = ?", req.Token).First(&existing).Error; err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "Token不存在",
				Type:    "token_not_found",
			},
		})
		return
	}

	// 构建更新字段
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.TokenLimit != nil {
		updates["token_limit"] = *req.TokenLimit
	}
	if req.RequestLimit != nil {
		updates["request_limit"] = *req.RequestLimit
	}
	if req.ExpireAt != nil {
		updates["expire_at"] = req.ExpireAt
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.AllowedModels != nil {
		updates["allowed_models"] = *req.AllowedModels
	}
	if req.IPWhitelist != nil {
		updates["ip_whitelist"] = *req.IPWhitelist
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "没有需要更新的字段",
				Type:    "invalid_request",
			},
		})
		return
	}

	if err := global.MTH_DB.Model(&model.AiToken{}).Where("token = ?", req.Token).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: fmt.Sprintf("更新Token失败: %v", err),
				Type:    "api_error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

// Detail 获取Token详情
func (h *TokenManageHandler) Detail(c *gin.Context) {
	var req TokenManageDetailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: fmt.Sprintf("请求参数错误: %v", err),
				Type:    "invalid_request",
			},
		})
		return
	}

	var aiToken model.AiToken
	if err := global.MTH_DB.Where("token = ?", req.Token).First(&aiToken).Error; err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "Token不存在",
				Type:    "token_not_found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": aiToken,
	})
}

// Usage 获取Token使用记录（分页+汇总）
func (h *TokenManageHandler) Usage(c *gin.Context) {
	var req TokenManageUsageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: fmt.Sprintf("请求参数错误: %v", err),
				Type:    "invalid_request",
			},
		})
		return
	}

	// 分页默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 解析时间范围
	startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "start_time格式错误，请使用 2006-01-02 15:04:05 格式",
				Type:    "invalid_request",
			},
		})
		return
	}
	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "end_time格式错误，请使用 2006-01-02 15:04:05 格式",
				Type:    "invalid_request",
			},
		})
		return
	}

	// 查询Token信息
	var aiToken model.AiToken
	if err := global.MTH_DB.Where("token = ?", req.Token).First(&aiToken).Error; err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error: model.ErrorDetail{
				Message: "Token不存在",
				Type:    "token_not_found",
			},
		})
		return
	}

	db := global.MTH_DB.Model(&model.AiUsageLog{}).
		Where("token_id = ? AND request_time BETWEEN ? AND ?", aiToken.ID, startTime, endTime)

	// 查询总数
	var total int64
	db.Count(&total)

	// 查询分页记录（时间降序）
	var records []model.AiUsageLog
	offset := (req.Page - 1) * req.PageSize
	db.Order("request_time DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&records)

	// 汇总使用记录
	var summaryResult struct {
		TotalReqs   int64
		TotalInput  int64
		TotalOutput int64
		TotalTokens int64
	}
	db.Select("COUNT(*) as total_reqs, COALESCE(SUM(request_tokens), 0) as total_input, COALESCE(SUM(response_tokens), 0) as total_output, COALESCE(SUM(total_tokens), 0) as total_tokens").
		Scan(&summaryResult)

	summary := TokenManageUsageSummary{
		TokenName:   aiToken.Name,
		TotalReqs:   summaryResult.TotalReqs,
		TotalInput:  summaryResult.TotalInput,
		TotalOutput: summaryResult.TotalOutput,
		TotalTokens: summaryResult.TotalTokens,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"records": records,
			"summary": summary,
			"pagination": gin.H{
				"page":      req.Page,
				"page_size": req.PageSize,
				"total":     total,
			},
		},
	})
}
