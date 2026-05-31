package admin

import (
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// BatchRefreshUsage 批量刷新账号用量（强制从上游拉取最新用量）。
// POST /api/v1/admin/accounts/batch-refresh-usage
//
// 与 BatchRefresh（刷新 token）对应，这里强制 GetUsage(force=true)。
// 仅对支持 usage 查询的账号有效；不支持的账号会返回失败明细但不影响其他账号。
func (h *AccountHandler) BatchRefreshUsage(c *gin.Context) {
	var req struct {
		AccountIDs []int64 `json:"account_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.AccountIDs) == 0 {
		response.BadRequest(c, "account_ids is required")
		return
	}

	ctx := c.Request.Context()

	const maxConcurrency = 10
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	var mu sync.Mutex
	var successCount, failedCount int
	var errors []gin.H

	for _, id := range req.AccountIDs {
		accountID := id
		g.Go(func() error {
			_, err := h.accountUsageService.GetUsage(gctx, accountID, true)
			mu.Lock()
			if err != nil {
				failedCount++
				errors = append(errors, gin.H{
					"account_id": accountID,
					"error":      err.Error(),
				})
			} else {
				successCount++
			}
			mu.Unlock()
			// 始终 return nil，避免 errgroup cancel 其他并发任务
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"total":   len(req.AccountIDs),
		"success": successCount,
		"failed":  failedCount,
		"errors":  errors,
	})
}
