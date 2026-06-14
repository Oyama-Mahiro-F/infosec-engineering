// Package product 产品注册智能合约
// 负责产品数字身份的全生命周期管理：注册、所有权转移、真伪验证
// 所有写操作均通过 AccessControl 合约进行权限校验
package product

import (
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
)

// ============================================================================
// 数据结构定义
// ============================================================================

// Product 产品数据结构
type Product struct {
	ProductID      string `json:"product_id"`       // 产品唯一标识（与NFC芯片UID通过SM3绑定）
	Brand          string `json:"brand"`            // 品牌名称
	Model          string `json:"model"`            // 产品型号
	BatchNo        string `json:"batch_no"`         // 生产批次号
	Manufacturer   string `json:"manufacturer"`     // 制造商区块链账户地址
	ProduceDate    string `json:"produce_date"`     // 生产日期 (YYYY-MM-DD)
	MaterialHash   string `json:"material_hash"`    // 原材料批次信息的SM3哈希值
	QCReportHash   string `json:"qc_report_hash"`   // 质检报告文件的IPFS/哈希地址
	CurrentOwner   string `json:"current_owner"`    // 当前持有者地址
	CurrentStatus  string `json:"current_status"`   // 产品状态: produced/in_transit/in_stock/sold
	CreatedAt      int64  `json:"created_at"`       // 链上注册时间戳
	UpdatedAt      int64  `json:"updated_at"`       // 最后更新区块时间
}

// OwnershipRecord 所有权变更记录
type OwnershipRecord struct {
	ProductID    string `json:"product_id"`
	FromOwner    string `json:"from_owner"`
	ToOwner      string `json:"to_owner"`
	TransferType string `json:"transfer_type"`
	TransferredAt int64 `json:"transferred_at"`
}

// ============================================================================
// ProductRegistry 合约结构体
// ============================================================================
type ProductRegistry struct {
	ctx code.Context
}

// ============================================================================
// 核心业务方法
// ============================================================================

// RegisterProduct 注册新产品（仅制造商可调用）
// 参数：productID, brand, model, batchNo, materialHash, qcReportHash
func (pr *ProductRegistry) RegisterProduct(ctx code.Context) code.Response {
	pr.ctx = ctx
	caller := ctx.Initiator()

	// 1. 权限校验：调用 AccessControl 合约验证调用者是否拥有 manufacturer_role
	if !pr.checkRole(caller, "manufacturer_role") {
		return code.Errors("permission denied: caller is not a registered manufacturer")
	}

	// 2. 参数解析
	args := ctx.Args()
	productID := string(args["product_id"])
	brand := string(args["brand"])
	model := string(args["model"])
	batchNo := string(args["batch_no"])
	materialHash := string(args["material_hash"])
	qcReportHash := string(args["qc_report_hash"])
	produceDate := string(args["produce_date"])

	if productID == "" || brand == "" || model == "" {
		return code.Errors("product_id, brand, and model are required fields")
	}

	// 3. 检查产品是否已存在（防止重复注册）
	existing, _ := ctx.GetObject([]byte("product:" + productID))
	if existing != nil {
		return code.Errors("product already registered: " + productID)
	}

	// 4. 构建产品对象
	product := Product{
		ProductID:     productID,
		Brand:         brand,
		Model:         model,
		BatchNo:       batchNo,
		Manufacturer:  caller,
		ProduceDate:   produceDate,
		MaterialHash:  materialHash,
		QCReportHash:  qcReportHash,
		CurrentOwner:  caller,
		CurrentStatus: "produced",
		CreatedAt:     ctx.Block().Timestamp,
		UpdatedAt:     ctx.Block().Timestamp,
	}

	data, err := json.Marshal(product)
	if err != nil {
		return code.Errors("marshal error: " + err.Error())
	}

	// 5. 存储产品数据
	if err := ctx.PutObject([]byte("product:"+productID), data); err != nil {
		return code.Errors("storage error: " + err.Error())
	}

	// 6. 记录初始所有权
	ownershipKey := fmt.Sprintf("ownership:%s:0", productID)
	initRecord := OwnershipRecord{
		ProductID:    productID,
		FromOwner:    "genesis",
		ToOwner:      caller,
		TransferType: "initial_registration",
		TransferredAt: ctx.Block().Timestamp,
	}
	ownershipData, _ := json.Marshal(initRecord)
	ctx.PutObject([]byte(ownershipKey), ownershipData)

	// 7. 发布事件
	ctx.EmitEvent("ProductRegistered", map[string]string{
		"product_id":    productID,
		"manufacturer":  caller,
		"brand":         brand,
		"model":         model,
	})

	return code.OK([]byte("product registered: " + productID))
}

