// Package model 定义系统所有数据结构
package model

import "time"

// ============================================================================
// 用户与认证
// ============================================================================

// User 系统用户
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // bcrypt哈希，不在JSON中暴露
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Role         string    `json:"role"` // admin/manufacturer/logistics/distributor/auditor/consumer
	BlockchainAddr string  `json:"blockchain_addr"` // 关联的区块链账户地址
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
	ExpireAt int64  `json:"expire_at"`
}

// ============================================================================
// 产品与溯源
// ============================================================================

// ProductRegisterRequest 产品注册请求（制造商上链）
type ProductRegisterRequest struct {
	ProductID    string `json:"product_id" binding:"required"`
	Brand        string `json:"brand" binding:"required"`
	Model        string `json:"model" binding:"required"`
	BatchNo      string `json:"batch_no" binding:"required"`
	MaterialHash string `json:"material_hash"`
	QCReportHash string `json:"qc_report_hash"`
	ProduceDate  string `json:"produce_date" binding:"required"`
}

// ProductResponse 产品响应
type ProductResponse struct {
	ProductID      string `json:"product_id"`
	Brand          string `json:"brand"`
	Model          string `json:"model"`
	BatchNo        string `json:"batch_no"`
	Manufacturer   string `json:"manufacturer"`
	ProduceDate    string `json:"produce_date"`
	MaterialHash   string `json:"material_hash"`
	QCReportHash   string `json:"qc_report_hash"`
	CurrentOwner   string `json:"current_owner"`
	CurrentStatus  string `json:"current_status"`
	CreatedAt      int64  `json:"created_at"`
}

// TransferRequest 所有权转移请求
type TransferRequest struct {
	ProductID    string `json:"product_id" binding:"required"`
	NewOwner     string `json:"new_owner" binding:"required"`
	TransferType string `json:"transfer_type" binding:"required"`
}

// ============================================================================
// 溯源记录
// ============================================================================

// TraceRecordRequest 溯源记录追加请求
type TraceRecordRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Action    string `json:"action" binding:"required"`
	Location  string `json:"location" binding:"required"`
	ExtraData string `json:"extra_data"`
	Signature string `json:"signature" binding:"required"`
}

// TraceRecordResponse 溯源记录响应
type TraceRecordResponse struct {
	RecordID  string `json:"record_id"`
	ProductID string `json:"product_id"`
	Operator  string `json:"operator"`
	Action    string `json:"action"`
	Location  string `json:"location"`
	Timestamp int64  `json:"timestamp"`
	ExtraData string `json:"extra_data"`
	Signature string `json:"signature"`
}

// TraceHistoryResponse 溯源历史响应
type TraceHistoryResponse struct {
	ProductID string                 `json:"product_id"`
	Count     int                    `json:"count"`
	Records   []TraceRecordResponse  `json:"records"`
}

// ============================================================================
// NFC验证
// ============================================================================

// NFCVerifyRequest NFC标签验证请求（消费者扫码）
type NFCVerifyRequest struct {
	TagUID     string `json:"tag_uid" binding:"required"`    // NFC芯片唯一UID
	SUNCode    string `json:"sun_code" binding:"required"`   // SUN动态认证码（CMAC）
	Counter    int64  `json:"counter" binding:"required"`    // 芯片内部计数器值
	ProductID  string `json:"product_id" binding:"required"` // 绑定的产品ID
}

// NFCVerifyResponse NFC验证响应
type NFCVerifyResponse struct {
	Authentic      bool              `json:"authentic"`
	ProductID      string            `json:"product_id"`
	Brand          string            `json:"brand"`
	Model          string            `json:"model"`
	NFCVerified    bool              `json:"nfc_verified"`    // NFC认证码是否验证通过
	ChainVerified  bool              `json:"chain_verified"`  // 区块链上是否存在该产品
	TraceHistory   *TraceHistoryResponse `json:"trace_history,omitempty"`
	Message        string            `json:"message"`
}

// ============================================================================
// 通用API响应
// ============================================================================

// APIResponse 统一API响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ============================================================================
// 数据统计
// ============================================================================

// StatsResponse 溯源统计响应
type StatsResponse struct {
	TotalProducts      int64            `json:"total_products"`
	TotalTraceRecords  int64            `json:"total_trace_records"`
	ProductsByBrand    map[string]int64 `json:"products_by_brand"`
	ProductsByStatus   map[string]int64 `json:"products_by_status"`
	RecentVerifications int64            `json:"recent_verifications"`
}

// AnomalyInfo 异常信息
type AnomalyInfo struct {
	Type        string `json:"type"`
	ProductID   string `json:"product_id"`
	Description string `json:"description"`
	DetectedAt  int64  `json:"detected_at"`
	Severity    string `json:"severity"` // high/medium/low
}
