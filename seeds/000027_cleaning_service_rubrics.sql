-- Panduan penilaian trainee untuk jabatan CLEANING SERVICE.
START TRANSACTION;

INSERT INTO trainee_assessment_aspects(code,name,sort_order) VALUES
('layanan','LAYANAN',1),('kemampuan','KEMAMPUAN',2),('etika','ETIKA',3),
('kerjasama','KERJASAMA',4),('keinginan_belajar','KEINGINAN BELAJAR',5)
ON DUPLICATE KEY UPDATE name=VALUES(name),sort_order=VALUES(sort_order);

INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
SELECT p.department_id,p.id,period.code,CONCAT(p.name,' - ',period.label),period.month_from,period.month_to,1
FROM positions p CROSS JOIN (
 SELECT 'cleaning_service_1_2' code,'Bulan 1-2' label,1 month_from,2 month_to
 UNION ALL SELECT 'cleaning_service_3_4','Bulan 3-4',3,4
 UNION ALL SELECT 'cleaning_service_5_6','Bulan 5-6',5,6
) period
WHERE p.code='cleaning_service'
ON DUPLICATE KEY UPDATE name=VALUES(name),month_from=VALUES(month_from),month_to=VALUES(month_to),is_active=1;

INSERT INTO trainee_assessment_criteria(department_id,position_id,aspect_id,code,description,weight,sort_order)
SELECT p.department_id,p.id,a.id,c.code,c.description,1,c.sort_order
FROM positions p CROSS JOIN (
 SELECT 'layanan' aspect_code,'cleaning_layanan_1' code,'Tidak menjaga kebersihan dan kerapian area kerja serta sarana prasarana kerja dibagiannya' description,1 sort_order
 UNION ALL SELECT 'layanan','cleaning_layanan_2','Tidak segera merespon kejadian darurat/kondisi khusus di area toko',2
 UNION ALL SELECT 'kemampuan','cleaning_kemampuan_1','Tidak melakukan pengecekkan kebersihan area sesuai jadwal',1
 UNION ALL SELECT 'kemampuan','cleaning_kemampuan_2','Tidak menggunakan alat pelindung diri saat bekerja di area risiko',2
 UNION ALL SELECT 'kemampuan','cleaning_kemampuan_3','Tidak segera memperbaiki kondisi kotor/rusak yang berpotensi merusak citra toko dan menimbulkan komplain dari internal dan eksternal',3
 UNION ALL SELECT 'etika','cleaning_etika_1','Sikap & Cara Berbicara tidak sesuai budaya dan peraturan perusahaan',1
 UNION ALL SELECT 'kerjasama','cleaning_kerjasama_1','Jarang berbagi informasi atau membantu rekan kerja saat dibutuhkan',1
 UNION ALL SELECT 'kerjasama','cleaning_kerjasama_2','Koordinasi dengan rekan kerja dan bagian lain sering kurang jelas sehingga pekerjaan terhambat',2
 UNION ALL SELECT 'keinginan_belajar','cleaning_belajar_1','Sering salah memahami arahan kerja dan perlu diingatkan atau dijelaskan kembali',1
 UNION ALL SELECT 'keinginan_belajar','cleaning_belajar_2','Kurang berinisiatif untuk belajar hal baru atau meningkatkan kemampuan kerja',2
) c JOIN trainee_assessment_aspects a ON a.code=c.aspect_code
WHERE p.code='cleaning_service'
ON DUPLICATE KEY UPDATE aspect_id=VALUES(aspect_id),description=VALUES(description),weight=VALUES(weight),sort_order=VALUES(sort_order);

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT c.id,p.id FROM trainee_assessment_criteria c
JOIN positions p ON p.id=c.position_id
WHERE p.code='cleaning_service' AND c.code LIKE 'cleaning_%';

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT r.id,c.id,s.score,s.finding_min,s.finding_max,s.description
FROM trainee_assessment_rubrics r
JOIN positions p ON p.id=r.position_id
JOIN trainee_assessment_criterion_positions cp ON cp.position_id=p.id
JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id AND c.code LIKE 'cleaning_%'
JOIN (
 SELECT 'standard' kind,5 score,0 finding_min,0 finding_max,'tidak ada temuan' description
 UNION ALL SELECT 'standard',4,1,2,'1x - 2x temuan'
 UNION ALL SELECT 'standard',3,3,3,'3x temuan'
 UNION ALL SELECT 'standard',2,4,4,'4x temuan'
 UNION ALL SELECT 'standard',1,5,NULL,'>= 5x temuan'
 UNION ALL SELECT 'early' kind,5,0,2,'< 3x temuan'
 UNION ALL SELECT 'early',4,3,3,'3x temuan'
 UNION ALL SELECT 'early',3,4,4,'4x temuan'
 UNION ALL SELECT 'early',2,5,5,'5x temuan'
 UNION ALL SELECT 'early',1,6,NULL,'>= 6x temuan'
 UNION ALL SELECT 'middle' kind,5,0,1,'< 2x temuan'
 UNION ALL SELECT 'middle',4,2,2,'2x temuan'
 UNION ALL SELECT 'middle',3,3,3,'3x temuan'
 UNION ALL SELECT 'middle',2,4,4,'4x temuan'
 UNION ALL SELECT 'middle',1,5,NULL,'>= 5x temuan'
 UNION ALL SELECT 'late' kind,5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late',4,1,1,'1x temuan'
 UNION ALL SELECT 'late',3,2,2,'2x temuan'
 UNION ALL SELECT 'late',2,3,3,'3x temuan'
 UNION ALL SELECT 'late',1,4,NULL,'>= 4x temuan'
 UNION ALL SELECT 'early_ability' kind,5,0,3,'< 4x temuan'
 UNION ALL SELECT 'early_ability',4,4,4,'4x temuan'
 UNION ALL SELECT 'early_ability',3,5,5,'5x temuan'
 UNION ALL SELECT 'early_ability',2,6,6,'6x temuan'
 UNION ALL SELECT 'early_ability',1,7,NULL,'>= 7x temuan'
 UNION ALL SELECT 'middle_ability' kind,5,0,2,'< 3x temuan'
 UNION ALL SELECT 'middle_ability',4,3,3,'3x temuan'
 UNION ALL SELECT 'middle_ability',3,4,4,'4x temuan'
 UNION ALL SELECT 'middle_ability',2,5,5,'5x temuan'
 UNION ALL SELECT 'middle_ability',1,6,NULL,'>= 6x temuan'
 UNION ALL SELECT 'late_ability' kind,5,0,1,'< 2x temuan'
 UNION ALL SELECT 'late_ability',4,2,2,'2x temuan'
 UNION ALL SELECT 'late_ability',3,3,3,'3x temuan'
 UNION ALL SELECT 'late_ability',2,4,4,'4x temuan'
 UNION ALL SELECT 'late_ability',1,5,NULL,'>= 5x temuan'
 UNION ALL SELECT 'late_social' kind,5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late_social',4,1,1,'1x temuan'
 UNION ALL SELECT 'late_social',3,2,2,'2x temuan'
 UNION ALL SELECT 'late_social',2,3,4,'3x - 4x temuan'
 UNION ALL SELECT 'late_social',1,5,NULL,'>= 5x temuan'
) s ON s.kind=CASE
 WHEN c.code IN ('cleaning_layanan_1','cleaning_layanan_2','cleaning_etika_1') AND r.month_from<5 THEN 'standard'
 WHEN c.code IN ('cleaning_layanan_1','cleaning_layanan_2','cleaning_etika_1') THEN 'late'
 WHEN c.code IN ('cleaning_kemampuan_1','cleaning_kemampuan_2','cleaning_kemampuan_3') AND r.month_from=1 THEN 'early_ability'
 WHEN c.code IN ('cleaning_kemampuan_1','cleaning_kemampuan_2','cleaning_kemampuan_3') AND r.month_from=3 THEN 'middle_ability'
 WHEN c.code IN ('cleaning_kemampuan_1','cleaning_kemampuan_2','cleaning_kemampuan_3') THEN 'late_ability'
 WHEN r.month_from=1 THEN 'early'
 WHEN r.month_from=3 THEN 'middle'
 ELSE 'late_social' END
WHERE p.code='cleaning_service'
  AND r.code IN ('cleaning_service_1_2','cleaning_service_3_4','cleaning_service_5_6')
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);

COMMIT;
