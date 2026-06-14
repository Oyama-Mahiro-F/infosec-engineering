// Package handler HTTP请求处理器
// 所有API端点的处理逻辑，负责请求参数校验、调用Service层、返回响应
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fansicheng/paddle-trace/internal/model"
	"github.com/fansicheng/paddle-trace/internal/service"
)

// ============================================================================
// Handler 依赖注入容器
// ============================================================================

// Handler 聚合所有Service依赖
type Handler struct {
	Blockchain *service.BlockchainService
	NFC        *service.NFCService
}

// NewHandler 创建Handler实例
func NewHandler(bs *service.BlockchainService, ns *service.NFCService) *Handler {
	return &Handler{
		Blockchain: bs,
		NFC:        ns,
	}
}

// ============================================================================
// 产品相关处理器
// ============================================================================

// RegisterProduct POST /api/v1/products
// 制造商注册新产品上链
func (h *Handler) RegisterProduct(c *gin.Context) {
	var req model.ProductRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:  400,
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	// 调用区块链服务注册产品
	contractArgs := map[string]string{
		"product_id":     req.ProductID,
		"brand":          req.Brand,
		"model":          req.Model,
		"batch_no":       req.BatchNo,
		"material_hash":  req.MaterialHash,
		"qc_report_hash": req.QCReportHash,
		"produce_date":   req.ProduceDate,
	}

	txID, err := h.Blockchain.RegisterProduct(c.Request.Context(), contractArgs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Code:  500,
			Error: "blockchain error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Code:    201,
		Message: "product registered successfully",
		Data: gin.H{
			"product_id": req.ProductID,
			"tx_id":      txID,
		},
	})
}

// GetProduct GET /api/v1/products/:id
// 查询产品溯源信息（消费者公开查询）
func (h *Handler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:  400,
			Error: "product_id is required",
		})
		return
	}

	// 查询区块链上的产品信息
	product, err := h.Blockchain.QueryProduct(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Code:  404,
			Error: "product not found: " + err.Error(),
		})
		return
	}

	// 同时查询溯源历史
	history, _ := h.Blockchain.QueryTraceHistory(c.Request.Context(), productID)

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    200,
		Message: "success",
		Data: gin.H{
			"product": product,
			"history": history,
		},
	})
}

// TransferProduct POST /api/v1/products/:id/transfer
// 转移产品所有权
func (h *Handler) TransferProduct(c *gin.Context) {
	var req model.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:  400,
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	txID, err := h.Blockchain.TransferOwnership(
		c.Request.Context(),
		req.ProductID,
		req.NewOwner,
		req.TransferType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Code:  500,
			Error: "transfer failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    200,
		Message: "ownership transferred",
		Data:    gin.H{"tx_id": txID},
	})
}

// AppendTraceRecord POST /api/v1/products/:id/trace
// 追加溯源记录
func (h *Handler) AppendTraceRecord(c *gin.Context) {
	var req model.TraceRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:  400,
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	contractArgs := map[string]string{
		"product_id": req.ProductID,
		"action":     req.Action,
		"location":   req.Location,
		"extra_data": req.ExtraData,
		"signature":  req.Signature,
	}

	txID, err := h.Blockchain.AppendTraceRecord(c.Request.Context(), contractArgs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Code:  500,
			Error: "blockchain error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Code:    201,
		Message: "trace record appended",
		Data:    gin.H{"tx_id": txID},
	})
}

// ============================================================================
// NFC验证处理器
// ============================================================================

// VerifyNFC POST /api/v1/products/verify-nfc
// 消费者NFC扫码真伪验证入口
// 此接口整合了NFC芯片验证和区块链真伪查询两个步骤
func (h *Handler) VerifyNFC(c *gin.Context) {
	var req model.NFCVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Code:  400,
			Error: "invalid request: " + err.Error(),
		})
		return
	}

	response := model.NFCVerifyResponse{
		ProductID: req.ProductID,
	}

	// 步骤1：验证NFC芯片SUN动态认证码
	nfcOK, nfcMsg, err := h.NFC.VerifySUN(req.TagUID, req.SUNCode, req.Counter)
	if err != nil {
		response.NFCVerified = false
		response.Authentic = false
		response.Message = "NFC verification failed: " + nfcMsg
		c.JSON(http.StatusOK, model.APIResponse{
			Code:    200,
			Message: "verification complete (NFC failed)",
			Data:    response,
		})
		return
	}
	response.NFCVerified = nfcOK

	// 步骤2：查询区块链验证产品是否存在
	chainResult, err := h.Blockchain.VerifyAuthenticity(c.Request.Context(), req.ProductID)
	if err != nil {
		response.ChainVerified = false
		response.Authentic = false
		response.Message = "blockchain verification failed: " + err.Error()
		c.JSON(http.StatusOK, model.APIResponse{
			Code:    200,
			Message: "verification complete (chain query failed)",
			Data:    response,
		})
		return
	}

	authentic, _ := chainResult["authentic"].(bool)
	response.ChainVerified = authentic

	// 步骤3：综合判定
	if nfcOK && authentic {
		response.Authentic = true
		response.Brand, _ = chainResult["brand"].(string)
		response.Model, _ = chainResult["model"].(string)
		response.Message = "✅ 验证通过：该球拍为正品，NFC芯片认证通过，区块链溯源记录完整"

		// 附加完整的溯源历史
		history, _ := h.Blockchain.QueryTraceHistory(c.Request.Context(), req.ProductID)
		if history != nil {
			// 转换history为TraceHistoryResponse
			traceHistory := &model.TraceHistoryResponse{
				ProductID: req.ProductID,
			}
			if count, ok := history["count"].(float64); ok {
				traceHistory.Count = int(count)
			}
			response.TraceHistory = traceHistory
		}
	} else if !nfcOK {
		response.Authentic = false
		response.Message = "❌ NFC标签验证失败：该标签可能为克隆品或已被篡改"
	} else {
		response.Authentic = false
		response.Message = "❌ 产品未在区块链注册：该球拍疑为假冒产品"
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Code:    200,
		Message: "verification complete",
		Data:    response,
	})
}

// ============================================================================
// 健康检查
// ============================================================================

// HealthCheck GET /api/v1/health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "paddle-trace-api",
		"version": "1.0.0",
	})
}
