# Script to update publication_year for all manga from MangaDex API
# This script fetches the year from MangaDex and updates the database

$ErrorActionPreference = "Stop"

# Database path
$dbPath = "$PSScriptRoot\..\data\manga.db"

# Check if database exists
if (-not (Test-Path $dbPath)) {
    Write-Error "Database not found at $dbPath"
    exit 1
}

Write-Host "Connecting to database..." -ForegroundColor Cyan

# Load SQLite assembly
Add-Type -Path "System.Data.SQLite.dll" -ErrorAction SilentlyContinue

# Function to execute SQLite query
function Invoke-SQLiteQuery {
    param(
        [string]$Query,
        [hashtable]$Parameters = @{}
    )
    
    $connection = New-Object System.Data.SQLite.SQLiteConnection
    $connection.ConnectionString = "Data Source=$dbPath"
    $connection.Open()
    
    try {
        $command = $connection.CreateCommand()
        $command.CommandText = $Query
        
        foreach ($key in $Parameters.Keys) {
            $null = $command.Parameters.AddWithValue($key, $Parameters[$key])
        }
        
        $adapter = New-Object System.Data.SQLite.SQLiteDataAdapter($command)
        $dataSet = New-Object System.Data.DataSet
        $null = $adapter.Fill($dataSet)
        
        return $dataSet.Tables[0]
    }
    finally {
        $connection.Close()
    }
}

# Function to execute non-query SQLite command
function Invoke-SQLiteNonQuery {
    param(
        [string]$Query,
        [hashtable]$Parameters = @{}
    )
    
    $connection = New-Object System.Data.SQLite.SQLiteConnection
    $connection.ConnectionString = "Data Source=$dbPath"
    $connection.Open()
    
    try {
        $command = $connection.CreateCommand()
        $command.CommandText = $Query
        
        foreach ($key in $Parameters.Keys) {
            $null = $command.Parameters.AddWithValue($key, $Parameters[$key])
        }
        
        return $command.ExecuteNonQuery()
    }
    finally {
        $connection.Close()
    }
}

# Get all manga with MangaDex IDs that have NULL publication_year
Write-Host "Fetching manga with NULL publication_year..." -ForegroundColor Cyan
$mangaList = Invoke-SQLiteQuery -Query @"
    SELECT id, title FROM manga 
    WHERE id LIKE 'md-%' 
    AND (publication_year IS NULL OR publication_year = 0)
    ORDER BY title
"@

$total = $mangaList.Rows.Count
Write-Host "Found $total manga with NULL publication_year" -ForegroundColor Yellow

if ($total -eq 0) {
    Write-Host "All manga already have publication years!" -ForegroundColor Green
    exit 0
}

$updated = 0
$failed = 0
$notFound = 0
$counter = 0

foreach ($row in $mangaList.Rows) {
    $counter++
    $mangaId = $row["id"]
    $title = $row["title"]
    
    # Extract MangaDex UUID from our ID format (md-<uuid>)
    $mdUuid = $mangaId.Substring(3)
    
    Write-Host "[$counter/$total] Processing: $title" -ForegroundColor Gray
    
    try {
        # Fetch from MangaDex API
        $url = "https://api.mangadex.org/manga/$mdUuid"
        $response = Invoke-RestMethod -Uri $url -Method Get -ErrorAction Stop
        
        if ($response.result -eq "ok" -and $response.data.attributes.year) {
            $year = $response.data.attributes.year
            Write-Host "  Found year: $year" -ForegroundColor Green
            
            # Update database
            $rowsAffected = Invoke-SQLiteNonQuery -Query "UPDATE manga SET publication_year = @year WHERE id = @id" -Parameters @{
                "@year" = $year
                "@id" = $mangaId
            }
            
            if ($rowsAffected -gt 0) {
                $updated++
                Write-Host "  ✓ Updated successfully" -ForegroundColor Green
            } else {
                Write-Host "  ! No rows updated" -ForegroundColor Yellow
            }
        } else {
            $notFound++
            Write-Host "  - No year available from API" -ForegroundColor Yellow
        }
        
        # Rate limiting - MangaDex allows 5 requests per second
        Start-Sleep -Milliseconds 250
        
    } catch {
        $failed++
        Write-Host "  ✗ Error: $_" -ForegroundColor Red
        
        # If rate limited, wait longer
        if ($_.Exception.Message -match "429|rate limit") {
            Write-Host "  Rate limited, waiting 60 seconds..." -ForegroundColor Yellow
            Start-Sleep -Seconds 60
        }
    }
    
    # Progress update every 50 items
    if ($counter % 50 -eq 0) {
        Write-Host ""
        Write-Host "Progress: $counter/$total processed" -ForegroundColor Cyan
        Write-Host "  Updated: $updated" -ForegroundColor Green
        Write-Host "  Not found: $notFound" -ForegroundColor Yellow
        Write-Host "  Failed: $failed" -ForegroundColor Red
        Write-Host ""
    }
}

Write-Host ""
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Update Complete!" -ForegroundColor Green
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Total processed: $total"
Write-Host "Successfully updated: $updated" -ForegroundColor Green
Write-Host "Year not available: $notFound" -ForegroundColor Yellow
Write-Host "Failed: $failed" -ForegroundColor Red
Write-Host ""

if ($updated -gt 0) {
    Write-Host "✓ $updated manga now have publication years!" -ForegroundColor Green
}
