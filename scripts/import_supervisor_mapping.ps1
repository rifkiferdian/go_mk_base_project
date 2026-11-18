param(
    [Parameter(Mandatory = $true)]
    [string]$SourcePath,
    [string]$Database = "hr_trainee_appreciation_assessment",
    [string]$MySqlPath = "C:\xampp8.2.12\mysql\bin\mysql.exe"
)

$ErrorActionPreference = "Stop"

function ConvertTo-SqlLiteral([string]$Value) {
    if ($null -eq $Value) { return "NULL" }
    return "'" + $Value.Replace("\", "\\").Replace("'", "''") + "'"
}

if (-not (Test-Path -LiteralPath $SourcePath)) {
    throw "File sumber tidak ditemukan: $SourcePath"
}

$rows = @(Import-Csv -LiteralPath $SourcePath -Delimiter "`t")
if ($rows.Count -eq 0) {
    throw "File sumber tidak berisi data."
}

$requiredColumns = @("NIK", "NAMA", "JABATAN USULAN", "spv")
foreach ($column in $requiredColumns) {
    if ($column -notin $rows[0].PSObject.Properties.Name) {
        throw "Kolom wajib tidak ditemukan: $column"
    }
}

$values = foreach ($row in $rows) {
    $supervisor = $row.spv.Trim().ToUpperInvariant()
    if ($supervisor -eq "ARFINAI") { $supervisor = "ARFIANI" }

    $employeeNumber = $row.NIK.Trim()
    if ($employeeNumber -eq "HL PP") {
        $normalizedName = ($row.NAMA -replace " \(\d+\)$", "").Trim().ToUpperInvariant()
        $employeeNumber = switch ($normalizedName) {
            "SUWARDI" { "960012" }
            "MUJIYONO" { "960007" }
            "SITTI ARIFAH" { "960008" }
            "DAHRI AHMAD" { "910001" }
            "KUPON" { "HLPP-KUPON" }
            "YUNAN SETIYAWARNI" { "HLPP-YUNAN-SETIYAWARNI" }
            "HARIS SERA PASARUA" { "HLPP-HARIS-SERA-PASARUA" }
            "PRASETYO BUDI SANTOSO" { "HLPP-PRASETYO-BUDI-SANTOSO" }
            "SUGENG RETNO" { "HLPP-SUGENG-RETNO" }
            default { throw "Tidak dapat menentukan NIK HLPP untuk $($row.NAMA)." }
        }
    }

    "(" + ((@(
        $employeeNumber,
        $row.NAMA.Trim(),
        $row.'JABATAN USULAN'.Trim(),
        $supervisor
    ) | ForEach-Object { ConvertTo-SqlLiteral $_ }) -join ",") + ")"
}

$valueSql = $values -join ",`n"
$sql = @"
SET NAMES utf8mb4;
START TRANSACTION;

CREATE TEMPORARY TABLE tmp_supervisor_mapping (
    employee_number VARCHAR(50) NOT NULL PRIMARY KEY,
    employee_name VARCHAR(150) NOT NULL,
    proposed_position VARCHAR(150) NOT NULL,
    supervisor_name VARCHAR(150) NOT NULL
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

INSERT INTO tmp_supervisor_mapping
    (employee_number, employee_name, proposed_position, supervisor_name)
VALUES
$valueSql;

CREATE TEMPORARY TABLE tmp_mapping_validation AS
SELECT
    m.employee_number,
    m.employee_name,
    m.proposed_position,
    m.supervisor_name,
    child.id AS child_employee_id,
    child_assignment.id AS child_assignment_id,
    child_assignment.store_id,
    child_assignment.started_at,
    supervisor.id AS supervisor_employee_id,
    supervisor.employee_number AS supervisor_number,
    position.name AS actual_position
FROM tmp_supervisor_mapping m
LEFT JOIN employees child ON child.employee_number = m.employee_number
LEFT JOIN employee_assignments child_assignment
    ON child_assignment.employee_id = child.id AND child_assignment.ended_at IS NULL
LEFT JOIN positions position ON position.id = child_assignment.position_id
LEFT JOIN employees supervisor
    ON UPPER(supervisor.name) COLLATE utf8mb4_unicode_ci = m.supervisor_name COLLATE utf8mb4_unicode_ci
    AND supervisor.employee_number LIKE 'SPV%';

SET @invalid_count = (
    SELECT COUNT(*)
    FROM tmp_mapping_validation
    WHERE child_employee_id IS NULL
       OR child_assignment_id IS NULL
       OR supervisor_employee_id IS NULL
       OR actual_position COLLATE utf8mb4_unicode_ci <> proposed_position COLLATE utf8mb4_unicode_ci
);

SELECT IF(@invalid_count = 0, 'VALIDATION_OK', CONCAT('VALIDATION_FAILED:', @invalid_count)) AS validation;

CREATE TEMPORARY TABLE tmp_supervisor_plan AS
SELECT
    ranked.supervisor_employee_id,
    ranked.supervisor_name,
    ranked.store_id,
    ranked.first_child_start,
    CASE
        WHEN ranked.supervisor_name IN ('RIRI', 'MARULI') THEN 'SPV SATPAM'
        WHEN ranked.supervisor_name IN ('HARYANTO', 'MISKAN', 'SIGIT') THEN 'SPV JASA'
        WHEN ranked.supervisor_name IN ('GUSTOMI', 'PRAMONO', 'LIA') THEN 'SPV GUDANG'
        WHEN ranked.supervisor_name = 'EKY' THEN 'SPV AKUNTANSI'
        ELSE 'SPV STORE'
    END AS supervisor_position
FROM (
    SELECT
        v.supervisor_employee_id,
        v.supervisor_name,
        v.store_id,
        MIN(v.started_at) OVER (PARTITION BY v.supervisor_employee_id) AS first_child_start,
        ROW_NUMBER() OVER (
            PARTITION BY v.supervisor_employee_id
            ORDER BY COUNT(*) DESC, v.store_id ASC
        ) AS store_rank
    FROM tmp_mapping_validation v
    WHERE @invalid_count = 0
    GROUP BY v.supervisor_employee_id, v.supervisor_name, v.store_id
) ranked
WHERE ranked.store_rank = 1;

INSERT INTO employee_assignments
    (employee_id, position_id, department_id, store_id, started_at, ended_at)
SELECT
    plan.supervisor_employee_id,
    position.id,
    position.department_id,
    plan.store_id,
    plan.first_child_start,
    NULL
FROM tmp_supervisor_plan plan
JOIN positions position
    ON position.name COLLATE utf8mb4_unicode_ci = plan.supervisor_position COLLATE utf8mb4_unicode_ci
JOIN departments department
    ON department.id = position.department_id AND department.name = 'MANAGEMENT'
LEFT JOIN employee_assignments current_assignment
    ON current_assignment.employee_id = plan.supervisor_employee_id
    AND current_assignment.ended_at IS NULL
WHERE current_assignment.id IS NULL
  AND @invalid_count = 0;

INSERT INTO employee_reporting_lines
    (employee_assignment_id, supervisor_assignment_id, started_at, ended_at)
SELECT
    validation.child_assignment_id,
    supervisor_assignment.id,
    validation.started_at,
    NULL
FROM tmp_mapping_validation validation
JOIN employee_assignments supervisor_assignment
    ON supervisor_assignment.employee_id = validation.supervisor_employee_id
    AND supervisor_assignment.ended_at IS NULL
LEFT JOIN employee_reporting_lines current_line
    ON current_line.employee_assignment_id = validation.child_assignment_id
    AND current_line.ended_at IS NULL
WHERE current_line.id IS NULL
  AND @invalid_count = 0;

SELECT
    (SELECT COUNT(*) FROM tmp_supervisor_mapping) AS source_rows,
    (SELECT COUNT(*) FROM tmp_mapping_validation WHERE child_assignment_id IS NOT NULL) AS matched_subordinates,
    (SELECT COUNT(*) FROM tmp_supervisor_plan) AS planned_supervisors,
    (SELECT COUNT(*) FROM employee_assignments a JOIN employees e ON e.id = a.employee_id WHERE e.employee_number LIKE 'SPV%' AND a.ended_at IS NULL) AS active_supervisor_assignments,
    (SELECT COUNT(*) FROM employee_reporting_lines WHERE ended_at IS NULL) AS active_reporting_lines;

COMMIT;
"@

$output = $sql | & $MySqlPath --user=root --host=127.0.0.1 --port=3306 --batch --raw $Database 2>&1
if ($LASTEXITCODE -ne 0) {
    throw ($output -join [Environment]::NewLine)
}

$output
