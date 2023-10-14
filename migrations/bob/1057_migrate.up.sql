CREATE TABLE IF NOT EXISTS public.academic_years (
    academic_year_id TEXT NOT NULL,
    school_id INTEGER NOT NULL,
    name text NOT NULL,
    start_year_date timestamp with time zone NOT NULL,
    end_year_date timestamp with time zone NOT NULL,
    status TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT pk__academic_years PRIMARY KEY (academic_year_id)
);

CREATE TABLE IF NOT EXISTS public.courses_academic_years (
    course_id TEXT NOT NULL,
    academic_year_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT pk__courses_academic_years PRIMARY KEY (course_id,academic_year_id),
    CONSTRAINT fk__courses_academic_years__course_id FOREIGN KEY (course_id) REFERENCES public.courses (course_id),
    CONSTRAINT fk__courses_academic_years__academic_year_id FOREIGN KEY (academic_year_id) REFERENCES public.academic_years (academic_year_id)
);

ALTER TABLE public.courses
  ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'COURSE_STATUS_NONE';
