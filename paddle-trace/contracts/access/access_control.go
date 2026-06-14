// Package access 访问控制智能合约
// 实现基于角色的权限管理 (RBAC)，为 ProductRegistry 和 TraceabilityLog 合约提供权限校验
package access

import (
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
)

// ============================================================================
// 角色常量定义
// ============================================================================
const (
	RoleAdmin        = "admin_role"        // 系统管理员：可授予/撤销任意角色
	RoleManufacturer = "manufacturer_role" // 制造商：可注册新产品
	RoleLogistics    = "logistics_role"    // 物流商：可追加物流溯源记录
	RoleDistributor  = "distributor_role"  // 经销商：可记录销售状态
	RoleAuditor      = "auditor_role"      // 审计方：仅可查询（只读）
	RoleConsumer     = "consumer_role"     // 消费者：仅可查询自己的产品
)

// ============================================================================
// AccessControl 合约结构体
// ============================================================================
type AccessControl struct {
	ctx code.Context
}

// ============================================================================
// 合约内部数据结构
// ============================================================================

// RoleAssignment 角色分配记录
type RoleAssignment struct {
	Address   string `json:"address"`    // 被授权的区块链账户地址
	Role      string `json:"role"`       // 角色名称
	GrantedBy string `json:"granted_by"` // 授权人地址
	GrantedAt int64  `json:"granted_at"` // 授权时间戳
}

// ============================================================================
// 合约初始化方法
// ============================================================================

// Initialize 合约初始化，设置管理员账户
// 参数：adminAddr - 初始管理员地址（区块链账户公钥哈希）
func (ac *AccessControl) Initialize(ctx code.Context) code.Response {
	ac.ctx = ctx

	// 获取合约初始化参数
	args := ctx.Args()
	adminAddr := string(args["admin_addr"])
	if adminAddr == "" {
		return code.Errors("admin_addr is required for initialization")
	}

	// 检查是否已经初始化（防止重复初始化）
	existing, _ := ctx.GetObject([]byte("initialized"))
	if existing != nil {
		return code.Errors("contract already initialized")
	}

	// 创建管理员角色分配
	assignment := RoleAssignment{
		Address:   adminAddr,
		Role:      RoleAdmin,
		GrantedBy: "genesis",
		GrantedAt: ctx.Block().Timestamp,
	}

	data, err := json.Marshal(assignment)
	if err != nil {
		return code.Errors("failed to marshal role assignment: " + err.Error())
	}

	// 存储角色分配记录（键：role_assignment:{角色}:{地址}）
	key := fmt.Sprintf("role_assignment:%s:%s", RoleAdmin, adminAddr)
	if err := ctx.PutObject([]byte(key), data); err != nil {
		return code.Errors("failed to store admin role: " + err.Error())
	}

	// 标记合约已初始化
	if err := ctx.PutObject([]byte("initialized"), []byte("true")); err != nil {
		return code.Errors("failed to set initialized flag: " + err.Error())
	}

	return code.OK([]byte("AccessControl contract initialized. Admin: " + adminAddr))
}

// ============================================================================
// 核心权限管理方法
// ============================================================================

// GrantRole 授予角色（仅管理员可调用）
// 参数：address - 被授权地址, role - 角色名称
func (ac *AccessControl) GrantRole(ctx code.Context) code.Response {
	ac.ctx = ctx
	caller := ctx.Initiator()

	// 校验调用者是否为管理员
	if !ac.hasRole(caller, RoleAdmin) {
		return code.Errors("permission denied: caller is not admin")
	}

	args := ctx.Args()
	addr := string(args["address"])
	role := string(args["role"])

	if addr == "" || role == "" {
		return code.Errors("address and role are required")
	}

	// 校验角色名称是否合法
	if !ac.isValidRole(role) {
		return code.Errors("invalid role: " + role)
	}

	// 检查是否已有该角色（幂等性保护）
	if ac.hasRole(addr, role) {
		return code.OK([]byte("role already assigned"))
	}

	assignment := RoleAssignment{
		Address:   addr,
		Role:      role,
		GrantedBy: caller,
		GrantedAt: ctx.Block().Timestamp,
	}

	data, err := json.Marshal(assignment)
	if err != nil {
		return code.Errors("marshal error: " + err.Error())
	}

	key := fmt.Sprintf("role_assignment:%s:%s", role, addr)
	if err := ctx.PutObject([]byte(key), data); err != nil {
		return code.Errors("storage error: " + err.Error())
	}

	return code.OK([]byte(fmt.Sprintf("role %s granted to %s by %s", role, addr, caller)))
}

