START TRANSACTION;
INSERT INTO trainee_assessment_rubrics(department_id,code,name,month_from,month_to)
SELECT d.id,p.code,p.name,p.month_from,p.month_to FROM departments d CROSS JOIN (
 SELECT 'gudang_1_2' code,'Gudang — Bulan 1–2' name,1 month_from,2 month_to
 UNION ALL SELECT 'gudang_3_4','Gudang — Bulan 3–4',3,4
 UNION ALL SELECT 'gudang_5_plus','Gudang — Bulan 5 dan seterusnya',5,NULL
) p WHERE d.code='gudang'
ON DUPLICATE KEY UPDATE code=VALUES(code);
COMMIT;
