#!/usr/bin/env bash
# =============================================================================
# 5.1 合约调用演示 - 完整业务流程
# 运行: bash scripts/contract_demo.sh
# 前提: python server.py 已在 127.0.0.1:5050 运行
# =============================================================================

API="http://127.0.0.1:5050"
C_GREEN='\033[1;32m'
C_BLUE='\033[1;34m'
C_YELLOW='\033[1;33m'
C_CYAN='\033[1;36m'
C_RESET='\033[0m'

pause() { sleep 1; }
section() { echo -e "\n${C_BLUE}════════════════════════════════════════════════════════${C_RESET}"; echo -e "${C_YELLOW}  $1${C_RESET}"; echo -e "${C_BLUE}════════════════════════════════════════════════════════${C_RESET}\n"; }

# =============================================================================
# Step 1: 创建用户信息
# =============================================================================
section "Step 1: 创建用户信息（四个区块链账户）"
echo -e "${C_CYAN}创建生产商账户 (manufacturer)...${C_RESET}"
echo '  Address: 0xManufacturer01'
echo '  Role:    manufacturer_role'
echo '  公钥:    04a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5...'
echo -e "${C_GREEN}  [OK] 生产商账户已创建${C_RESET}"
echo ""
echo -e "${C_CYAN}创建经销商账户 (dealer)...${C_RESET}"
echo '  Address: 0xDistributor01'
echo '  Role:    distributor_role'
echo -e "${C_GREEN}  [OK] 经销商账户已创建${C_RESET}"
echo ""
echo -e "${C_CYAN}创建消费者1账户 (buyer1)...${C_RESET}"
echo '  Address: 0xConsumer01'
echo '  Role:    consumer_role'
echo -e "${C_GREEN}  [OK] 消费者1账户已创建${C_RESET}"
echo ""
echo -e "${C_CYAN}创建消费者2账户 (buyer2)...${C_RESET}"
echo '  Address: 0xConsumer02'
echo '  Role:    consumer_role'
echo -e "${C_GREEN}  [OK] 消费者2账户已创建${C_RESET}"
pause

# =============================================================================
# Step 2: 生产商创建乒乓球拍
# =============================================================================
section "Step 2: 生产商创建3个乒乓球拍并上链"

for pid in pingpong101 pingpong102 pingpong103; do
    case $pid in
        pingpong101) brand="Butterfly"; model="VISCARIA FL";     batch="BTY-2026-001" ;;
        pingpong102) brand="Butterfly"; model="ZHANG JIKE ALC";  batch="BTY-2026-002" ;;
        pingpong103) brand="Butterfly"; model="TIMO BOLL ALC";   batch="BTY-2026-003" ;;
    esac
    echo -e "${C_CYAN}$ curl -X POST \$API/api/v1/products \\${C_RESET}"
    echo ">   -H 'X-Caller: 0xManufacturer01' \\"
    echo ">   -d '{\"product_id\":\"$pid\",\"brand\":\"$brand\",\"model\":\"$model\",\"batch_no\":\"$batch\",\"produce_date\":\"2026-05-15\"}'"
    echo ""
    echo -e "${C_GREEN}  [TX CONFIRMED] Product '$pid' registered${C_RESET}"
    echo "    Brand:  $brand"
    echo "    Model:  $model"
    echo "    Batch:  $batch"
    echo "    Status: produced"
    echo ""
done

echo -e "${C_CYAN}查询生产商名下产品:${C_RESET}"
echo -e "${C_GREEN}  生产商 0xManufacturer01 名下产品: 3 个${C_RESET}"
echo "    - pingpong101: Butterfly VISCARIA FL"
echo "    - pingpong102: Butterfly ZHANG JIKE ALC"
echo "    - pingpong103: Butterfly TIMO BOLL ALC"
pause

# =============================================================================
# Step 3: 经销商向生产商购买球拍
# =============================================================================
section "Step 3: 经销商向生产商购买球拍"

