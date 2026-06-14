"""
乒乓球拍防伪溯源系统 - 完整可运行后端
====================================================
内嵌区块链（哈希链式存储）+ NFC SUN动态认证 + REST API
直接运行: python server.py
测试方式: 浏览器打开 frontend/consumer/index.html 或 curl
"""

import hashlib
import hmac
import json
import time
import secrets
from datetime import datetime, timezone
from flask import Flask, request, jsonify
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# =============================================================================
# 第一部分：区块链引擎（内存版，哈希链式存储，不可篡改）
# =============================================================================

class Block:
    """区块结构"""
    def __init__(self, index, timestamp, data, previous_hash):
        self.index = index
        self.timestamp = timestamp
        self.data = data          # 交易数据
        self.previous_hash = previous_hash
        self.hash = self.compute_hash()

    def compute_hash(self):
        payload = f"{self.index}{self.timestamp}{json.dumps(self.data, sort_keys=True)}{self.previous_hash}"
        return hashlib.sha256(payload.encode()).hexdigest()

class Blockchain:
    """
    简易区块链引擎
    - 链式哈希保证不可篡改
    - 支持合约调用、产品注册、溯源记录
    """
    def __init__(self):
        self.chain = []
        self.pending_operations = []
        # 合约存储（模拟XuperChain的KV存储）
        self.state = {}  # key -> value

        # 创建创世块
        genesis = Block(0, int(time.time()), {"type": "genesis"}, "0" * 64)
        self.chain.append(genesis)

    def add_block(self, data):
        last = self.chain[-1]
        block = Block(len(self.chain), int(time.time()), data, last.hash)
        self.chain.append(block)
        return block

    def put_state(self, key, value):
        """模拟合约 PutObject"""
        self.state[key] = value

    def get_state(self, key):
        """模拟合约 GetObject"""
        return self.state.get(key)

    def delete_state(self, key):
        """模拟合约 DeleteObject"""
        if key in self.state:
            del self.state[key]

    def verify_chain(self):
        """验证整条链的完整性"""
        for i in range(1, len(self.chain)):
            current = self.chain[i]
            previous = self.chain[i-1]
            if current.previous_hash != previous.hash:
                return False, i
            if current.hash != current.compute_hash():
                return False, i
        return True, -1

    def get_block_by_tx(self, tx_type, key):
        """按交易类型和键查找区块"""
        for block in self.chain:
            d = block.data
            if d.get("type") == tx_type and d.get("key") == key:
                return block
        return None


# 全局区块链实例
bc = Blockchain()


# =============================================================================
# 第二部分：智能合约逻辑
# =============================================================================

# ---- 角色常量 ----
ROLES = {
    "admin_role":        "系统管理员",
    "manufacturer_role": "制造商",
    "logistics_role":    "物流商",
    "distributor_role":  "经销商",
    "auditor_role":      "审计方",
    "consumer_role":     "消费者",
}

# 初始角色分配（区块链地址 -> 角色列表）
role_assignments = {}

def grant_role(addr, role, granted_by="genesis"):
    key = f"role:{role}:{addr}"
    if bc.get_state(key):
        return False, "role already assigned"
    bc.put_state(key, json.dumps({
        "address": addr, "role": role,
        "granted_by": granted_by,
        "granted_at": int(time.time())
    }))
    if addr not in role_assignments:
        role_assignments[addr] = []
    role_assignments[addr].append(role)
    return True, f"role {role} granted to {addr}"

def has_role(addr, role):
    key = f"role:{role}:{addr}"
    return bc.get_state(key) is not None

# 初始化演示账户
INIT_ACCOUNTS = {
    "0xManufacturer01":  "manufacturer_role",
    "0xLogistics01":     "logistics_role",
    "0xDistributor01":   "distributor_role",
    "0xAdmin01":         "admin_role",
    "0xAuditor01":       "auditor_role",
}
for acc, role in INIT_ACCOUNTS.items():
    grant_role(acc, role)


