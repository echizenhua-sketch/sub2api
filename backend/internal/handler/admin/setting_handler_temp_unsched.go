package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GetAccountTempUnschedulableRules 读取全局账号错误处理规则。
// GET /api/v1/admin/settings/account-temp-unschedulable-rules
func (h *SettingHandler) GetAccountTempUnschedulableRules(c *gin.Context) {
	cfg := h.settingService.GetGlobalTempUnschedulableRules(c.Request.Context())
	if cfg.Rules == nil {
		cfg.Rules = []service.TempUnschedulableRule{}
	}
	response.Success(c, gin.H{
		"enabled": cfg.Enabled,
		"rules":   cfg.Rules,
	})
}

// UpdateAccountTempUnschedulableRules 写入全局账号错误处理规则。
// PUT /api/v1/admin/settings/account-temp-unschedulable-rules
func (h *SettingHandler) UpdateAccountTempUnschedulableRules(c *gin.Context) {
	var req struct {
		Enabled bool                            `json:"enabled"`
		Rules   []service.TempUnschedulableRule `json:"rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	// 基本校验：错误码合法、时长非负。
	for i := range req.Rules {
		if req.Rules[i].ErrorCode < 100 || req.Rules[i].ErrorCode > 599 {
			response.BadRequest(c, "error_code must be between 100 and 599")
			return
		}
		if req.Rules[i].DurationMinutes < 0 {
			response.BadRequest(c, "duration_minutes must be >= 0")
			return
		}
	}
	cfg := service.GlobalTempUnschedulableConfig{
		Enabled: req.Enabled,
		Rules:   req.Rules,
	}
	if err := h.settingService.SetGlobalTempUnschedulableRules(c.Request.Context(), cfg); err != nil {
		response.InternalError(c, "Failed to save rules: "+err.Error())
		return
	}
	response.Success(c, gin.H{"message": "saved"})
}