// TransferOwnership 转移产品所有权
// 参数：productID, newOwner, transferType
// 仅当前持有者可调用，转移后产品所有权记录不可逆
func (pr *ProductRegistry) TransferOwnership(ctx code.Context) code.Response {
	pr.ctx = ctx
	caller := ctx.Initiator()

	args := ctx.Args()
	productID := string(args["product_id"])
	newOwner := string(args["new_owner"])
	transferType := string(args["transfer_type"])

	if productID == "" || newOwner == "" || transferType == "" {
		return code.Errors("product_id, new_owner, and transfer_type are required")
	}

	// 1. 获取产品数据
	data, _ := ctx.GetObject([]byte("product:" + productID))
	if data == nil {
		return code.Errors("product not found: " + productID)
	}

	var product Product
	if err := json.Unmarshal(data, &product); err != nil {
		return code.Errors("unmarshal error: " + err.Error())
	}

	// 2. 校验调用者是否为当前持有者
	if product.CurrentOwner != caller {
		return code.Errors("transfer denied: caller is not the current owner")
	}

	// 3. 校验转移类型合法性
	validTypes := map[string]bool{
		"manufacturer_to_logistics":    true,
		"logistics_to_distributor":     true,
		"distributor_to_consumer":      true,
		"consumer_to_consumer":         true,
	}
	if !validTypes[transferType] {
		return code.Errors("invalid transfer type: " + transferType)
	}

	// 4. 更新产品所有权和状态
	oldOwner := product.CurrentOwner
	product.CurrentOwner = newOwner
	product.UpdatedAt = ctx.Block().Timestamp

	// 根据转移类型更新状态
	switch transferType {
	case "manufacturer_to_logistics":
		product.CurrentStatus = "in_transit"
	case "logistics_to_distributor":
		product.CurrentStatus = "in_stock"
	case "distributor_to_consumer", "consumer_to_consumer":
		product.CurrentStatus = "sold"
	}

	// 5. 存储更新后的产品数据
	updatedData, _ := json.Marshal(product)
	if err := ctx.PutObject([]byte("product:"+productID), updatedData); err != nil {
		return code.Errors("update error: " + err.Error())
	}

	// 6. 记录所有权变更历史
	ownershipIndex := fmt.Sprintf("%d", ctx.Block().Timestamp)
	ownershipKey := fmt.Sprintf("ownership:%s:%s", productID, ownershipIndex)
	record := OwnershipRecord{
		ProductID:    productID,
		FromOwner:    oldOwner,
		ToOwner:      newOwner,
		TransferType: transferType,
		TransferredAt: ctx.Block().Timestamp,
	}
	recordData, _ := json.Marshal(record)
	ctx.PutObject([]byte(ownershipKey), recordData)

	// 7. 发布事件
	ctx.EmitEvent("OwnershipTransferred", map[string]string{
		"product_id":    productID,
		"from":          oldOwner,
		"to":            newOwner,
		"transfer_type": transferType,
	})

	return code.OK([]byte(fmt.Sprintf("ownership transferred: %s -> %s", oldOwner, newOwner)))
}

// VerifyAuthenticity 验证产品真伪（公开查询方法，消费者可调用）
// 参数：productID
// 返回：产品信息（若存在）或错误（若未注册/不存在）
func (pr *ProductRegistry) VerifyAuthenticity(ctx code.Context) code.Response {
	pr.ctx = ctx

	args := ctx.Args()
	productID := string(args["product_id"])
	if productID == "" {
		return code.Errors("product_id is required")
	}

	data, _ := ctx.GetObject([]byte("product:" + productID))
	if data == nil {
		// 产品未在区块链注册 → 疑为假冒产品
		result, _ := json.Marshal(map[string]interface{}{
			"authentic":   false,
			"product_id":  productID,
			"reason":      "product not registered on blockchain",
			"recommendation": "This product is NOT registered on the blockchain. It may be a counterfeit.",
		})
		return code.OK(result)
	}

	var product Product
	if err := json.Unmarshal(data, &product); err != nil {
		return code.Errors("internal error: " + err.Error())
	}

	result, _ := json.Marshal(map[string]interface{}{
		"authentic":      true,
		"product_id":     product.ProductID,
		"brand":          product.Brand,
		"model":          product.Model,
		"manufacturer":   product.Manufacturer,
		"produce_date":   product.ProduceDate,
		"current_status": product.CurrentStatus,
		"current_owner":  product.CurrentOwner,
		"registered_at":  product.CreatedAt,
		"message":        "✅ Authentic product verified on blockchain",
	})

	return code.OK(result)
}

// GetProduct 查询产品详细信息（公开查询方法）
func (pr *ProductRegistry) GetProduct(ctx code.Context) code.Response {
	pr.ctx = ctx

	args := ctx.Args()
	productID := string(args["product_id"])
	if productID == "" {
		return code.Errors("product_id is required")
	}

	data, _ := ctx.GetObject([]byte("product:" + productID))
	if data == nil {
		return code.Errors("product not found: " + productID)
	}

	return code.OK(data)
}

// ListProductsByManufacturer 查询指定制造商的所有产品
func (pr *ProductRegistry) ListProductsByManufacturer(ctx code.Context) code.Response {
	pr.ctx = ctx

	args := ctx.Args()
	manufacturer := string(args["manufacturer"])
	if manufacturer == "" {
		return code.Errors("manufacturer address is required")
	}

	// 使用迭代器遍历所有产品（生产环境中需考虑分页）
	iter := ctx.NewIterator("product:", "")
	defer iter.Close()

	var products []Product
	for iter.Next() {
		var product Product
		if err := json.Unmarshal(iter.Value(), &product); err != nil {
			continue
		}
		if product.Manufacturer == manufacturer {
			products = append(products, product)
		}
	}

	result, _ := json.Marshal(map[string]interface{}{
		"manufacturer": manufacturer,
		"count":        len(products),
		"products":     products,
	})

	return code.OK(result)
}

// ============================================================================
// 内部辅助方法
// ============================================================================

// checkRole 通过跨合约调用验证调用者角色
func (pr *ProductRegistry) checkRole(addr, role string) bool {
	// 调用 AccessControl 合约的 HasRole 方法
	callArgs := map[string][]byte{
		"address": []byte(addr),
		"role":    []byte(role),
	}

	resp, err := pr.ctx.Call("access_control", "HasRole", callArgs)
	if err != nil {
		return false
	}

	if resp.Status >= 400 {
		return false
	}

	// 解析返回结果
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return false
	}

	hasRole, ok := result["has_role"].(bool)
	return ok && hasRole
}