# ---- 产品注册合约 ----
def contract_register_product(caller, args):
    """ProductRegistry.RegisterProduct"""
    if not has_role(caller, "manufacturer_role"):
        return {"error": "permission denied: caller is not manufacturer"}

    pid = args.get("product_id")
    if bc.get_state(f"product:{pid}"):
        return {"error": f"product already registered: {pid}"}

    product = {
        "product_id":      pid,
        "brand":           args.get("brand", ""),
        "model":           args.get("model", ""),
        "batch_no":        args.get("batch_no", ""),
        "produce_date":    args.get("produce_date", ""),
        "manufacturer":    caller,
        "material_hash":   args.get("material_hash", ""),
        "qc_report_hash":  args.get("qc_report_hash", ""),
        "current_owner":   caller,
        "current_status":  "produced",
        "created_at":      int(time.time()),
        "updated_at":      int(time.time()),
    }

    bc.put_state(f"product:{pid}", json.dumps(product))

    # 初始所有权记录
    bc.put_state(f"ownership:{pid}:0", json.dumps({
        "product_id": pid, "from_owner": "genesis",
        "to_owner": caller, "transfer_type": "initial_registration",
        "transferred_at": int(time.time())
    }))

    # 记录到区块链区块
    bc.add_block({
        "type": "RegisterProduct", "key": pid,
        "caller": caller, "product": product,
        "timestamp": int(time.time())
    })

    return {"success": True, "message": f"product registered: {pid}"}


def contract_transfer_ownership(caller, args):
    """ProductRegistry.TransferOwnership"""
    pid = args.get("product_id")
    new_owner = args.get("new_owner")
    transfer_type = args.get("transfer_type")

    raw = bc.get_state(f"product:{pid}")
    if not raw:
        return {"error": f"product not found: {pid}"}

    product = json.loads(raw)
    if product["current_owner"] != caller:
        return {"error": "transfer denied: caller is not current owner"}

    status_map = {
        "manufacturer_to_logistics": "in_transit",
        "logistics_to_distributor": "in_stock",
        "distributor_to_consumer": "sold",
        "consumer_to_consumer": "sold",
    }
    old_owner = product["current_owner"]
    product["current_owner"] = new_owner
    product["current_status"] = status_map.get(transfer_type, product["current_status"])
    product["updated_at"] = int(time.time())

    bc.put_state(f"product:{pid}", json.dumps(product))

    idx = int(time.time())
    bc.put_state(f"ownership:{pid}:{idx}", json.dumps({
        "product_id": pid, "from_owner": old_owner,
        "to_owner": new_owner, "transfer_type": transfer_type,
        "transferred_at": int(time.time())
    }))

    bc.add_block({
        "type": "TransferOwnership", "key": pid,
        "caller": caller, "from": old_owner, "to": new_owner,
        "transfer_type": transfer_type, "timestamp": int(time.time())
    })

    return {"success": True, "message": f"ownership: {old_owner} -> {new_owner}"}


def contract_verify_authenticity(args):
    """ProductRegistry.VerifyAuthenticity"""
    pid = args.get("product_id")
    raw = bc.get_state(f"product:{pid}")
    if not raw:
        return {"authentic": False, "product_id": pid,
                "reason": "product not registered on blockchain",
                "recommendation": "This product is NOT registered. It may be counterfeit."}

    product = json.loads(raw)
    return {
        "authentic": True, "product_id": pid,
        "brand": product["brand"], "model": product["model"],
        "manufacturer": product["manufacturer"],
        "produce_date": product["produce_date"],
        "current_status": product["current_status"],
        "current_owner": product["current_owner"],
        "registered_at": product["created_at"],
    }


def contract_query_product(args):
    """ProductRegistry.GetProduct"""
    pid = args.get("product_id")
    raw = bc.get_state(f"product:{pid}")
    if not raw:
        return None
    return json.loads(raw)


# ---- 溯源记录合约 ----
VALID_ACTIONS = {"produce", "outbound", "transit", "arrival", "inbound", "on_sale", "sold", "return", "quality_check"}

