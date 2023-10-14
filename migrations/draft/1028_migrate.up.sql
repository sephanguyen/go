CREATE TABLE IF NOT EXISTS public.github_repo_state (
	org text NOT NULL,
    repo text NOT NULL,
    is_blocked bool NOT NULL,
	CONSTRAINT github_repo_pkey PRIMARY KEY (org,repo)
);

INSERT INTO public.github_repo_state (org,repo,is_blocked) VALUES
	('manabie-com','backend',false);

ALTER TABLE public.github_pr_statistic
    ADD COLUMN IF NOT EXISTS repo TEXT NULL,
    ADD COLUMN IF NOT EXISTS base_branch TEXT NULL,
    ADD COLUMN IF NOT EXISTS merge_status TEXT NULL,
    ADD CONSTRAINT repo_pr_number UNIQUE (repo, number);