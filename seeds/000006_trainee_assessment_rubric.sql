START TRANSACTION;
INSERT INTO departments(code,name) VALUES ('gudang','Gudang')
ON DUPLICATE KEY UPDATE code=VALUES(code);
INSERT INTO trainee_assessment_aspects(code,name,sort_order) VALUES
('layanan','LAYANAN',1),('kemampuan','KEMAMPUAN',2),('etika','ETIKA',3),
('kerjasama','KERJASAMA',4),('keinginan_belajar','KEINGINAN BELAJAR',5)
ON DUPLICATE KEY UPDATE code=VALUES(code);

INSERT INTO trainee_assessment_criteria(department_id,aspect_id,code,description,weight,sort_order)
SELECT d.id,a.id,c.code,c.description,1,c.sort_order
FROM departments d
CROSS JOIN (
 SELECT 'layanan' aspect_code,'layanan_1' code,'Tidak menjaga kebersihan dan kerapian area kerja serta sarana prasarana kerja dibagiannya' description,1 sort_order
 UNION ALL SELECT 'layanan','layanan_2','Tidak sigap dalam melakukan tanggung jawab pekerjaan dan atau tidak ramah dalam melayani konsumen',2
 UNION ALL SELECT 'kemampuan','kemampuan_1','Tidak melakukan penataan barang gudang berdasarkan prinsip FEFO',1
 UNION ALL SELECT 'kemampuan','kemampuan_2','Tidak melakukan pengecekan kondisi fisik barang datang (sampling)',2
 UNION ALL SELECT 'kemampuan','kemampuan_3','Kesalahan dalam melaksanakan alur mekanisme kerja',3
 UNION ALL SELECT 'etika','etika_1','Sikap & Cara Berbicara tidak sesuai budaya dan peraturan perusahaan',1
 UNION ALL SELECT 'kerjasama','kerjasama_1','Jarang berbagi informasi atau membantu rekan kerja saat dibutuhkan',1
 UNION ALL SELECT 'kerjasama','kerjasama_2','Koordinasi dengan rekan kerja dan bagian lain sering kurang jelas sehingga pekerjaan terhambat',2
 UNION ALL SELECT 'keinginan_belajar','keinginan_belajar_1','Sering salah memahami arahan kerja dan perlu diingatkan atau dijelaskan kembali',1
 UNION ALL SELECT 'keinginan_belajar','keinginan_belajar_2','Kurang berinisiatif untuk belajar hal baru atau meningkatkan kemampuan kerja',2
) c JOIN trainee_assessment_aspects a ON a.code=c.aspect_code
WHERE d.code='gudang'
ON DUPLICATE KEY UPDATE code=VALUES(code);
COMMIT;
