# Quick check of publication years in database
$dbPath = "x:\Bao2023toPresent\IU\4th year\NetCentric\Project\netcentric-manga\mangahub\data\mangahub.db"

Write-Host "Checking publication_year values in database..." -ForegroundColor Cyan

# Check if System.Data.SQLite is available
try {
    Add-Type -AssemblyName System.Data.SQLite -ErrorAction Stop
    
    $connection = New-Object System.Data.SQLite.SQLiteConnection("Data Source=$dbPath")
    $connection.Open()
    
    # Query for manga with publication years
    $command = $connection.CreateCommand()
    $command.CommandText = @"
        SELECT id, title, publication_year 
        FROM manga 
        WHERE publication_year IS NOT NULL AND publication_year > 0
        LIMIT 20
"@
    
    $reader = $command.ExecuteReader()
    
    Write-Host ""
    Write-Host "Manga with publication years:" -ForegroundColor Green
    Write-Host "=" * 80
    
    $count = 0
    while ($reader.Read()) {
        $count++
        $id = $reader["id"]
        $title = $reader["title"]
        $year = $reader["publication_year"]
        Write-Host "$count. $title ($year)" -ForegroundColor White
    }
    
    $reader.Close()
    
    # Count total manga with years
    $command = $connection.CreateCommand()
    $command.CommandText = "SELECT COUNT(*) as count FROM manga WHERE publication_year IS NOT NULL AND publication_year > 0"
    $totalWithYears = [int]$command.ExecuteScalar()
    
    # Count total manga
    $command.CommandText = "SELECT COUNT(*) as count FROM manga"
    $totalManga = [int]$command.ExecuteScalar()
    
    # Count NULL years
    $nullYears = $totalManga - $totalWithYears
    
    $connection.Close()
    
    Write-Host ""
    Write-Host "=" * 80
    Write-Host "Summary:" -ForegroundColor Cyan
    Write-Host "  Total manga: $totalManga"
    Write-Host "  With years: $totalWithYears" -ForegroundColor Green
    Write-Host "  NULL years: $nullYears" -ForegroundColor $(if ($nullYears -eq 0) { "Green" } else { "Yellow" })
    Write-Host ""
    
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative: Open DB Browser for SQLite and run this query:" -ForegroundColor Yellow
    Write-Host "SELECT id, title, publication_year FROM manga WHERE publication_year > 0 LIMIT 20;" -ForegroundColor Cyan
}
