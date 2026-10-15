-- Data-preserving reverse migration is ambiguous after criteria/rules are edited.
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Rollback refused: restore the pre-000007 backup if required';
