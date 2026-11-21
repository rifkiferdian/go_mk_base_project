-- Panduan penilaian trainee untuk jabatan LABELING GUDANG.
START TRANSACTION;

INSERT INTO trainee_assessment_aspects(code,name,sort_order) VALUES
('layanan','LAYANAN',1),('kemampuan','KEMAMPUAN',2),('etika','ETIKA',3),
('kerjasama','KERJASAMA',4),('keinginan_belajar','KEINGINAN BELAJAR',5)
ON DUPLICATE KEY UPDATE name=VALUES(name),sort_order=VALUES(sort_order);

INSERT INTO trainee_assessment_rubrics(department_id,position_id,code,name,month_from,month_to,is_active)
SELECT p.department_id,p.id,x.code,CONCAT(p.name,' - ',x.label),x.month_from,x.month_to,1
FROM positions p CROSS JOIN (
 SELECT 'labeling_gudang_1_2' code,'Bulan 1-2' label,1 month_from,2 month_to
 UNION ALL SELECT 'labeling_gudang_3_4','Bulan 3-4',3,4
 UNION ALL SELECT 'labeling_gudang_5_6','Bulan 5-6',5,6
) x WHERE p.code='labeling_gudang'
ON DUPLICATE KEY UPDATE name=VALUES(name),month_from=VALUES(month_from),month_to=VALUES(month_to),is_active=1;

INSERT INTO trainee_assessment_criteria(department_id,position_id,aspect_id,code,description,weight,sort_order)
SELECT p.department_id,p.id,a.id,c.code,c.description,1,c.sort_order
FROM positions p CROSS JOIN (
 SELECT 'layanan' aspect_code,'labeling_layanan_1' code,'Tidak menjaga kebersihan dan kerapian area kerja serta sarana prasarana kerja dibagiannya' description,1 sort_order
 UNION ALL SELECT 'layanan','labeling_layanan_2','Tidak sigap dalam melakukan tanggung jawab pekerjaan dan atau tidak ramah dalam melayani konsumen',2
 UNION ALL SELECT 'kemampuan','labeling_kemampuan_1','Penempelan label menutupi tulisan atau informasi tentang produk, dengan posisi tulisan tidak mudah di baca oleh pelanggan',1
 UNION ALL SELECT 'kemampuan','labeling_kemampuan_2','Memastikan semua produk sudah terlabeli',2
 UNION ALL SELECT 'kemampuan','labeling_kemampuan_3','Kesalahan dalam melaksanakan alur mekanisme kerja',3
 UNION ALL SELECT 'etika','labeling_etika_1','Sikap & Cara Berbicara tidak sesuai budaya dan peraturan perusahaan',1
 UNION ALL SELECT 'kerjasama','labeling_kerjasama_1','Jarang berbagi informasi atau membantu rekan kerja saat dibutuhkan',1
 UNION ALL SELECT 'kerjasama','labeling_kerjasama_2','Koordinasi dengan rekan kerja dan bagian lain sering kurang jelas sehingga pekerjaan terhambat',2
 UNION ALL SELECT 'keinginan_belajar','labeling_belajar_1','Sering salah memahami arahan kerja dan perlu diingatkan atau dijelaskan kembali',1
 UNION ALL SELECT 'keinginan_belajar','labeling_belajar_2','Kurang berinisiatif untuk belajar hal baru atau meningkatkan kemampuan kerja',2
) c JOIN trainee_assessment_aspects a ON a.code=c.aspect_code
WHERE p.code='labeling_gudang'
ON DUPLICATE KEY UPDATE aspect_id=VALUES(aspect_id),description=VALUES(description),weight=VALUES(weight),sort_order=VALUES(sort_order);

INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT c.id,p.id FROM trainee_assessment_criteria c JOIN positions p ON p.id=c.position_id
WHERE p.code='labeling_gudang' AND c.code LIKE 'labeling_%';

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT r.id,c.id,s.score,s.finding_min,s.finding_max,s.description
FROM trainee_assessment_rubrics r
JOIN positions p ON p.id=r.position_id
JOIN trainee_assessment_criterion_positions cp ON cp.position_id=p.id
JOIN trainee_assessment_criteria c ON c.id=cp.criterion_id AND c.code LIKE 'labeling_%'
JOIN (
 SELECT 'standard' kind,5 score,0 finding_min,0 finding_max,'tidak ada temuan' description
 UNION ALL SELECT 'standard',4,1,2,'1x - 2x temuan'
 UNION ALL SELECT 'standard',3,3,3,'3x temuan'
 UNION ALL SELECT 'standard',2,4,4,'4x temuan'
 UNION ALL SELECT 'standard',1,5,NULL,'>= 5x temuan'
 UNION ALL SELECT 'early',5,0,2,'< 3x temuan'
 UNION ALL SELECT 'early',4,3,3,'3x temuan'
 UNION ALL SELECT 'early',3,4,4,'4x temuan'
 UNION ALL SELECT 'early',2,5,5,'5x temuan'
 UNION ALL SELECT 'early',1,6,NULL,'>= 6x temuan'
 UNION ALL SELECT 'middle',5,0,1,'< 2x temuan'
 UNION ALL SELECT 'middle',4,2,2,'2x temuan'
 UNION ALL SELECT 'middle',3,3,3,'3x temuan'
 UNION ALL SELECT 'middle',2,4,4,'4x temuan'
 UNION ALL SELECT 'middle',1,5,NULL,'>= 5x temuan'
 UNION ALL SELECT 'late',5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late',4,1,1,'1x temuan'
 UNION ALL SELECT 'late',3,2,2,'2x temuan'
 UNION ALL SELECT 'late',2,3,3,'3x temuan'
 UNION ALL SELECT 'late',1,4,NULL,'>= 4x temuan'
 UNION ALL SELECT 'late_social',5,0,0,'tidak ada temuan'
 UNION ALL SELECT 'late_social',4,1,1,'1x temuan'
 UNION ALL SELECT 'late_social',3,2,2,'2x temuan'
 UNION ALL SELECT 'late_social',2,3,4,'3x - 4x temuan'
 UNION ALL SELECT 'late_social',1,5,NULL,'>= 5x temuan'
) s ON s.kind=CASE
 WHEN c.code IN ('labeling_layanan_1','labeling_layanan_2','labeling_etika_1') AND r.month_from<5 THEN 'standard'
 WHEN c.code IN ('labeling_layanan_1','labeling_layanan_2','labeling_etika_1') THEN 'late'
 WHEN c.code IN ('labeling_kemampuan_1','labeling_kemampuan_2','labeling_kemampuan_3') AND r.month_from=1 THEN 'early'
 WHEN c.code IN ('labeling_kemampuan_1','labeling_kemampuan_2','labeling_kemampuan_3') AND r.month_from=3 THEN 'middle'
 WHEN c.code IN ('labeling_kemampuan_1','labeling_kemampuan_2','labeling_kemampuan_3') THEN 'late'
 WHEN r.month_from=1 THEN 'early'
 WHEN r.month_from=3 THEN 'middle'
 ELSE 'late_social' END
WHERE p.code='labeling_gudang'
  AND r.code IN ('labeling_gudang_1_2','labeling_gudang_3_4','labeling_gudang_5_6')
ON DUPLICATE KEY UPDATE finding_min=VALUES(finding_min),finding_max=VALUES(finding_max),description=VALUES(description);

COMMIT;