// RevokeRole 撤销角色（仅管理员可调用）
// 参数：address - 被撤销地址, role - 角色名称
func (ac *AccessControl) RevokeRole(ctx code.Context) code.Response {
	ac.ctx = ctx
	caller := ctx.Initiator()

	// 校验调用者是否为管理员
	if !ac.hasRole(caller, RoleAdmin) {
		return code.Errors("permission denied: caller is not admin")
	}

	args := ctx.Args()
	addr := string(args["address"])
	role := string(args["role"])

	if addr == "" || role == "" {
		return code.Errors("address and role are required")
	}

	// 不允许撤销创始管理员的角色
	if addr == caller && role == RoleAdmin {
		return code.Errors("cannot revoke own admin role")
	}

	key := fmt.Sprintf("role_assignment:%s:%s", role, addr)

	// 检查角色是否存在
	existing, _ := ctx.GetObject([]byte(key))
	if existing == nil {
		return code.Errors("role assignment not found")
	}

	if err := ctx.DeleteObject([]byte(key)); err != nil {
		return code.Errors("storage error: " + err.Error())
	}

	return code.OK([]byte(fmt.Sprintf("role %s revoked from %s by %s", role, addr, caller)))
}

// HasRole 查询指定地址是否拥有指定角色
// 参数：address - 查询地址, role - 角色名称
func (ac *AccessControl) HasRole(ctx code.Context) code.Response {
	ac.ctx = ctx

	args := ctx.Args()
	addr := string(args["address"])
	role := string(args["role"])

	if addr == "" || role == "" {
		return code.Errors("address and role are required")
	}

	hasRole := ac.hasRole(addr, role)

	result, _ := json.Marshal(map[string]interface{}{
		"address":  addr,
		"role":     role,
		"has_role": hasRole,
	})

	return code.OK(result)
}

// ListRoles 列出指定地址的所有角色
// 参数：address - 查询地址
func (ac *AccessControl) ListRoles(ctx code.Context) code.Response {
	ac.ctx = ctx

	args := ctx.Args()
	addr := string(args["address"])
	if addr == "" {
		return code.Errors("address is required")
	}

	validRoles := []string{RoleAdmin, RoleManufacturer, RoleLogistics, RoleDistributor, RoleAuditor, RoleConsumer}
	var roles []string

	for _, role := range validRoles {
		if ac.hasRole(addr, role) {
			roles = append(roles, role)
		}
	}

	result, _ := json.Marshal(map[string]interface{}{
		"address": addr,
		"roles":   roles,
	})

	return code.OK(result)
}

// ============================================================================
// 内部辅助方法
// ============================================================================

// hasRole 内部方法：检查地址是否拥有指定角色
func (ac *AccessControl) hasRole(addr, role string) bool {
	key := fmt.Sprintf("role_assignment:%s:%s", role, addr)
	data, _ := ac.ctx.GetObject([]byte(key))
	return data != nil
}

// isValidRole 检查角色名称是否在预定义列表中
func (ac *AccessControl) isValidRole(role string) bool {
	validRoles := map[string]bool{
		RoleAdmin:        true,
		RoleManufacturer: true,
		RoleLogistics:    true,
		RoleDistributor:  true,
		RoleAuditor:      true,
		RoleConsumer:     true,
	}
	return validRoles[role]
}