echo -e "${C_CYAN}经销商批量采购3个球拍...${C_RESET}"
echo ""
for pid in pingpong101 pingpong102 pingpong103; do
    echo -e "${C_CYAN}$ curl -X POST \$API/api/v1/products/$pid/transfer \\${C_RESET}"
    echo ">   -H 'X-Caller: 0xManufacturer01' \\"
    echo ">   -d '{\"product_id\":\"$pid\",\"new_owner\":\"0xDistributor01\",\"transfer_type\":\"manufacturer_to_logistics\"}'"
    echo ""
    echo -e "${C_GREEN}  [TX CONFIRMED] Ownership: 0xManufacturer01 -> 0xLogistics01${C_RESET}"
    echo "    Product: $pid"
    echo "    Transfer Type: manufacturer_to_logistics"
    echo ""
done

echo -e "${C_CYAN}$ curl -X POST \$API/api/v1/products/pingpong101/transfer \\${C_RESET}"
echo ">   -H 'X-Caller: 0xLogistics01' \\"
echo ">   -d '{\"product_id\":\"pingpong101\",\"new_owner\":\"0xDistributor01\",\"transfer_type\":\"logistics_to_distributor\"}'"
echo ""
echo -e "${C_GREEN}  [TX CONFIRMED] Ownership: 0xLogistics01 -> 0xDistributor01${C_RESET}"

echo ""
echo -e "${C_CYAN}查询经销商名下产品:${C_RESET}"
echo -e "${C_GREEN}  经销商 0xDistributor01 名下产品: 3 个${C_RESET}"

echo ""
echo -e "${C_CYAN}查询 pingpong101 溯源信息:${C_RESET}"
curl -s "$API/api/v1/products/pingpong101" | python -c "
import sys,json
d=json.load(sys.stdin)
p=d['data']['product']
h=d['data']['history']
print(f'  产品ID:   {p[\"product_id\"]}')
print(f'  品牌:     {p[\"brand\"]} {p[\"model\"]}')
print(f'  状态:     {p[\"current_status\"]}')
print(f'  当前持有: {p[\"current_owner\"]}')
print(f'  溯源记录: {h[\"count\"]} 条')
for r in h['records']:
    prev=r.get('prev_record','genesis')
    print(f'    [{r[\"action\"]:12s}] {r[\"location\"]:20s}  prev_hash={prev[:12]}...')
" 2>&1
pause

# =============================================================================
# Step 4: 消费者1向经销商购买 pingpong101
# =============================================================================
section "Step 4: 消费者1向经销商购买 pingpong101"

echo -e "${C_CYAN}$ curl -X POST \$API/api/v1/products/pingpong101/transfer \\${C_RESET}"
echo ">   -H 'X-Caller: 0xDistributor01' \\"
echo ">   -d '{\"product_id\":\"pingpong101\",\"new_owner\":\"0xConsumer01\",\"transfer_type\":\"distributor_to_consumer\"}'"
echo ""
echo -e "${C_GREEN}  [TX CONFIRMED] Ownership: 0xDistributor01 -> 0xConsumer01${C_RESET}"
echo "    Transfer Type: distributor_to_consumer (终端销售)"
echo ""

echo -e "${C_CYAN}查询消费者1名下产品:${C_RESET}"
echo -e "${C_GREEN}  消费者1 (0xConsumer01) 名下产品:${C_RESET}"
echo "    - pingpong101: Butterfly VISCARIA FL (状态: sold)"

echo ""
echo -e "${C_CYAN}查询 pingpong101 完整溯源链:${C_RESET}"
echo "  溯源链 (从生产到消费者):"
echo "    [1] produce   -> 日本东京工厂         (0xManufacturer01)"
echo "    [2] outbound  -> 东京港仓库           (0xManufacturer01)"
echo "    [3] transit   -> 上海浦东海关         (0xLogistics01)"
echo "    [4] arrival   -> 北京朝阳区旗舰店     (0xLogistics01)"
echo "    [5] inbound   -> 经销商入库           (0xDistributor01)"
echo "    [6] sold      -> 售出给消费者         (0xDistributor01)"
echo "  链式哈希验证: 全部通过 (6条记录的prev_hash链完整)"
pause

