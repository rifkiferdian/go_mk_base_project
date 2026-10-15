-- A criterion belongs to a department and is shared by every training month.
-- Only its score rules vary by rubric/month period.
ALTER TABLE trainee_assessment_criteria
    ADD COLUMN department_id BIGINT UNSIGNED NULL AFTER rubric_id,
    ADD CONSTRAINT fk_criterion_department FOREIGN KEY (department_id) REFERENCES departments(id);

-- The pre-period master is the canonical Gudang criterion set. Criteria made by
-- migration 000006 are mapped back to these canonical rows by code.
UPDATE trainee_assessment_criteria c
JOIN departments d ON d.code='gudang'
SET c.department_id=d.id
WHERE c.rubric_id IS NULL;
UPDATE trainee_assessment_criteria c
JOIN trainee_assessment_rubrics r ON r.id=c.rubric_id
SET c.department_id=r.department_id;

-- Preserve any details already created with a duplicated period criterion.
UPDATE trainee_assessment_details ad
JOIN trainee_assessment_criteria duplicate ON duplicate.id=ad.criterion_id AND duplicate.rubric_id IS NOT NULL
JOIN trainee_assessment_criteria canonical ON canonical.rubric_id IS NULL
    AND canonical.department_id=duplicate.department_id AND canonical.code=duplicate.code
SET ad.criterion_id=canonical.id;

RENAME TABLE trainee_assessment_score_rules TO trainee_assessment_score_rules_period_source;
CREATE TABLE trainee_assessment_score_rules (
    rubric_id BIGINT UNSIGNED NOT NULL,
    criterion_id BIGINT UNSIGNED NOT NULL,
    score TINYINT UNSIGNED NOT NULL,
    finding_min SMALLINT UNSIGNED NOT NULL,
    finding_max SMALLINT UNSIGNED NULL,
    description VARCHAR(150) NOT NULL,
    PRIMARY KEY (rubric_id, criterion_id, score),
    CONSTRAINT trainee_rules_score_check CHECK (score BETWEEN 1 AND 5),
    CONSTRAINT trainee_rules_range_check CHECK (finding_max IS NULL OR finding_max >= finding_min),
    FOREIGN KEY (rubric_id) REFERENCES trainee_assessment_rubrics(id),
    FOREIGN KEY (criterion_id) REFERENCES trainee_assessment_criteria(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO trainee_assessment_score_rules(rubric_id,criterion_id,score,finding_min,finding_max,description)
SELECT duplicate.rubric_id,canonical.id,s.score,s.finding_min,s.finding_max,s.description
FROM trainee_assessment_score_rules_period_source s
JOIN trainee_assessment_criteria duplicate ON duplicate.id=s.criterion_id AND duplicate.rubric_id IS NOT NULL
JOIN trainee_assessment_criteria canonical ON canonical.rubric_id IS NULL
    AND canonical.department_id=duplicate.department_id AND canonical.code=duplicate.code;

DROP TABLE trainee_assessment_score_rules_period_source;
DELETE FROM trainee_assessment_criteria WHERE rubric_id IS NOT NULL;
ALTER TABLE trainee_assessment_criteria
    DROP FOREIGN KEY fk_criteria_rubric,
    DROP INDEX criterion_rubric_code,
    DROP COLUMN rubric_id,
    MODIFY department_id BIGINT UNSIGNED NOT NULL,
    ADD UNIQUE KEY criterion_department_code (department_id,code),
    ADD KEY criterion_department_sort (department_id,aspect_id,sort_order);

-- Use one stable wording for all Gudang periods.
UPDATE trainee_assessment_criteria c
JOIN departments d ON d.id=c.department_id AND d.code='gudang'
SET c.description=CASE c.code
 WHEN 'layanan_1' THEN 'Tidak menjaga kebersihan dan kerapian area kerja serta sarana prasarana kerja dibagiannya'
 WHEN 'layanan_2' THEN 'Tidak sigap dalam melakukan tanggung jawab pekerjaan dan atau tidak ramah dalam melayani konsumen'
 WHEN 'kemampuan_1' THEN 'Tidak melakukan penataan barang gudang berdasarkan prinsip FEFO'
 WHEN 'kemampuan_2' THEN 'Tidak melakukan pengecekan kondisi fisik barang datang (sampling)'
 WHEN 'kemampuan_3' THEN 'Kesalahan dalam melaksanakan alur mekanisme kerja'
 WHEN 'etika_1' THEN 'Sikap & Cara Berbicara tidak sesuai budaya dan peraturan perusahaan'
 WHEN 'kerjasama_1' THEN 'Jarang berbagi informasi atau membantu rekan kerja saat dibutuhkan'
 WHEN 'kerjasama_2' THEN 'Koordinasi dengan rekan kerja dan bagian lain sering kurang jelas sehingga pekerjaan terhambat'
 WHEN 'keinginan_belajar_1' THEN 'Sering salah memahami arahan kerja dan perlu diingatkan atau dijelaskan kembali'
 WHEN 'keinginan_belajar_2' THEN 'Kurang berinisiatif untuk belajar hal baru atau meningkatkan kemampuan kerja'
 ELSE c.description END;
