ALTER TABLE ONLY public.classes DROP CONSTRAINT classes_subjects_check;

ALTER TABLE public.classes ADD CHECK (subjects <@ ARRAY[
                'SUBJECT_MATHS',
                'SUBJECT_BIOLOGY',
                'SUBJECT_PHYSICS',
                'SUBJECT_CHEMISTRY',
                'SUBJECT_GEOGRAPHY',
                'SUBJECT_ENGLISH',
                'SUBJECT_ENGLISH_2',
                'SUBJECT_JAPANESE',
                'SUBJECT_SCIENCE',
                'SUBJECT_SOCIAL_STUDIES',
                'SUBJECT_LITERATURE'
            ]);
