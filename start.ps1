# start.ps1 - One Command to Run Full Vehicle Commerce System (Supabase Cloud DB)
$ErrorActionPreference = "Continue"

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "   Vehicle E-commerce & Multi-Branch Management System  " -ForegroundColor Yellow
Write-Host "         STARTING SERVICES (Supabase Powered)           " -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host ""

# 1. Build and start the Go Backend API from the current source
Write-Host "[1/2] Starting Go Backend API on port 8080..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$PSScriptRoot\backend'; go build -o api.exe ./cmd/api; if (`$LASTEXITCODE -eq 0) { Write-Host '🚀 Go Backend API (Port 8080)' -ForegroundColor Green; .\api.exe } else { Write-Host 'Backend build failed; close any running api.exe and retry.' -ForegroundColor Red }"

# 2. Start Next.js Frontend
Write-Host "[2/2] Starting Next.js Frontend on port 3000..." -ForegroundColor Cyan
$frontendPath = Join-Path $PSScriptRoot "frontend"
if (-not (Test-Path (Join-Path $frontendPath "package.json"))) {
    Write-Host "Frontend files are missing from the frontend folder." -ForegroundColor Red
    exit 1
}
if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
    Write-Host "Node.js and npm are required to start the frontend. Install Node.js, then reopen this terminal." -ForegroundColor Red
    exit 1
}
if (-not (Test-Path (Join-Path $frontendPath "node_modules"))) {
    Push-Location $frontendPath
    npm install
    $installExitCode = $LASTEXITCODE
    Pop-Location
    if ($installExitCode -ne 0) {
        Write-Host "Frontend dependency installation failed." -ForegroundColor Red
        exit $installExitCode
    }
}
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$frontendPath'; Write-Host '⚡ Next.js Frontend (Port 3000)' -ForegroundColor Cyan; npm run dev"

# Wait and open browser
Start-Sleep -Seconds 3
Start-Process "http://localhost:3000"

Write-Host ""
Write-Host "========================================================" -ForegroundColor Green
Write-Host "   ALL SERVICES ARE NOW RUNNING! (NO DOCKER NEEDED)     " -ForegroundColor Green
Write-Host "   - Customer Showroom & Admin: http://localhost:3000   " -ForegroundColor White
Write-Host "   - Go Backend REST API:       http://localhost:8080/api/v1" -ForegroundColor White
Write-Host "   - Cloud Database:            Supabase (PostgreSQL)   " -ForegroundColor White
Write-Host "========================================================" -ForegroundColor Green
Write-Host "   To stop all services anytime, run: .\stop.ps1        " -ForegroundColor Yellow
Write-Host ""
