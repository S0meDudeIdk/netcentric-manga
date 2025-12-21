# Script to update total_chapters for manga that currently have 0
# This will fetch the data from MAL and update the database

Write-Host "Updating total_chapters for manga with 0 chapters..." -ForegroundColor Cyan

# Change to project root
Set-Location -Path "$PSScriptRoot/.."

# Build and run the update script
Write-Host "Compiling Go update script..." -ForegroundColor Yellow
go run scripts/update-chapters.go

Write-Host "Done!" -ForegroundColor Green
