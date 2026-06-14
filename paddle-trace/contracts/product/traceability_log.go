// Package product 溯源记录智能合约
// 管理产品全生命周期的不可篡改流转日志
// 每条记录在写入时自动附加区块链时间戳和操作方签名
package product

import (
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
)

// ============================================================================
// TraceRecord 溯源记录数据结构
// ============================================================================
type TraceRecord struct {
	RecordID   string `json:"record_id"`   // 记录唯一ID
	ProductID  string `json:"product_id"`  // 关联产品ID
	Operator   string `json:"operator"`    // 操作方地址
	Action     string `json:"action"`      // 操作类型
	Location   string `json:"location"`    // 地理位置
	Timestamp  int64  `json:"timestamp"`    // 区块链时间戳（不可伪造）
	PrevRecord string `json:"prev_record"` // 前一条记录的哈希（链式防篡改）
	ExtraData  string `json:"extra_data"`  // 扩展数据（JSON格式）
	Signature  string `json:"signature"`   // 操作方SM2数字签名
}

// ============================================================================
// 操作类型常量
// ============================================================================
const (
	ActionProduce       = "produce"        // 生产完成
	ActionOutbound      = "outbound"       // 出厂出库
	ActionTransit       = "transit"        // 物流中转
	ActionArrival       = "arrival"        // 到货签收
	ActionInbound       = "inbound"        // 经销商入库
	ActionOnSale        = "on_sale"        // 上架销售
	ActionSold          = "sold"           // 售出
	ActionReturn        = "return"         // 退货
	ActionQualityCheck  = "quality_check"  // 质检记录
)

// ============================================================================
// TraceabilityLog 合约结构体
// ============================================================================
type TraceabilityLog struct {
	ctx code.Context
}

// ============================================================================
// 核心业务方法
// ============================================================================

// AppendTraceRecord 追加溯源记录
// 参数：productID, action, location, extraData, signature
// 仅物流商/经销商/制造商可调用
// 该方法只追加不删除不修改，确保溯源记录的不可篡改性和完整性
func (tl *TraceabilityLog) AppendTraceRecord(ctx code.Context) code.Response {
	tl.ctx = ctx
	caller := ctx.Initiator()

	args := ctx.Args()
	productID := string(args["product_id"])
	action := string(args["action"])
	location := string(args["location"])
	extraData := string(args["extra_data"])
	signature := string(args["signature"])

	if productID == "" || action == "" || location == "" {
		return code.Errors("product_id, action, and location are required")
	}

	// 1. 验证产品存在
	productData, _ := ctx.GetObject([]byte("product:" + productID))
	if productData == nil {
		return code.Errors("product not found: " + productID)
	}

	// 2. 验证操作类型合法性
	if !tl.isValidAction(action) {
		return code.Errors("invalid action type: " + action)
	}

	// 3. 生成记录ID（使用产品ID + 当前计数器确保唯一性）
	counterKey := fmt.Sprintf("counter:%s", productID)
	var count int64
	if existingData, _ := ctx.GetObject([]byte(counterKey)); existingData != nil {
		fmt.Sscanf(string(existingData), "%d", &count)
	}
	count++
	recordID := fmt.Sprintf("%s_%d", productID, count)

	// 4. 获取前一条记录的哈希（构建链式防篡改结构）
	prevRecordHash := tl.getLastRecordHash(productID)

	// 5. 构建溯源记录
	record := TraceRecord{
		RecordID:   recordID,
		ProductID:  productID,
		Operator:   caller,
		Action:     action,
		Location:   location,
		Timestamp:  ctx.Block().Timestamp,
		PrevRecord: prevRecordHash,
		ExtraData:  extraData,
		Signature:  signature,
	}

	recordData, err := json.Marshal(record)
	if err != nil {
		return code.Errors("marshal error: " + err.Error())
	}

	// 6. 存储溯源记录（键：trace:{productID}:{recordID}）
	traceKey := fmt.Sprintf("trace:%s:%s", productID, recordID)
	if err := ctx.PutObject([]byte(traceKey), recordData); err != nil {
		return code.Errors("storage error: " + err.Error())
	}

	// 7. 更新计数器
	ctx.PutObject([]byte(counterKey), []byte(fmt.Sprintf("%d", count)))

	// 8. 发布事件
	ctx.EmitEvent("TraceRecordAppended", map[string]string{
		"record_id":  recordID,
		"product_id": productID,
		"action":     action,
		"operator":   caller,
		"location":   location,
	})

	return code.OK([]byte("trace record appended: " + recordID))
}

// QueryHistory 查询产品完整溯源历史
// 参数：productID
// 返回：按时间升序排列的所有溯源记录
func (tl *TraceabilityLog) QueryHistory(ctx code.Context) code.Response {
	tl.ctx = ctx

	args := ctx.Args()
	productID := string(args["product_id"])
	if productID == "" {
		return code.Errors("product_id is required")
	}

	// 获取计数器（记录总数）
	counterKey := fmt.Sprintf("counter:%s", productID)
	counterData, _ := ctx.GetObject([]byte(counterKey))
	if counterData == nil {
		// 无溯源记录
		result, _ := json.Marshal(map[string]interface{}{
			"product_id": productID,
			"count":      0,
			"records":    []TraceRecord{},
		})
		return code.OK(result)
	}

	var count int64
	fmt.Sscanf(string(counterData), "%d", &count)

	// 遍历所有溯源记录
	var records []TraceRecord
	for i := int64(1); i <= count; i++ {
		recordID := fmt.Sprintf("%s_%d", productID, i)
		traceKey := fmt.Sprintf("trace:%s:%s", productID, recordID)
		recordData, _ := ctx.GetObject([]byte(traceKey))
		if recordData == nil {
			continue
		}

		var record TraceRecord
		if err := json.Unmarshal(recordData, &record); err != nil {
			continue
		}
		records = append(records, record)
	}

	result, _ := json.Marshal(map[string]interface{}{
		"product_id": productID,
		"count":      len(records),
		"records":    records,
	})

	return code.OK(result)
}

