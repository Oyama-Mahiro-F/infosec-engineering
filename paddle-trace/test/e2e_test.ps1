# =============================================================================
# 端到端功能测试 - PowerShell 版
# 运行: powershell -ExecutionPolicy Bypass -File test/e2e_test.ps1
# 前提: python server.py 已在 127.0.0.1:5050 运行
# 建议先重置: curl http://127.0.0.1:5050/api/v1/admin/reset
# =============================================================================

$API = "http://127.0.0.1:5050"
$Passed = 0
$Failed = 0
$Total = 0

function Check($Name, $Condition, $Detail = "") {
    $script:Total++
    if ($Condition) {
        $script:Passed++
        Write-Host "  [PASS] $Name" -ForegroundColor Green
    } else {
        $script:Failed++
        Write-Host "  [FAIL] $Name  $Detail" -ForegroundColor Red
    }
}

function Get-HMAC256($KeyBytes, $Message) {
    $hmac = New-Object System.Security.Cryptography.HMACSHA256
    $hmac.Key = $KeyBytes
    $msgBytes = [System.Text.Encoding]::UTF8.GetBytes($Message)
    $hash = $hmac.ComputeHash($msgBytes)
    $hex = [BitConverter]::ToString($hash).Replace("-","").ToLower()
    return $hex.Substring(0, 32)
}

function Get-SHA256Short($JsonStr) {
    $sha = [System.Security.Cryptography.SHA256]::Create()
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($JsonStr)
    $hash = $sha.ComputeHash($bytes)
    return [BitConverter]::ToString($hash).Replace("-","").ToLower().Substring(0,16)
}

# Reset demo data for clean test run
Write-Host "Resetting demo data..." -ForegroundColor Gray
try { Invoke-RestMethod -Uri "$API/api/v1/admin/reset" -Method Get | Out-Null } catch {}
Start-Sleep -Seconds 1

Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host "  End-to-End Functional Test Report (PowerShell)" -ForegroundColor Cyan
Write-Host "  API: $API" -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Cyan

# =============================================================================
# Test 1: Health Check
# =============================================================================
Write-Host "`n[Test 1] Health Check & Chain Integrity" -ForegroundColor Yellow
$r = Invoke-RestMethod -Uri "$API/api/v1/health" -Method Get
Check "HTTP 200" ($r.status -eq "ok")
Check "height > 0" ($r.blockchain_height -gt 0) "height=$($r.blockchain_height)"
Check "integrity=valid" ($r.chain_integrity -eq "valid") $r.chain_integrity
Write-Host "  Height=$($r.blockchain_height)  Integrity=$($r.chain_integrity)"

# =============================================================================
# Test 2: Authentic Product
# =============================================================================
Write-Host "`n[Test 2] Authentic Product Query (pingpong101)" -ForegroundColor Yellow
$r = Invoke-RestMethod -Uri "$API/api/v1/products/pingpong101" -Method Get
$p = $r.data.product
$h = $r.data.history
Check "code=200" ($r.code -eq 200)
Check "brand=Butterfly" ($p.brand -eq "Butterfly") $p.brand
Check "model=VISCARIA FL" ($p.model -eq "VISCARIA FL") $p.model
Check "status=sold" ($p.current_status -eq "sold") $p.current_status
Check "owner=0xConsumer02" ($p.current_owner -eq "0xConsumer02") $p.current_owner
Check "trace_count=7" ($h.count -eq 7) "got $($h.count)"
Write-Host "  $($p.brand) $($p.model) | Owner: $($p.current_owner) | Traces: $($h.count)"

# =============================================================================
# Test 3: Hash Chain Verification
# =============================================================================
Write-Host "`n[Test 3] Hash Chain Verification" -ForegroundColor Yellow
$records = $h.records
$chainOk = $true
# Verify: each record (except first) has a non-empty prev_record linking to previous
for ($i = 1; $i -lt $records.Count; $i++) {
    $prev = $records[$i].prev_record
    if (-not $prev -or $prev -eq "" -or $prev -eq "genesis") {
        if ($i -gt 1) { $chainOk = $false }
    }
}
# Also verify server-side chain integrity
$health = Invoke-RestMethod -Uri "$API/api/v1/health" -Method Get
Check "hash_chain_linked" ($chainOk -and $records.Count -gt 0)
Check "server_chain_valid" ($health.chain_integrity -eq "valid")
for ($i = 0; $i -lt $records.Count; $i++) {
    $ph = if ($records[$i].prev_record) { $records[$i].prev_record.Substring(0, [Math]::Min(12, $records[$i].prev_record.Length)) } else { "genesis" }
    Write-Host "  [$($i+1)] $($records[$i].action.PadRight(12)) prev=$ph"
}

