// Package service 区块链交互服务
// 封装与百度超级链的交互逻辑，包括智能合约调用和交易管理
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fansicheng/paddle-trace/config"
)

// ============================================================================
// 区块链服务
// ============================================================================

// BlockchainService 区块链交互服务
// 封装XuperChain SDK调用，为上层Handler提供简化的区块链操作接口
type BlockchainService struct {
	cfg         *config.Config
	contractAcc string // 合约账户名称
	client      *XChainClient // XuperChain SDK客户端
}

// NewBlockchainService 创建区块链服务实例
func NewBlockchainService(cfg *config.Config) (*BlockchainService, error) {
	// 初始化XuperChain客户端连接
	client, err := NewXChainClient(cfg.XChain.Nodes, cfg.XChain.ChainName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to XuperChain: %w", err)
	}

	return &BlockchainService{
		cfg:         cfg,
		contractAcc: "XC1234567890123456@paddle", // 合约账户
		client:      client,
	}, nil
}

// RegisterProduct 调用ProductRegistry合约注册新产品
func (bs *BlockchainService) RegisterProduct(ctx context.Context, req map[string]string) (string, error) {
	// 构建合约调用参数
	args := map[string][]byte{
		"product_id":    []byte(req["product_id"]),
		"brand":         []byte(req["brand"]),
		"model":         []byte(req["model"]),
		"batch_no":      []byte(req["batch_no"]),
		"material_hash": []byte(req["material_hash"]),
		"qc_report_hash": []byte(req["qc_report_hash"]),
		"produce_date":  []byte(req["produce_date"]),
	}

	// 调用智能合约
	resp, err := bs.client.InvokeContract(
		bs.contractAcc,
		"product_registry",
		"RegisterProduct",
		args,
	)
	if err != nil {
		return "", fmt.Errorf("contract invoke failed: %w", err)
	}

	if resp.Status >= 400 {
		return "", fmt.Errorf("contract error: %s", string(resp.Body))
	}

	log.Printf("[Blockchain] Product registered: %s, txid: %s",
		req["product_id"], resp.TxID)

	return resp.TxID, nil
}

// TransferOwnership 调用ProductRegistry合约转移产品所有权
func (bs *BlockchainService) TransferOwnership(ctx context.Context, productID, newOwner, transferType string) (string, error) {
	args := map[string][]byte{
		"product_id":    []byte(productID),
		"new_owner":     []byte(newOwner),
		"transfer_type": []byte(transferType),
	}

	resp, err := bs.client.InvokeContract(
		bs.contractAcc,
		"product_registry",
		"TransferOwnership",
		args,
	)
	if err != nil {
		return "", err
	}

	log.Printf("[Blockchain] Ownership transferred: %s -> %s, txid: %s",
		productID, newOwner, resp.TxID)

	return resp.TxID, nil
}

// VerifyAuthenticity 调用ProductRegistry合约验证产品真伪
func (bs *BlockchainService) VerifyAuthenticity(ctx context.Context, productID string) (map[string]interface{}, error) {
	args := map[string][]byte{
		"product_id": []byte(productID),
	}

	resp, err := bs.client.InvokeContract(
		bs.contractAcc,
		"product_registry",
		"VerifyAuthenticity",
		args,
	)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %w", err)
	}

	return result, nil
}

// QueryProduct 查询产品信息
func (bs *BlockchainService) QueryProduct(ctx context.Context, productID string) (map[string]interface{}, error) {
	args := map[string][]byte{
		"product_id": []byte(productID),
	}

	resp, err := bs.client.InvokeContract(
		bs.contractAcc,
		"product_registry",
		"GetProduct",
		args,
	)
	if err != nil {
		return nil, err
	}

	var product map[string]interface{}
	if err := json.Unmarshal(resp.Body, &product); err != nil {
		return nil, fmt.Errorf("unmarshal product error: %w", err)
	}

	return product, nil
}

// AppendTraceRecord 调用TraceabilityLog合约追加溯源记录
func (bs *BlockchainService) AppendTraceRecord(ctx context.Context, req map[string]string) (string, error) {
	args := map[string][]byte{
		"product_id": []byte(req["product_id"]),
		"action":     []byte(req["action"]),
		"location":   []byte(req["location"]),
		"extra_data": []byte(req["extra_data"]),
		"signature":  []byte(req["signature"]),
	}

	resp, err := bs.client.InvokeContract(
		bs.contractAcc,
		"traceability_log",
		"AppendTraceRecord",
		args,
	)
	if err != nil {
		return "", err
	}

	log.Printf("[Blockchain] Trace record appended for %s: %s, txid: %s",
		req["product_id"], req["action"], resp.TxID)

	return resp.TxID, nil
}

// QueryTraceHistory 查询产品溯源历史
func (bs *BlockchainService) QueryTraceHistory(ctx context.Context, productID string) (map[string]interface{}, error) {
	args := map[string][]byte{
		"product_id": []byte(productID),
	}

	resp, err := bs.client.InvokeContract(
		bs.contractAcc,
		"traceability_log",
		"QueryHistory",
		args,
	)
	if err != nil {
		return nil, err
	}

	var history map[string]interface{}
	if err := json.Unmarshal(resp.Body, &history); err != nil {
		return nil, fmt.Errorf("unmarshal history error: %w", err)
	}

	return history, nil
}

// ============================================================================
// XuperChain客户端（简化实现）
// ============================================================================

// XChainClient XuperChain SDK客户端封装
type XChainClient struct {
	nodes     []string
	chainName string
}

// ContractResponse 合约调用响应
type ContractResponse struct {
	TxID   string
	Status int
	Body   []byte
	GasUsed int64
}

// NewXChainClient 创建XuperChain客户端
func NewXChainClient(nodes []string, chainName string) (*XChainClient, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no blockchain nodes configured")
	}

	// 原型阶段：使用第一个节点作为主要RPC入口
	// 生产环境应实现负载均衡和故障转移
	log.Printf("[XChain] Connecting to %d nodes on chain '%s'", len(nodes), chainName)
	log.Printf("[XChain] Primary node: %s", nodes[0])

	// TODO: 集成XuperChain Go SDK完整初始化
	// 包括：加载客户端私钥、建立gRPC连接、验证节点健康状态

	return &XChainClient{
		nodes:     nodes,
		chainName: chainName,
	}, nil
}

// InvokeContract 调用智能合约方法
func (c *XChainClient) InvokeContract(account, contract, method string, args map[string][]byte) (*ContractResponse, error) {
	// 原型阶段：构建并记录合约调用
	// 生产环境应通过XuperChain Go SDK的InvokeContract方法实际调用

	log.Printf("[XChain] Invoke: account=%s, contract=%s, method=%s", account, contract, method)
	for k, v := range args {
		log.Printf("[XChain]   arg[%s] = %s", k, string(v))
	}

	// 模拟合约调用延迟（PBFT共识需3轮广播，约1-3秒）
	time.Sleep(500 * time.Millisecond)

	// 返回模拟的成功响应
	return &ContractResponse{
		TxID:   fmt.Sprintf("tx_%d_%s", time.Now().UnixNano(), method[:min(8, len(method))]),
		Status: 200,
		Body:   []byte(`{"status":"success"}`),
		GasUsed: 0, // 联盟链零Gas
	}, nil
}

// Close 关闭客户端连接
func (c *XChainClient) Close() error {
	log.Println("[XChain] Client closed")
	return nil
}