def contract_append_trace(caller, args):
    """TraceabilityLog.AppendTraceRecord"""
    pid = args.get("product_id")
    action = args.get("action")
    location = args.get("location", "")
    extra_data = args.get("extra_data", "")
    signature = args.get("signature", "")

    if not bc.get_state(f"product:{pid}"):
        return {"error": f"product not found: {pid}"}
    if action not in VALID_ACTIONS:
        return {"error": f"invalid action: {action}"}

    # 获取计数器
    counter_raw = bc.get_state(f"counter:{pid}")
    count = int(counter_raw) + 1 if counter_raw else 1

    record_id = f"{pid}_{count}"

    # 前一条记录哈希
    prev_hash = "genesis"
    if count > 1:
        prev_raw = bc.get_state(f"trace:{pid}:{pid}_{count-1}")
        if prev_raw:
            prev_hash = hashlib.sha256(prev_raw.encode()).hexdigest()[:16]

    record = {
        "record_id": record_id, "product_id": pid,
        "operator": caller, "action": action,
        "location": location, "timestamp": int(time.time()),
        "prev_record": prev_hash, "extra_data": extra_data,
        "signature": signature,
    }

    record_json = json.dumps(record)
    bc.put_state(f"trace:{pid}:{record_id}", record_json)
    bc.put_state(f"counter:{pid}", str(count))

    bc.add_block({
        "type": "AppendTraceRecord", "key": record_id,
        "caller": caller, "trace": record,
        "timestamp": int(time.time())
    })

    return {"success": True, "message": f"trace record appended: {record_id}"}


def contract_query_history(args):
    """TraceabilityLog.QueryHistory"""
    pid = args.get("product_id")
    counter_raw = bc.get_state(f"counter:{pid}")
    if not counter_raw:
        return {"product_id": pid, "count": 0, "records": []}

    count = int(counter_raw)
    records = []
    for i in range(1, count + 1):
        rid = f"{pid}_{i}"
        raw = bc.get_state(f"trace:{pid}:{rid}")
        if raw:
            records.append(json.loads(raw))

    return {"product_id": pid, "count": len(records), "records": records}


# =============================================================================
# 第三部分：NFC SUN 动态认证引擎
# =============================================================================

class NFCService:
    """NTAG 424 DNA SUN认证服务"""
    def __init__(self):
        self.keys = {}       # tag_uid -> aes_key (bytes)
        self.counters = {}   # tag_uid -> last_counter

    def register_tag(self, tag_uid, aes_key_hex):
        key = bytes.fromhex(aes_key_hex)
        self.keys[tag_uid] = key
        self.counters[tag_uid] = 0
        return True

    def verify_sun(self, tag_uid, sun_code_hex, counter):
        """验证SUN动态认证码"""
        if tag_uid not in self.keys:
            return False, "unknown tag", "tag not registered"

        # 防重放
        last = self.counters.get(tag_uid, 0)
        if counter <= last:
            return False, f"replay attack: counter {counter} <= {last}", "replay"

        # 计算期望CMAC
        key = self.keys[tag_uid]
        expected = self._compute_cmac(key, tag_uid, counter)
        received = bytes.fromhex(sun_code_hex)

        if not hmac.compare_digest(expected, received):
            return False, "SUN mismatch: possible clone", "mismatch"

        # 验证通过
        self.counters[tag_uid] = counter
        return True, "authentic NTAG 424 DNA chip", "ok"

    def _compute_cmac(self, key, tag_uid, counter):
        """AES-CMAC 简化实现（原型用HMAC-SHA256替代）"""
        msg = f"{tag_uid}:{counter}".encode()
        return hmac.new(key, msg, hashlib.sha256).digest()[:16]


nfc = NFCService()
# 注册演示标签
nfc.register_tag("04A2B3C4D5E6F7", "00112233445566778899AABBCCDDEEFF")
nfc.register_tag("04F7E6D5C4B3A2", "FFEEDDCCBBAA99887766554433221100")


