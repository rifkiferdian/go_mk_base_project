START TRANSACTION;
INSERT INTO trainee_assessment_aspects (code, name, sort_order) VALUES
('layanan','LAYANAN',1), ('kemampuan','KEMAMPUAN',2), ('etika','ETIKA',3),
('kerjasama','KERJASAMA',4), ('keinginan_belajar','KEINGINAN BELAJAR',5)
ON DUPLICATE KEY UPDATE code=VALUES(code);

INSERT INTO trainee_assessment_criteria (aspect_id, code, description, weight, sort_order)
SELECT a.id, rubric.code, rubric.description, 1, rubric.sort_order
FROM trainee_assessment_aspects a JOIN (
 SELECT 'layanan' aspect_code, 'layanan_1' code, 'Tidak menjaga kebersihan dan kerapian area kerja serta sarana prasarana kerja di bagiannya' description, 1 sort_order
 UNION ALL SELECT 'layanan','layanan_2','Tidak sigap dalam melakukan tanggung jawab pekerjaan dan atau tidak ramah dalam melayani konsumen',2
 UNION ALL SELECT 'kemampuan','kemampuan_1','Tidak melakukan penataan barang gudang berdasarkan prinsip FEFO',1
 UNION ALL SELECT 'kemampuan','kemampuan_2','Tidak melakukan pengecekan kondisi fisik barang datang (sampling)',2
 UNION ALL SELECT 'kemampuan','kemampuan_3','Kesalahan dalam melaksanakan alur mekanisme kerja',3
 UNION ALL SELECT 'etika','etika_1','Sikap & Cara Berbicara tidak sesuai budaya dan peraturan perusahaan',1
 UNION ALL SELECT 'kerjasama','kerjasama_1','Jarang berbagi informasi atau membantu rekan kerja saat dibutuhkan',1
 UNION ALL SELECT 'kerjasama','kerjasama_2','Koordinasi dengan rekan kerja dan bagian lain sering kurang jelas sehingga pekerjaan terhambat',2
 UNION ALL SELECT 'keinginan_belajar','keinginan_belajar_1','Sering salah memahami arahan kerja dan perlu diingatkan atau dijelaskan kembali',1
 UNION ALL SELECT 'keinginan_belajar','keinginan_belajar_2','Kurang berinisiatif untuk belajar hal baru atau meningkatkan kemampuan kerja',2
) rubric ON rubric.aspect_code=a.code
ON DUPLICATE KEY UPDATE code=VALUES(code);

-- Layanan dan etika: 0 / 1-2 / 3 / 4 / >=5 temuan.
-- Subaspek lainnya: 0-2 / 3 / 4 / 5 / >=6 temuan.
INSERT INTO trainee_assessment_score_rules (criterion_id, score, finding_min, finding_max, description)
SELECT c.id, s.score,
 CASE WHEN c.code IN ('layanan_1','layanan_2','etika_1') THEN s.standard_min ELSE s.other_min END,
 CASE WHEN c.code IN ('layanan_1','layanan_2','etika_1') THEN s.standard_max ELSE s.other_max END,
 CASE WHEN c.code IN ('layanan_1','layanan_2','etika_1') THEN s.standard_label ELSE s.other_label END
FROM trainee_assessment_criteria c CROSS JOIN (
 SELECT 5 score, 0 standard_min, 0 standard_max, 'Tidak ada temuan' standard_label, 0 other_min, 2 other_max, '< 3x temuan' other_label
 UNION ALL SELECT 4,1,2,'1x - 2x temuan',3,3,'3x temuan'
 UNION ALL SELECT 3,3,3,'3x temuan',4,4,'4x temuan'
 UNION ALL SELECT 2,4,4,'4x temuan',5,5,'5x temuan'
 UNION ALL SELECT 1,5,NULL,'>= 5x temuan',6,NULL,'>= 6x temuan'
) s
WHERE c.code IN ('layanan_1','layanan_2','kemampuan_1','kemampuan_2','kemampuan_3','etika_1','kerjasama_1','kerjasama_2','keinginan_belajar_1','keinginan_belajar_2')
ON DUPLICATE KEY UPDATE score=VALUES(score);
COMMIT;
