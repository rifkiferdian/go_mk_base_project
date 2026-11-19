CREATE TABLE trainee_assessment_criterion_positions (
    criterion_id BIGINT UNSIGNED NOT NULL,
    position_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY(criterion_id,position_id),
    KEY criterion_positions_position(position_id,criterion_id),
    CONSTRAINT criterion_positions_criterion_fk FOREIGN KEY(criterion_id) REFERENCES trainee_assessment_criteria(id) ON DELETE CASCADE,
    CONSTRAINT criterion_positions_position_fk FOREIGN KEY(position_id) REFERENCES positions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT id,position_id FROM trainee_assessment_criteria;

-- Gabungkan empat set subaspek Checker menjadi set milik CHECKER ATK & CRAFT.
INSERT IGNORE INTO trainee_assessment_criterion_positions(criterion_id,position_id)
SELECT canonical.id,target.id
FROM trainee_assessment_criteria canonical
JOIN positions source ON source.id=canonical.position_id AND source.code='checker_atk_craft'
JOIN positions target ON target.code IN ('checker_atk_craft','checker_busana','checker_gudang','checker_supermarket')
WHERE canonical.code LIKE 'checker_%';

UPDATE trainee_assessment_score_rules sr
JOIN trainee_assessment_criteria duplicate ON duplicate.id=sr.criterion_id
JOIN positions dp ON dp.id=duplicate.position_id AND dp.code IN ('checker_busana','checker_gudang','checker_supermarket')
JOIN trainee_assessment_criteria canonical ON canonical.code=duplicate.code
JOIN positions cp ON cp.id=canonical.position_id AND cp.code='checker_atk_craft'
SET sr.criterion_id=canonical.id;

DELETE c FROM trainee_assessment_criteria c
JOIN positions p ON p.id=c.position_id
WHERE p.code IN ('checker_busana','checker_gudang','checker_supermarket') AND c.code LIKE 'checker_%';