# =============================================================================
# 第四部分：演示数据初始化
# =============================================================================

def init_demo_data():
    """预置演示产品数据，构建完整溯源链"""
    manufacturer = "0xManufacturer01"
    logistics = "0xLogistics01"
    distributor = "0xDistributor01"

    # 产品1：完整溯源链
    contract_register_product(manufacturer, {
        "product_id": "pingpong101", "brand": "Butterfly",
        "model": "VISCARIA FL", "batch_no": "BTY-2026-001",
        "material_hash": "sm3_hash_materials_abc", "qc_report_hash": "ipfs://QmReportXYZ",
        "produce_date": "2026-05-15"
    })
    contract_transfer_ownership(manufacturer, {"product_id": "pingpong101", "new_owner": logistics, "transfer_type": "manufacturer_to_logistics"})
    contract_append_trace(manufacturer, {"product_id": "pingpong101", "action": "outbound", "location": "日本东京工厂", "extra_data": "{}", "signature": "sig_mfr_01"})
    contract_append_trace(logistics, {"product_id": "pingpong101", "action": "transit", "location": "东京港仓库", "extra_data": "{}", "signature": "sig_log_01"})
    contract_append_trace(logistics, {"product_id": "pingpong101", "action": "arrival", "location": "上海浦东国际机场:海关", "extra_data": "{}", "signature": "sig_log_02"})
    contract_transfer_ownership(logistics, {"product_id": "pingpong101", "new_owner": distributor, "transfer_type": "logistics_to_distributor"})
    contract_append_trace(distributor, {"product_id": "pingpong101", "action": "inbound", "location": "北京朝阳区旗舰店", "extra_data": "{}", "signature": "sig_dist_01"})
    contract_transfer_ownership(distributor, {"product_id": "pingpong101", "new_owner": "0xConsumer01", "transfer_type": "distributor_to_consumer"})

    # 产品2：较短溯源链（仍在经销商库存）
    contract_register_product(manufacturer, {
        "product_id": "pingpong102", "brand": "Butterfly",
        "model": "ZHANG JIKE ALC", "batch_no": "BTY-2026-002",
        "material_hash": "sm3_hash_materials_def", "qc_report_hash": "ipfs://QmReportUVW",
        "produce_date": "2026-05-20"
    })
    contract_transfer_ownership(manufacturer, {"product_id": "pingpong102", "new_owner": logistics, "transfer_type": "manufacturer_to_logistics"})
    contract_append_trace(logistics, {"product_id": "pingpong102", "action": "transit", "location": "上海区域配送中心", "extra_data": "{}", "signature": "sig_log_03"})
    contract_transfer_ownership(logistics, {"product_id": "pingpong102", "new_owner": distributor, "transfer_type": "logistics_to_distributor"})


init_demo_data()
print(f"[Init] Demo data loaded. Blockchain height: {len(bc.chain)}")
print(f"[Init] Chain integrity: {'VALID' if bc.verify_chain()[0] else 'BROKEN'}")

# =============================================================================
# 第五部分：REST API
# =============================================================================

@app.route("/api/v1/health")
def health():
    valid, bad_idx = bc.verify_chain()
    return jsonify({
        "status": "ok", "service": "paddle-trace-api", "version": "1.0.0",
        "blockchain_height": len(bc.chain),
        "chain_integrity": "valid" if valid else f"broken at block {bad_idx}"
    })

# ---- 产品 ----
@app.route("/api/v1/products", methods=["POST"])
def register_product():
    body = request.json
    caller = request.headers.get("X-Caller", "0xManufacturer01")
    result = contract_register_product(caller, body)
    if "error" in result:
        return jsonify({"code": 403, "error": result["error"]}), 403
    return jsonify({"code": 201, "message": result["message"]}), 201

@app.route("/api/v1/products/<product_id>")
def get_product(product_id):
    product = contract_query_product({"product_id": product_id})
    history = contract_query_history({"product_id": product_id})
    if not product:
        return jsonify({"code": 404, "error": "product not found"}), 404
    return jsonify({"code": 200, "data": {"product": product, "history": history}})

