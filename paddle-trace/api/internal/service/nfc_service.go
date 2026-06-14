// Package service NFC标签安全验证服务
// 实现NTAG 424 DNA芯片的SUN（Secure Unique NFC）动态认证码验证逻辑
// 此服务是系统"芯片不可克隆"安全保证的核心
package service

import (
	"crypto/aes"
	"crypto/cmac" // golang.org/x/crypto/cmac
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// ============================================================================
// NFC验证服务
// ============================================================================

// NFCService NFC标签验证服务
type NFCService struct {
	mu            sync.RWMutex
	keyStore      map[string][]byte // tagUID -> AES-128密钥映射（原型阶段内存存储）
	counterStore  map[string]int64  // tagUID -> 最后成功验证的计数器值（防重放）
}

// NewNFCService 创建NFC验证服务实例
func NewNFCService() *NFCService {
	return &NFCService{
		keyStore:     make(map[string][]byte),
		counterStore: make(map[string]int64),
	}
}

// RegisterTag 注册新NFC标签
// 将标签UID与AES密钥关联存储
// 参数：tagUID - 标签硬件UID, aesKey - AES-128密钥（16字节hex编码）
func (s *NFCService) RegisterTag(tagUID, aesKeyHex string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	aesKey, err := hex.DecodeString(aesKeyHex)
	if err != nil {
		return fmt.Errorf("invalid AES key: %w", err)
	}
	if len(aesKey) != 16 {
		return errors.New("AES key must be 16 bytes (128 bits)")
	}

	s.keyStore[tagUID] = aesKey
	s.counterStore[tagUID] = 0
	return nil
}

// VerifySUN 验证SUN动态认证码
// 这是防伪溯源的物理安全核心——确保NFC芯片为原始芯片，非克隆品
//
// 验证流程：
// 1. 使用标签对应的AES-128密钥计算期望的CMAC认证码
// 2. 比对接收到的SUN认证码与计算值
// 3. 检查计数器值是否递增（防重放攻击）
//
// 参数：
//   - tagUID: NFC芯片唯一硬件UID
//   - sunCode: 芯片生成的SUN动态认证码（hex编码）
//   - counter: 芯片内部单调计数器的当前值
//
// 返回：验证是否通过，以及描述信息
func (s *NFCService) VerifySUN(tagUID, sunCodeHex string, counter int64) (bool, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. 获取标签对应的AES密钥
	aesKey, exists := s.keyStore[tagUID]
	if !exists {
		return false, "tag not registered", errors.New("unknown tag UID: " + tagUID)
	}

	// 2. 防重放攻击：检查计数器是否递增
	lastCounter := s.counterStore[tagUID]
	if counter <= lastCounter {
		return false,
			fmt.Sprintf("replay attack detected: counter %d <= last %d", counter, lastCounter),
			errors.New("replay attack detected")
	}

	// 3. 计算期望的SUN认证码
	// SUN码 = CMAC(AES_Key, UID || counter)
	expectedCode, err := s.computeCMAC(aesKey, tagUID, counter)
	if err != nil {
		return false, "computation error", err
	}

	// 4. 恒时比较（防时序侧信道攻击）
	receivedCode, err := hex.DecodeString(sunCodeHex)
	if err != nil {
		return false, "invalid SUN code format", err
	}

	if subtle.ConstantTimeCompare(expectedCode, receivedCode) != 1 {
		return false,
			"SUN verification failed: code mismatch (possible clone or man-in-the-middle)",
			errors.New("SUN code mismatch")
	}

	// 5. 验证通过：更新计数器
	s.counterStore[tagUID] = counter

	return true, "SUN verified: authentic NTAG 424 DNA chip", nil
}

// computeCMAC 计算AES-CMAC认证码
// 输入：AES密钥、标签UID、计数器值
// 输出：16字节CMAC认证码
func (s *NFCService) computeCMAC(key []byte, tagUID string, counter int64) ([]byte, error) {
	// 构建CMAC输入：UID || counter (16字节)
	uidBytes, err := hex.DecodeString(tagUID)
	if err != nil {
		return nil, fmt.Errorf("invalid UID: %w", err)
	}

	input := make([]byte, 16)
	copy(input[0:8], uidBytes[:min(8, len(uidBytes))])
	input[8]  = byte(counter >> 56)
	input[9]  = byte(counter >> 48)
	input[10] = byte(counter >> 40)
	input[11] = byte(counter >> 32)
	input[12] = byte(counter >> 24)
	input[13] = byte(counter >> 16)
	input[14] = byte(counter >> 8)
	input[15] = byte(counter)

	// 创建AES-CMAC
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes init error: %w", err)
	}

	// 使用AES-CMAC算法计算认证码
	mac, err := cmac.New(block)
	if err != nil {
		return nil, fmt.Errorf("cmac init error: %w", err)
	}

	mac.Write(input)
	return mac.Sum(nil), nil
}

// GetLastCounter 获取标签的最后验证计数器值（用于调试）
func (s *NFCService) GetLastCounter(tagUID string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	counter, exists := s.counterStore[tagUID]
	return counter, exists
}

// ============================================================================
// 辅助函数
// ============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
