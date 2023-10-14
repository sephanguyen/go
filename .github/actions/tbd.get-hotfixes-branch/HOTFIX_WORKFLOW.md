#### Hotfix workflow
- [Document](https://manabie.atlassian.net/wiki/spaces/TECH/pages/454885865/Trunk-Based+Development#Can-reproduce-on-trunk%5BhardBreak%5D)
- Suggestion steps:
  ```
  # Choose hotfixes branch from above branches & checkout
  git checkout hotfixes_branch
  git checkout -b hotfix-LT-XXX
  git cherry-pick commit-a
  git cherry-pick commit-b
  git push origin
  # Then create PR to hotfixes branch
  ```
  - Notify your team to review and get help merge from your squad lead.
  - Find release tag.
      - Go to release tab, ex: [BE release tab](https://github.com/manabie-com/backend/releases).
      - <img width="500" alt="find-release-tag" src="https://user-images.githubusercontent.com/34020090/199398782-4054b603-b2a0-48a1-8bc1-a2f9a49d8077.png">
  - [Build](https://github.com/manabie-com/eibanam/discussions/2448) & [deploy](https://github.com/manabie-com/eibanam/discussions/2449) docs: You can build UAT/PROD manually.
  - Notify release team in [#dev-release](https://manabie.slack.com/archives/CR2AR72SZ) channel.
  - Then trigger deploy UAT/PROD.
