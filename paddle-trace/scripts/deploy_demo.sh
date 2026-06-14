#!/usr/bin/env bash
# =============================================================================
# XuperChain 部署演示脚本
# 模拟在 Ubuntu 22.04 下部署百度超级链和智能合约的完整流程
# 运行: bash deploy_demo.sh
# 所有输出均为模拟，用于截图展示部署步骤
# =============================================================================

C_RESET='\033[0m'
C_GREEN='\033[1;32m'
C_BLUE='\033[1;34m'
C_YELLOW='\033[1;33m'
C_RED='\033[1;31m'
C_CYAN='\033[1;36m'

pause() { sleep 1.5; }
header() { echo -e "\n${C_BLUE}══════════════════════════════════════════════════════════════${C_RESET}"; echo -e "${C_YELLOW}  $1${C_RESET}"; echo -e "${C_BLUE}══════════════════════════════════════════════════════════════${C_RESET}\n"; }

# =============================================================================
# Step 1: 环境准备
# =============================================================================
header "Step 1: 环境准备 - 安装 XuperChain 依赖"
echo -e "${C_CYAN}$ go version${C_RESET}"
echo "go version go1.21.5 linux/amd64"
echo ""
echo -e "${C_CYAN}$ gcc --version | head -1${C_RESET}"
echo "gcc (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0"
echo ""
echo -e "${C_CYAN}$ git --version${C_RESET}"
echo "git version 2.34.1"
pause

# =============================================================================
# Step 2: 下载并编译 XuperChain
# =============================================================================
header "Step 2: 下载并编译百度超级链 (XuperChain)"
echo -e "${C_CYAN}$ git clone https://github.com/xuperchain/xuperchain.git${C_RESET}"
echo "Cloning into 'xuperchain'..."
echo "remote: Enumerating objects: 28456, done."
echo "remote: Counting objects: 100% (28456/28456), done."
echo "remote: Compressing objects: 100% (8931/8931), done."
echo "remote: Total 28456 (delta 18234), reused 27312 (delta 17211)"
echo "Receiving objects: 100% (28456/28456), 45.32 MiB | 8.21 MiB/s, done."
echo "Resolving deltas: 100% (18234/18234), done."
echo ""
echo -e "${C_CYAN}$ cd xuperchain && make${C_RESET}"
echo "Building xuperchain..."
echo "  -> compiling chain core..."
echo "  -> compiling consensus module (PBFT)..."
echo "  -> compiling P2P network module..."
echo "  -> compiling crypto module (SM2/SM3)..."
echo "  -> compiling smart contract engine (WASM)..."
echo "  -> compiling CLI tools..."
echo -e "${C_GREEN}Build complete: output/xchain, output/xchain-cli${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain --version${C_RESET}"
echo "xchain version 5.3.0 (build: 2026-01-15, commit: a3f2c8e)"
echo "  consensus: PBFT, XPoS"
echo "  crypto: SM2, SM3, ECDSA"
echo "  contract: WASM (Go, C++)"
pause

# =============================================================================
# Step 3: 创建网络配置
# =============================================================================
header "Step 3: 生成7节点联盟链网络配置"
echo -e "${C_CYAN}$ ./output/xchain-cli createChain --name paddle_chain${C_RESET}"
echo "Creating blockchain genesis config..."
echo "  -> generating genesis block..."
echo "  -> configuring 7 consensus nodes..."
echo "  -> setting consensus algorithm: PBFT"
echo "  -> block interval: 3s"
echo "  -> max block size: 2MB"
echo -e "${C_GREEN}Genesis config created: ./data/genesis/xuper.json${C_RESET}"
echo ""

echo -e "${C_CYAN}$ for i in 1 2 3 4 5 6 7; do${C_RESET}"
echo ">   ./output/xchain-cli createNode --name node\${i} --key-type sm2"
echo "> done"
echo ""
for i in 1 2 3 4 5 6 7; do
    echo -e "  ${C_GREEN}[OK]${C_RESET} Node $i created: ./data/node$i/"
    echo "      address: $(echo "SM2_$(openssl rand -hex 20)" | cut -c1-40)"
    echo "      consensus key: generated (SM2)"
    sleep 0.3
done
pause

# =============================================================================
# Step 4: 启动7节点网络
# =============================================================================
header "Step 4: 启动7节点PBFT联盟链网络"
echo -e "${C_CYAN}$ ./output/xchain --conf ./data/node1/xchain.yaml &${C_RESET}"
for i in 1 2 3 4 5 6 7; do
    echo -e "  ${C_GREEN}[Node $i]${C_RESET} Starting..."
    echo "  [Node $i] Loading config from ./data/node$i/xchain.yaml"
    echo "  [Node $i] P2P listening on 0.0.0.0:3710$i"
    echo "  [Node $i] RPC listening on 0.0.0.0:4710$i"
    echo "  [Node $i] Connecting to peers..."
    sleep 0.4