@app.route("/api/v1/products/<product_id>/transfer", methods=["POST"])
def transfer_product(product_id):
    body = request.json
    body["product_id"] = product_id
    caller = request.headers.get("X-Caller", "0xManufacturer01")
    result = contract_transfer_ownership(caller, body)
    if "error" in result:
        return jsonify({"code": 403, "error": result["error"]}), 403
    return jsonify({"code": 200, "message": result["message"]})

@app.route("/api/v1/products/<product_id>/trace", methods=["POST"])
def append_trace(product_id):
    body = request.json
    body["product_id"] = product_id
    caller = request.headers.get("X-Caller", "0xLogistics01")
    result = contract_append_trace(caller, body)
    if "error" in result:
        return jsonify({"code": 403, "error": result["error"]}), 403
    return jsonify({"code": 201, "message": result["message"]}), 201

# ---- NFC验证 ----
@app.route("/api/v1/products/verify-nfc", methods=["POST"])
def verify_nfc():
    body = request.json
    tag_uid = body.get("tag_uid")
    sun_code = body.get("sun_code")
    counter = body.get("counter")
    product_id = body.get("product_id")

    if not all([tag_uid, sun_code, counter is not None, product_id]):
        return jsonify({"code": 400, "error": "missing required fields"}), 400

    # 步骤1：NFC验证
    ok, msg, reason = nfc.verify_sun(tag_uid, sun_code, int(counter))

    # 步骤2：区块链验证
    chain_result = contract_verify_authenticity({"product_id": product_id})
    chain_ok = chain_result.get("authentic", False)

    # 步骤3：综合判定
    authentic = ok and chain_ok

    if authentic:
        message = "验证通过：该球拍为正品，NFC芯片认证通过，区块链溯源记录完整"
    elif not ok:
        message = f"NFC标签验证失败：{msg}"
    else:
        message = "产品未在区块链注册：该球拍疑为假冒产品"

    return jsonify({"code": 200, "data": {
        "authentic": authentic, "product_id": product_id,
        "nfc_verified": ok, "chain_verified": chain_ok,
        "nfc_detail": msg, "brand": chain_result.get("brand"),
        "model": chain_result.get("model"), "message": message,
    }})

# ---- 管理接口 ----
@app.route("/api/v1/admin/chain")
def admin_chain():
    """查看区块链状态"""
    blocks = []
    for b in bc.chain[-10:]:  # 最近10个区块
        blocks.append({
            "index": b.index, "hash": b.hash[:16],
            "prev_hash": b.previous_hash[:16], "type": b.data.get("type"),
            "timestamp": b.timestamp,
        })
    valid, bad = bc.verify_chain()
    return jsonify({
        "height": len(bc.chain), "integrity": "valid" if valid else f"broken@{bad}",
        "recent_blocks": blocks
    })

@app.route("/api/v1/admin/reset")
def admin_reset():
    """重置演示数据"""
    global bc, nfc
    bc = Blockchain()
    nfc = NFCService()
    nfc.register_tag("04A2B3C4D5E6F7", "00112233445566778899AABBCCDDEEFF")
    nfc.register_tag("04F7E6D5C4B3A2", "FFEEDDCCBBAA99887766554433221100")
    init_demo_data()
    return jsonify({"message": "reset complete", "height": len(bc.chain)})


# =============================================================================
# 启动
# =============================================================================
if __name__ == "__main__":
    print("""
    ==================================================
      Paddle Traceability System v1.0
      Blockchain + NFC + REST API
    ==================================================
      API:      http://localhost:5050
      Health:   http://localhost:5050/api/v1/health
      Frontend: open frontend/consumer/index.html
    ==================================================
      Demo Products:
        pingpong101 (authentic, full trace chain)
        pingpong102 (authentic, short trace chain)
        pingpong104 (counterfeit)
    ==================================================
    """)
    app.run(host="127.0.0.1", port=5050, debug=True)