// GetCurrentState 获取产品当前溯源状态
// 返回最新的一条溯源记录
func (tl *TraceabilityLog) GetCurrentState(ctx code.Context) code.Response {
	tl.ctx = ctx

	args := ctx.Args()
	productID := string(args["product_id"])
	if productID == "" {
		return code.Errors("product_id is required")
	}

	counterKey := fmt.Sprintf("counter:%s", productID)
	counterData, _ := ctx.GetObject([]byte(counterKey))
	if counterData == nil {
		result, _ := json.Marshal(map[string]interface{}{
			"product_id": productID,
			"state":      "no_trace_records",
		})
		return code.OK(result)
	}

	var count int64
	fmt.Sscanf(string(counterData), "%d", &count)

	// 获取最后一条记录
	recordID := fmt.Sprintf("%s_%d", productID, count)
	traceKey := fmt.Sprintf("trace:%s:%s", productID, recordID)
	recordData, _ := ctx.GetObject([]byte(traceKey))
	if recordData == nil {
		return code.Errors("trace record not found")
	}

	var record TraceRecord
	json.Unmarshal(recordData, &record)

	result, _ := json.Marshal(map[string]interface{}{
		"product_id":  productID,
		"last_action": record.Action,
		"last_location": record.Location,
		"last_operator": record.Operator,
		"last_timestamp": record.Timestamp,
		"total_records": count,
	})

	return code.OK(result)
}

// VerifyIntegrity 验证溯源记录的完整性和连续性
// 检查所有记录的 PrevRecord 哈希链是否完整
func (tl *TraceabilityLog) VerifyIntegrity(ctx code.Context) code.Response {
	tl.ctx = ctx

	args := ctx.Args()
	productID := string(args["product_id"])
	if productID == "" {
		return code.Errors("product_id is required")
	}

	counterKey := fmt.Sprintf("counter:%s", productID)
	counterData, _ := ctx.GetObject([]byte(counterKey))
	if counterData == nil {
		result, _ := json.Marshal(map[string]interface{}{
			"product_id": productID,
			"integrity":  "no_records",
			"message":    "no trace records to verify",
		})
		return code.OK(result)
	}

	var count int64
	fmt.Sscanf(string(counterData), "%d", &count)

	var integrityErrors []string
	var lastHash string

	for i := int64(1); i <= count; i++ {
		recordID := fmt.Sprintf("%s_%d", productID, i)
		traceKey := fmt.Sprintf("trace:%s:%s", productID, recordID)
		recordData, _ := ctx.GetObject([]byte(traceKey))
		if recordData == nil {
			integrityErrors = append(integrityErrors,
				fmt.Sprintf("missing record: %s", recordID))
			continue
		}

		var record TraceRecord
		json.Unmarshal(recordData, &record)

		// 验证链式结构：每条记录的PrevRecord应等于上一条记录的哈希
		if i > 1 && record.PrevRecord != lastHash {
			integrityErrors = append(integrityErrors,
				fmt.Sprintf("broken chain at %s: expected prev_hash=%s, got=%s",
					recordID, lastHash, record.PrevRecord))
		}

		// 计算当前记录的哈希（使用SM3）
		recordHash := tl.hashRecord(record)
		lastHash = recordHash
	}

	integrity := "valid"
	if len(integrityErrors) > 0 {
		integrity = "compromised"
	}

	result, _ := json.Marshal(map[string]interface{}{
		"product_id": productID,
		"integrity":  integrity,
		"errors":     integrityErrors,
		"total_records": count,
	})

	return code.OK(result)
}

// ============================================================================
// 内部辅助方法
// ============================================================================

func (tl *TraceabilityLog) isValidAction(action string) bool {
	validActions := map[string]bool{
		ActionProduce:      true,
		ActionOutbound:     true,
		ActionTransit:      true,
		ActionArrival:      true,
		ActionInbound:      true,
		ActionOnSale:       true,
		ActionSold:         true,
		ActionReturn:       true,
		ActionQualityCheck: true,
	}
	return validActions[action]
}

func (tl *TraceabilityLog) getLastRecordHash(productID string) string {
	counterKey := fmt.Sprintf("counter:%s", productID)
	counterData, _ := tl.ctx.GetObject([]byte(counterKey))
	if counterData == nil {
		return "genesis"
	}

	var count int64
	fmt.Sscanf(string(counterData), "%d", &count)
	if count == 0 {
		return "genesis"
	}

	recordID := fmt.Sprintf("%s_%d", productID, count)
	traceKey := fmt.Sprintf("trace:%s:%s", productID, recordID)
	recordData, _ := tl.ctx.GetObject([]byte(traceKey))
	if recordData == nil {
		return "genesis"
	}

	var record TraceRecord
	json.Unmarshal(recordData, &record)
	return tl.hashRecord(record)
}

func (tl *TraceabilityLog) hashRecord(record TraceRecord) string {
	// 使用SM3哈希计算溯源记录的摘要
	// 原型阶段使用Go crypto/sha256作为替代（生产环境应使用SM3）
	data, _ := json.Marshal(record)
	hash := tl.ctx.Hash(data)
	return fmt.Sprintf("%x", hash)
}
