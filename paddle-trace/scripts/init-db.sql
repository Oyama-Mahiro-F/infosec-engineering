-- =============================================================================
-- 乒乓球拍防伪溯源系统 - PostgreSQL初始化脚本
-- 创建用户表、密钥表、操作日志表等业务支撑表结构
-- =============================================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64)  NOT NULL UNIQUE,
    password_hash   VARCHAR(256) NOT NULL,          -- bcrypt哈希
    email           VARCHAR(128) NOT NULL UNIQUE,
    phone           VARCHAR(20),
    role            VARCHAR(32)  NOT NULL DEFAULT 'consumer', -- admin/manufacturer/logistics/distributor/auditor/consumer
    blockchain_addr VARCHAR(128),                   -- 关联的区块链账户地址
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 角色索引
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_blockchain_addr ON users(blockchain_addr);

-- NFC标签密钥表（加密存储）
CREATE TABLE IF NOT EXISTS nfc_tag_keys (
    id              BIGSERIAL PRIMARY KEY,
    tag_uid         VARCHAR(64)   NOT NULL UNIQUE,  -- NFC芯片硬件UID
    aes_key_encrypted TEXT        NOT NULL,          -- SM4加密后的AES-128密钥
    product_id      VARCHAR(128),                    -- 绑定的产品ID
    registered_by   VARCHAR(128),                    -- 注册方（品牌商地址）
    registered_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_nfc_keys_tag_uid ON nfc_tag_keys(tag_uid);
CREATE INDEX idx_nfc_keys_product_id ON nfc_tag_keys(product_id);

-- NFC验证日志表（审计用）
CREATE TABLE IF NOT EXISTS nfc_verification_logs (
    id              BIGSERIAL PRIMARY KEY,
    tag_uid         VARCHAR(64)   NOT NULL,
    product_id      VARCHAR(128),
    counter         BIGINT        NOT NULL,          -- 芯片计数器值
    verification_ok BOOLEAN       NOT NULL,          -- 是否通过验证
    fail_reason     VARCHAR(256),                    -- 失败原因
    client_ip       VARCHAR(64),                     -- 请求来源IP
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nfc_logs_tag_uid ON nfc_verification_logs(tag_uid);
CREATE INDEX idx_nfc_logs_created_at ON nfc_verification_logs(created_at);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT,
    action          VARCHAR(64)   NOT NULL,          -- 操作类型
    resource_type   VARCHAR(32),                     -- 资源类型(product/trace/role)
    resource_id     VARCHAR(128),                    -- 资源ID
    tx_id           VARCHAR(128),                    -- 区块链交易ID
    detail          JSONB,                           -- 操作详情（半结构化）
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_op_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_op_logs_created_at ON operation_logs(created_at);

-- 插入默认管理员账户（密码：admin123，bcrypt哈希）
INSERT INTO users (username, password_hash, email, role, blockchain_addr)
VALUES ('admin', '$2a$12$LJ3m4ys3GZfnYMz8kVsKaOSXWDRUPZGxKytAaCz5NFHcsG7vPJqOK',
        'admin@paddle-trace.com', 'admin', '0xAdminBlockchainAddr0000000001')
ON CONFLICT (username) DO NOTHING;

-- 插入演示制造商账户
INSERT INTO users (username, password_hash, email, role, blockchain_addr)
VALUES ('butterfly_factory', '$2a$12$LJ3m4ys3GZfnYMz8kVsKaOSXWDRUPZGxKytAaCz5NFHcsG7vPJqOK',
        'factory@butterfly.com', 'manufacturer', '0xManufacturerBlockchainAddr0002')
ON CONFLICT (username) DO NOTHING;
