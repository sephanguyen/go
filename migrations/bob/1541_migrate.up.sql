ALTER TABLE public.user_tag DROP CONSTRAINT IF EXISTS user_tag__user_tag_type__check;

ALTER TABLE public.user_tag ADD CONSTRAINT user_tag__user_tag_type__check CHECK (
  user_tag_type = ANY (ARRAY[
		'USER_TAG_TYPE_STUDENT',
		'USER_TAG_TYPE_STUDENT_DISCOUNT',
		'USER_TAG_TYPE_PARENT',
		'USER_TAG_TYPE_PARENT_DISCOUNT',
        'USER_TAG_TYPE_STAFF'
  ]::text[])
);
