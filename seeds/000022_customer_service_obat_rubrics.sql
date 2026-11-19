-- Panduan penilaian trainee Customer Service dan Obat pada bagian OPERATIONAL.
-- Kriteria dipakai bersama; batas temuan dibedakan per periode rubrik.
START TRANSACTION;

INSERT INTO trainee_assessment_aspects(code,name,sort_order) VALUES
('layanan','LAYANAN',1),
('kemampuan','KEMAMPUAN',2),
('etika','ETIKA',3),
('kerjasama','KERJASAMA',4),
('keinginan_belajar','KEINGINAN BELAJAR',5)
ON DUPLICATE KEY UPDATE name=VALUES(name),sort_order=VALUES(sort_order);

INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
SELECT d.id,j.id,p.code,p.name,p.month_from,p.month_to,1
FROM departments d JOIN positions j ON j.department_id=d.id AND j.code='customer_service' CROSS JOIN (
 SELECT 'cs_obat_1_2' code,'Customer Service — Bulan 1–2' name,1 month_from,2 month_to
 UNION ALL SELECT 'cs_obat_3_4','Customer Service — Bulan 3–4',3,4
 UNION ALL SELECT 'cs_obat_5_6','Customer Service — Bulan 5–6',5,6
) p
WHERE d.code='operational'
ON DUPLICATE KEY UPDATE name=VALUES(name),month_from=VALUES(month_from),month_to=VALUES(month_to),is_active=1;

INSERT INTO trainee_assessment_criteria(department_id,position_id,aspect_id,code,description,weight,sort_order)
SELECT d.id,j.id,a.id,c.code,c.description,1,c.sort_order
FROM departments d
JOIN positions j ON j.department_id=d.id AND j.code='customer_service'
CROSS JOIN (
 SELECT 'layanan' aspect_code,'cs_layanan_1' code,'Tidak menjaga kebersihan dan kerapian barang, area kerja, serta sarana prasarana kerja dibagiannya' description,1 sort_order
 UNION ALL SELECT 'layanan','cs_layanan_2','Tidak sigap dalam melakukan tanggung jawab pekerjaan dan atau tidak ramah dalam melayani konsumen',2
 UNION ALL SELECT 'kemampuan','cs_kemampuan_1','Tidak menjalankan mekanisme SOP audio promo, pemutaran lagu, dan greeting customer service',1
 UNION ALL SELECT 'kemampuan','cs_kemampuan_2','Ketidaksesuaian stock hadiah (terjadi out of stock hadiah member)',2
 UNION ALL SELECT 'kemampuan','cs_kemampuan_3','Tidak memahami alur administrasi komplain',3
 UNION ALL SELECT 'etika','cs_etika_1','Sikap & Cara Berbicara tidak sesuai budaya dan peraturan perusahaan',1
 UNION ALL SELECT 'kerjasama','cs_kerjasama_1','Jarang berbagi informasi atau membantu rekan kerja saat dibutuhkan',1
 UNION ALL SELECT 'kerjasama','cs_kerjasama_2','Koordinasi dengan rekan kerja dan bagian lain sering kurang jelas sehingga pekerjaan terhambat',2
 UNION ALL SELECT 'keinginan_belajar','cs_belajar_1','Sering salah memahami arahan kerja dan perlu diingatkan atau dijelaskan kembali',1
 UNION ALL SELECT 'keinginan_belajar','cs_belajar_2','Kurang berinisiatif untuk belajar hal baru atau meningkatkan kemampuan kerja',2
) c
JOIN trainee_assessment_aspects a ON a.code=c.aspect_code
WHERE d.code='operational'
ON DUPLICATE KEY UPDATE aspect_id=VALUES(aspect_id),description=VALUES(description),weight=VALUES(weight),sort_order=VALUES(sort_order);

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT r.id,c.id,s.score,s.finding_min,s.finding_max,s.description
FROM trainee_assessment_rubrics r
JOIN positions p ON p.id=r.position_id AND p.code='customer_service'
JOIN trainee_assessment_criteria c ON c.position_id=p.id AND c.code LIKE 'cs_%'
JOIN (
 SELECT 'standard' kind,5 score,0 finding_min,0 finding_max,'tidak ada temuan' description
 UNION ALL SELECT 'standard',4,1,2,'1x - 2x temuan'
 UNION ALL SELECT 'standard',3,3,3,'3x temuan'
 UNION ALL SELECT 'standard',2,4,4,'4x temuan'
 UNION ALL SELECT 'standard',1,5,NULL,'≥ 5x temuan'
 UNION ALL SELECT 'early',5,0,2,'< 3x temuan'
 UNION ALL SELECT 'early',4,3,3,'3x temuan'
 UNION ALL SELECT 'early',3,4,4,'4x temuan'
 UNION ALL SELECT 'early',2,5,5,'5x temuan'
 UNION ALL SELECT 'early',1,6,NULL,'≥ 6x temuan'
 UNION ALL SELECT 'middle',5,0,1,'< 2x temuan'
 UNION ALL SELECT 'middle',4,2,2,'2x temuan'
 UNION ALL SELECT 'middle',3,3,3,'3x temuan'
 UNION ALL SELECT 'middle',2,4,4,'4x temuan'
 UNION ALL SELECT 'middle',1,5,NULL,'≥ 5x temuan'
 UNION ALL SELECT 'late',5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late',4,1,1,'1x temuan'
 UNION ALL SELECT 'late',3,2,2,'2x temuan'
 UNION ALL SELECT 'late',2,3,3,'3x temuan'
 UNION ALL SELECT 'late',1,4,NULL,'≥ 4x temuan'
 UNION ALL SELECT 'late_social',5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late_social',4,1,1,'1x temuan'
 UNION ALL SELECT 'late_social',3,2,2,'2x temuan'
 UNION ALL SELECT 'late_social',2,3,4,'3x - 4x temuan'
 UNION ALL SELECT 'late_social',1,5,NULL,'≥ 5x temuan'
) s ON s.kind=CASE
 WHEN r.month_from=5 AND c.code IN ('cs_kerjasama_1','cs_kerjasama_2','cs_belajar_1','cs_belajar_2') THEN 'late_social'
 WHEN r.month_from=5 THEN 'late'
 WHEN c.code IN ('cs_layanan_1','cs_layanan_2','cs_etika_1') THEN 'standard'
 WHEN r.month_from=1 THEN 'early'
 ELSE 'middle' END
WHERE r.code IN ('cs_obat_1_2','cs_obat_3_4','cs_obat_5_6')
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);

COMMIT;