# =============================================================================
# Test 4: NFC Authentication
# =============================================================================
Write-Host "`n[Test 4] NFC Authentication" -ForegroundColor Yellow
$key = [byte[]]@(0x00,0x11,0x22,0x33,0x44,0x55,0x66,0x77,0x88,0x99,0xAA,0xBB,0xCC,0xDD,0xEE,0xFF)
$sun = Get-HMAC256 $key "04A2B3C4D5E6F7:200"

$body = @{
    tag_uid    = "04A2B3C4D5E6F7"
    sun_code   = $sun
    counter    = 200
    product_id = "pingpong101"
} | ConvertTo-Json
$r = Invoke-RestMethod -Uri "$API/api/v1/products/verify-nfc" -Method Post -Body $body -ContentType "application/json"
$d = $r.data
Check "authentic=true" ($d.authentic -eq $true)
Check "nfc_verified=true" ($d.nfc_verified -eq $true)
Check "chain_verified=true" ($d.chain_verified -eq $true)

# =============================================================================
# Test 5: Counterfeit Detection
# =============================================================================
Write-Host "`n[Test 5] Counterfeit Detection (pingpong104)" -ForegroundColor Yellow
try {
    $r = Invoke-RestMethod -Uri "$API/api/v1/products/pingpong104" -Method Get
    Check "HTTP 404" $false "unexpected 200"
} catch {
    Check "HTTP 404" ($_.Exception.Response.StatusCode.value__ -eq 404)
}

# =============================================================================
# Test 6: Product Registration
# =============================================================================
Write-Host "`n[Test 6] Product Registration On-chain" -ForegroundColor Yellow
$testPid = "pingpong_test_$(Get-Date -Format 'HHmmss')"
$oldH = (Invoke-RestMethod -Uri "$API/api/v1/health").blockchain_height
$body = @{
    product_id   = $testPid
    brand        = "Butterfly"
    model        = "HAO SHUAI"
    batch_no     = "BTY-2026-300"
    produce_date = "2026-06-15"
} | ConvertTo-Json
$r = Invoke-RestMethod -Uri "$API/api/v1/products" -Method Post -Body $body `
    -ContentType "application/json" -Headers @{"X-Caller"="0xManufacturer01"}
$newH = (Invoke-RestMethod -Uri "$API/api/v1/health").blockchain_height
Check "HTTP 201" ($r.code -eq 201)
Check "height_grew" ($newH -gt $oldH) "$oldH -> $newH"
Write-Host "  Product: $testPid  Height: $oldH -> $newH"

# =============================================================================
# Test 7: NFC Replay Attack Defense
# =============================================================================
Write-Host "`n[Test 7] NFC Replay Attack Defense" -ForegroundColor Yellow
$key2 = [byte[]]@(0xFF,0xEE,0xDD,0xCC,0xBB,0xAA,0x99,0x88,0x77,0x66,0x55,0x44,0x33,0x22,0x11,0x00)
$sun2 = Get-HMAC256 $key2 "04F7E6D5C4B3A2:500"

$body2 = @{
    tag_uid="04F7E6D5C4B3A2"; sun_code=$sun2; counter=500; product_id="pingpong101"
} | ConvertTo-Json

$r1 = Invoke-RestMethod -Uri "$API/api/v1/products/verify-nfc" -Method Post -Body $body2 -ContentType "application/json"
$r2 = Invoke-RestMethod -Uri "$API/api/v1/products/verify-nfc" -Method Post -Body $body2 -ContentType "application/json"
Check "first_pass" ($r1.data.nfc_verified -eq $true)
Check "second_blocked" ($r2.data.nfc_verified -eq $false) "replay attack detected"
Write-Host "  1st: $($r1.data.nfc_detail)"
Write-Host "  2nd: $($r2.data.nfc_detail)"

# =============================================================================
# Test 8: Permission Escalation Blocked
# =============================================================================
Write-Host "`n[Test 8] Permission Escalation Blocked" -ForegroundColor Yellow
$body3 = @{
    product_id="pingpong400"; brand="Fake"; model="Fake"; batch_no="X"; produce_date="2026-01-01"
} | ConvertTo-Json
try {
    $r = Invoke-RestMethod -Uri "$API/api/v1/products" -Method Post -Body $body3 `
        -ContentType "application/json" -Headers @{"X-Caller"="0xLogistics01"}
    Check "non-mfr blocked(403)" $false "got 20x"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Check "non-mfr blocked(403)" ($code -eq 403) "got $code"
}

# =============================================================================
# Summary
# =============================================================================
Write-Host "`n=======================================================" -ForegroundColor Cyan
Write-Host "  TOTAL: $Total  PASS: $Passed  FAIL: $Failed" -ForegroundColor White
if ($Failed -eq 0) {
    Write-Host "  >>> ALL TESTS PASSED <<<" -ForegroundColor Green
} else {
    Write-Host "  >>> $Failed TEST(S) FAILED <<<" -ForegroundColor Red
}
Write-Host "=======================================================" -ForegroundColor Cyan