done
echo ""
echo "Waiting for consensus to stabilize..."
sleep 2
echo -e "${C_GREEN}[PBFT] Consensus established: 7/7 nodes online${C_RESET}"
echo -e "${C_GREEN}[PBFT] Current view: 0, leader: node1${C_RESET}"
echo -e "${C_GREEN}[PBFT] Block height: 1 (genesis)${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli status --host 127.0.0.1:47101${C_RESET}"
echo "{"
echo '  "name": "paddle_chain",'
echo '  "height": 1,'
echo '  "consensus": "PBFT",'
echo '  "peers": 7,'
echo '  "block_interval": "3s",'
echo '  "crypto": "SM2/SM3"'
echo "}"
pause

# =============================================================================
# Step 5: 创建合约账户
# =============================================================================
header "Step 5: 创建合约账户"

echo -e "${C_YELLOW}[注意：以下密钥和密码仅为演示用，生产环境请使用强随机密钥]${C_RESET}"
echo ""

echo -e "${C_CYAN}$ ./output/xchain-cli account new --crypto sm2${C_RESET}"
echo "Creating new account with SM2 key pair..."
echo "  Private Key: 9a3f7c2e1b5d..."
echo -e "  ${C_GREEN}Public Key:  04a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0${C_RESET}"
echo -e "  ${C_GREEN}Address:     XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli account set --desc \"paddle_contract_account\"${C_RESET}"
echo -e "${C_GREEN}Account label set: paddle_contract_account${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli account balance${C_RESET}"
echo "Address: XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1"
echo "Balance: 100000000000 (100 billion UTXO)"
pause

# =============================================================================
# Step 6: 编译智能合约
# =============================================================================
header "Step 6: 编译三个核心智能合约"
echo -e "${C_CYAN}$ cd contracts/access && GOOS=js GOARCH=wasm go build -o access_control.wasm${C_RESET}"
echo "Compiling AccessControl contract..."
echo "  -> parsing Go source..."
echo "  -> type checking..."
echo "  -> compiling to WASM..."
echo -e "${C_GREEN}  [OK] access_control.wasm (127 KB)${C_RESET}"
echo ""
echo -e "${C_CYAN}$ cd contracts/product && GOOS=js GOARCH=wasm go build -o product_registry.wasm${C_RESET}"
echo "Compiling ProductRegistry contract..."
echo -e "${C_GREEN}  [OK] product_registry.wasm (184 KB)${C_RESET}"
echo ""
echo -e "${C_CYAN}$ cd contracts/product && GOOS=js GOARCH=wasm go build -o traceability_log.wasm${C_RESET}"
echo "Compiling TraceabilityLog contract..."
echo -e "${C_GREEN}  [OK] traceability_log.wasm (196 KB)${C_RESET}"
pause

# =============================================================================
# Step 7: 部署合约到链上
# =============================================================================
header "Step 7: 部署智能合约到 XuperChain 网络"
echo -e "${C_CYAN}$ ./output/xchain-cli wasm deploy \\${C_RESET}"
echo ">   --account XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1 \\"
echo ">   --contract access_control \\"
echo ">   --wasm ./contracts/access/access_control.wasm \\"
echo ">   --init-args '{\"admin_addr\":\"0xAdmin01\"}'"
echo ""
echo "Building deployment transaction..."
echo "Signing with account key (SM2)..."
echo "Broadcasting to network..."
echo "Waiting for PBFT consensus (3 rounds)..."
sleep 2
echo -e "${C_GREEN}[TX CONFIRMED] txid: a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1${C_RESET}"
echo -e "${C_GREEN}  Contract 'access_control' deployed at block #2${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli wasm deploy \\${C_RESET}"
echo ">   --account XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1 \\"
echo ">   --contract product_registry \\"
echo ">   --wasm ./contracts/product/product_registry.wasm"
echo ""
sleep 2
echo -e "${C_GREEN}[TX CONFIRMED] txid: b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2${C_RESET}"
echo -e "${C_GREEN}  Contract 'product_registry' deployed at block #3${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli wasm deploy \\${C_RESET}"
echo ">   --account XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1 \\"
echo ">   --contract traceability_log \\"
echo ">   --wasm ./contracts/product/traceability_log.wasm"
echo ""
sleep 2
echo -e "${C_GREEN}[TX CONFIRMED] txid: c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3${C_RESET}"
echo -e "${C_GREEN}  Contract 'traceability_log' deployed at block #4${C_RESET}"
pause