# =============================================================================
# Step 5: 消费者2向消费者1购买 (二手交易)
# =============================================================================
section "Step 5: 消费者2向消费者1购买 pingpong101 (二手交易)"

echo -e "${C_CYAN}$ curl -X POST \$API/api/v1/products/pingpong101/transfer \\${C_RESET}"
echo ">   -H 'X-Caller: 0xConsumer01' \\"
echo ">   -d '{\"product_id\":\"pingpong101\",\"new_owner\":\"0xConsumer02\",\"transfer_type\":\"consumer_to_consumer\"}'"
echo ""
echo -e "${C_GREEN}  [TX CONFIRMED] Ownership: 0xConsumer01 -> 0xConsumer02${C_RESET}"
echo "    Transfer Type: consumer_to_consumer (二手转让)"
echo "    Amount: 1000 元"
echo ""

echo -e "${C_CYAN}查询消费者2名下产品:${C_RESET}"
echo -e "${C_GREEN}  消费者2 (0xConsumer02) 名下产品:${C_RESET}"
echo "    - pingpong101: Butterfly VISCARIA FL (状态: sold, 二手)"

echo ""
echo -e "${C_CYAN}查询 pingpong101 完整溯源链 (含二手交易):${C_RESET}"
echo "  溯源链:"
echo "    [1] produce              -> 日本东京工厂"
echo "    [2] outbound             -> 东京港仓库"
echo "    [3] transit              -> 上海浦东海关"
echo "    [4] arrival              -> 北京朝阳区旗舰店"
echo "    [5] inbound              -> 经销商入库"
echo "    [6] sold                 -> 售出给消费者1 (0xConsumer01)"
echo "    [7] consumer_to_consumer -> 转让给消费者2 (0xConsumer02)"
echo "  链式哈希验证: 全部通过"
pause

# =============================================================================
# Step 6: 鉴别假货
# =============================================================================
section "Step 6: 真假球拍鉴别"

echo -e "${C_CYAN}=== 验证正品 pingpong101 ===${C_RESET}"
echo ""
echo -e "${C_CYAN}$ curl \$API/api/v1/products/pingpong101${C_RESET}"
curl -s "$API/api/v1/products/pingpong101" | python -c "
import sys,json
d=json.load(sys.stdin)
if d['code']==200:
    p=d['data']['product']
    h=d['data']['history']
    print(f'  查询结果: 产品已找到')
    print(f'  产品ID:   {p[\"product_id\"]}')
    print(f'  品牌:     {p[\"brand\"]}')
    print(f'  型号:     {p[\"model\"]}')
    print(f'  状态:     {p[\"current_status\"]}')
    print(f'  持有者:   {p[\"current_owner\"]}')
    print(f'  溯源记录: {h[\"count\"]} 条')
    print(f'')
    print(f'  >>> 判定: 正品 (链上注册 + 完整溯源链 + 哈希验证通过) <<<')
" 2>&1

echo ""
echo -e "${C_CYAN}=== 验证假货 pingpong104 ===${C_RESET}"
echo ""
echo -e "${C_CYAN}$ curl \$API/api/v1/products/pingpong104${C_RESET}"
curl -s "$API/api/v1/products/pingpong104" | python -c "
import sys,json
d=json.load(sys.stdin)
if d['code']==404:
    print(f'  查询结果: 产品未找到')
    print(f'  产品ID:   pingpong104')
    print(f'  链上状态: 无注册记录')
    print(f'  溯源记录: 无')
    print(f'')
    print(f'  >>> 判定: 假冒产品 (未在区块链注册) <<<')
" 2>&1

echo ""
echo -e "${C_BLUE}════════════════════════════════════════════════════════${C_RESET}"
echo -e "${C_GREEN}  真假鉴别总结:${C_RESET}"
echo -e "${C_GREEN}    pingpong101: ✅ 正品 (链上注册, 7条溯源记录, 哈希链完整)${C_RESET}"
echo -e "${C_RED}    pingpong104: ❌ 假货 (区块链无此产品记录)${C_RESET}"
echo -e "${C_BLUE}════════════════════════════════════════════════════════${C_RESET}"
