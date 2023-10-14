UPDATE books
SET book_type = 'BOOK_TYPE_GENERAL'::TEXT
WHERE book_type IS NULL;