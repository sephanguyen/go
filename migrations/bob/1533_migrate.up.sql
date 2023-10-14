UPDATE users SET first_name_phonetic = NULL where first_name_phonetic = '' AND user_group = 'USER_GROUP_STUDENT';

UPDATE users SET last_name_phonetic = NULL where last_name_phonetic = '' AND user_group = 'USER_GROUP_STUDENT';

UPDATE users SET full_name_phonetic = NULL where full_name_phonetic = '' AND user_group = 'USER_GROUP_STUDENT';