# =============================================================================
# Step 8: 查看已部署的合约
# =============================================================================
header "Step 8: 查看已部署的合约列表"
echo -e "${C_CYAN}$ ./output/xchain-cli wasm list --account XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1${C_RESET}"
echo "Deployed contracts:"
echo "┌──────────────────┬──────────────────────────────────────────────┬──────────┬────────┐"
echo "│ Contract Name    │ Contract Address                             │ Version  │ Status │"
echo "├──────────────────┼──────────────────────────────────────────────┼──────────┼────────┤"
echo "│ access_control   │ XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1 │ v1.0.0   │ ACTIVE │"
echo "│ product_registry │ XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1 │ v1.0.0   │ ACTIVE │"
echo "│ traceability_log │ XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1 │ v1.0.0   │ ACTIVE │"
echo "└──────────────────┴──────────────────────────────────────────────┴──────────┴────────┘"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli wasm query --contract product_registry --method methods${C_RESET}"
echo "ProductRegistry methods:"
echo "  registerProduct(product_id, brand, model, batch_no, ...)"
echo "  transferOwnership(product_id, new_owner, transfer_type)"
echo "  verifyAuthenticity(product_id)"
echo "  getProduct(product_id)"
echo "  listProductsByManufacturer(manufacturer)"
pause

# =============================================================================
# Step 9: 调用合约测试
# =============================================================================
header "Step 9: 调用合约 - 注册产品 & 验证"
echo -e "${C_CYAN}$ ./output/xchain-cli wasm invoke \\${C_RESET}"
echo ">   --contract product_registry \\"
echo ">   --method registerProduct \\"
echo ">   --args '{\"product_id\":\"pingpong101\",\"brand\":\"Butterfly\",\"model\":\"VISCARIA FL\"}'"
echo ""
sleep 1.5
echo -e "${C_GREEN}[TX CONFIRMED] txid: d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4${C_RESET}"
echo -e "${C_GREEN}  Product 'pingpong101' registered at block #5${C_RESET}"
echo ""
echo -e "${C_CYAN}$ ./output/xchain-cli wasm query \\${C_RESET}"
echo ">   --contract product_registry \\"
echo ">   --method verifyAuthenticity \\"
echo ">   --args '{\"product_id\":\"pingpong101\"}'"
echo ""
echo '{'
echo '  "authentic": true,'
echo '  "product_id": "pingpong101",'
echo '  "brand": "Butterfly",'
echo '  "model": "VISCARIA FL",'
echo '  "manufacturer": "XC1A2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1",'
echo '  "current_status": "produced",'
echo '  "block_height": 5'
echo '}'
pause

# =============================================================================
# Step 10: 区块链状态总结
# =============================================================================
header "Step 10: 部署完成 - 区块链网络状态"
echo -e "${C_CYAN}$ curl -s http://127.0.0.1:47101/v1/status | python -m json.tool${C_RESET}"
echo "{"
echo '  "network": "paddle_chain",'
echo '  "consensus": "PBFT",'
echo '  "nodes": 7,'
echo '  "nodes_online": 7,'
echo '  "current_height": 5,'
echo '  "deployed_contracts": 3,'
echo '  "contract_names": ["access_control", "product_registry", "traceability_log"],'
echo '  "crypto": "SM2/SM3",'
echo '  "block_interval": "3s",'
echo '  "uptime": "5m 23s"'
echo "}"
echo ""

echo -e "${C_GREEN}╔══════════════════════════════════════════════════╗${C_RESET}"
echo -e "${C_GREEN}║  部署完成!                                        ║${C_RESET}"
echo -e "${C_GREEN}║                                                  ║${C_RESET}"
echo -e "${C_GREEN}║  区块链:    XuperChain 5.3.0 (PBFT)               ║${C_RESET}"
echo -e "${C_GREEN}║  节点数:    7 (品牌方x2+物流商x2+监管+协会+运维)    ║${C_RESET}"
echo -e "${C_GREEN}║  合约数:    3 (access_control, product_registry,   ║${C_RESET}"
echo -e "${C_GREEN}║                  traceability_log)                ║${C_RESET}"
echo -e "${C_GREEN}║  密码算法:  SM2/SM3 (国密)                        ║${C_RESET}"
echo -e "${C_GREEN}║  出块时间:  3s                                    ║${C_RESET}"
echo -e "${C_GREEN}╚══════════════════════════════════════════════════╝${C_RESET}"
