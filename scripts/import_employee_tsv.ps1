param(
    [Parameter(Mandatory=$true)][string]$SourcePath,
    [string]$EnvPath = '.env'
)

$ErrorActionPreference = 'Stop'

function SqlText([string]$Value) {
    return "'" + $Value.Replace("'", "''") + "'"
}

function EmployeeNumber([string]$Nik, [string]$Name) {
    if ($Nik -ne 'HL PP') { return $Nik }
    if ($Name -match '\((\d{6})\)') { return $Matches[1] }
    $slug = ($Name.ToUpperInvariant() -replace '[^A-Z0-9]+','-').Trim('-')
    return ('HLPP-' + $slug).Substring(0, [Math]::Min(50, 5 + $slug.Length))
}

function CleanName([string]$Name) {
    return ($Name -replace '\s*\(\d{6}\)\s*$','' -replace '\s*\(MK\s*8\)\s*$','').Trim()
}

$config = @{}
Get-Content -LiteralPath $EnvPath | ForEach-Object {
    if ($_ -match '^([^#=]+)=(.*)$') {
        $config[$Matches[1].Trim()] = $Matches[2].Trim().Trim('"').Trim("'")
    }
}

$rows = Get-Content -Encoding UTF8 -LiteralPath $SourcePath |
    Select-Object -Skip 1 |
    Where-Object { $_.Trim() } |
    ForEach-Object {
        $columns = $_ -split "`t"
        if ($columns.Count -ne 6) { throw "Baris TSV tidak valid: $_" }
        $period = $columns[5].Trim()
        if ($period -notmatch '^(\d{2})/(\d{2})/(\d{4})\s*-\s*(\d{2})/(\d{2})/(\d{4})$') {
            throw "Periode tidak valid: $period"
        }
        $start = "$($Matches[3])-$($Matches[2])-$($Matches[1])"
        $end = "$($Matches[6])-$($Matches[5])-$($Matches[4])"
        $number = EmployeeNumber ($columns[0].Trim()) ($columns[1].Trim())
        if ($number -eq '260304' -and $end -eq '2026-04-30') { $end = '2027-04-30' }
        if ([datetime]$end -lt [datetime]$start) { throw "Periode berakhir sebelum mulai untuk NIK $number" }
        [pscustomobject]@{
            EmployeeNumber = $number
            Name = CleanName ($columns[1].Trim())
            StoreCode = 'MK' + $columns[2].Trim()
            DepartmentName = $columns[3].Trim()
            PositionName = $columns[4].Trim()
            StartedAt = $start
            EmploymentEndedAt = $end
            EmploymentStatus = if ($columns[0].Trim() -eq 'HL PP') { 'post_retirement_daily_worker' } else { 'trainee' }
        }
    }

$duplicates = $rows | Group-Object EmployeeNumber | Where-Object Count -gt 1
if ($duplicates) { throw "NIK hasil normalisasi tidak unik: $($duplicates.Name -join ', ')" }

$values = $rows | ForEach-Object {
    '(' + (@(
        SqlText $_.EmployeeNumber
        SqlText $_.Name
        SqlText $_.StoreCode
        SqlText $_.DepartmentName
        SqlText $_.PositionName
        SqlText $_.StartedAt
        SqlText $_.EmploymentEndedAt
        SqlText $_.EmploymentStatus
    ) -join ',') + ')'
}

$today = Get-Date -Format 'yyyy-MM-dd'
$sql = @"
START TRANSACTION;
CREATE TEMPORARY TABLE import_employees (
 employee_number VARCHAR(50) NOT NULL PRIMARY KEY,
 employee_name VARCHAR(150) NOT NULL,
 store_code VARCHAR(50) NOT NULL,
 department_name VARCHAR(150) NOT NULL,
 position_name VARCHAR(150) NOT NULL,
 started_at DATE NOT NULL,
 employment_ended_at DATE NOT NULL,
 employment_status VARCHAR(50) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
INSERT INTO import_employees VALUES
$($values -join ",`n");

INSERT INTO employees(employee_number,name,status)
SELECT employee_number,employee_name,IF(employment_ended_at<'$today','inactive','active')
FROM import_employees
ON DUPLICATE KEY UPDATE name=VALUES(name),status=VALUES(status);

INSERT INTO employee_profiles(employee_id,joined_date,employment_ended_at,employment_status)
SELECT e.id,i.started_at,i.employment_ended_at,i.employment_status
FROM import_employees i JOIN employees e ON e.employee_number=i.employee_number
ON DUPLICATE KEY UPDATE joined_date=VALUES(joined_date),employment_ended_at=VALUES(employment_ended_at),employment_status=VALUES(employment_status);

INSERT INTO employee_assignments(employee_id,position_id,department_id,store_id,started_at,ended_at)
SELECT e.id,p.id,d.id,s.id,i.started_at,IF(i.employment_ended_at<'$today',i.employment_ended_at,NULL)
FROM import_employees i
JOIN employees e ON e.employee_number=i.employee_number
JOIN departments d ON d.name=i.department_name
JOIN positions p ON p.department_id=d.id AND p.name=i.position_name
JOIN stores s ON s.code=i.store_code AND s.deleted_at IS NULL
ON DUPLICATE KEY UPDATE position_id=VALUES(position_id),department_id=VALUES(department_id),store_id=VALUES(store_id),ended_at=VALUES(ended_at);

DROP TEMPORARY TABLE import_employees;
COMMIT;
"@

$mysql = 'C:\xampp8.2.12\mysql\bin\mysql.exe'
$env:MYSQL_PWD = $config['DB_PASSWORD']
try {
    $sql | & $mysql -h $config['DB_HOST'] -P $config['DB_PORT'] -u $config['DB_USER'] --default-character-set=utf8mb4 $config['DB_NAME']
    if ($LASTEXITCODE -ne 0) { throw 'Impor database gagal.' }
} finally {
    Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
}

Write-Output "IMPORTED=$($rows.Count)"
