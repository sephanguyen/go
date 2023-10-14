## Access Control

In this folder you can find an accurate list of all members that allowed to access our data.

* People that sit and/or speak together should stay in the same file.
* Anyone may belong to 1 or many squads/functions.

## Adding a new member

Requirements:
- Member's Github account
- Member's designated squad(s) and function(s)

Reference PR: [#5863](https://github.com/manabie-com/backend/pull/5863/files)

Steps:

1. Create a PR to `develop` branch to add the info block of the new member. For example:

```hcl
    {
      name  = "anhcnt197vn"                 // Member's name: use real name or Github username
      email = "tuananh.cao@manabie.com"     // Member's company email address

      github = {
        account = "anhcnt197vn"             // Member's Github account. Must be an existing and valid Github account.
        role    = "member"                  // Member's role. Should usually be "member".
      }

      squads = [
        {
          name = "lesson"                   // Member's squad ("lesson" squad in this case).
          role = "member"                   // Member's role in the squad. Should usually be "member".
        },
        {
          name = "syllabus"                 // Possible to assign member to multiple squads.
          role = "member"
        },
      ]

      functions = [
        {
          name = "web"                      // Member's function ("web" function in this case).
          role = "member"                   // Member's role in the function. Should usually be "member".
        },
      ]
    },
```

Notes:
- A member (idenitified by `name` and `email`) can exist in only one file.
But they still can be assigned to multiple squads/functions.
- When creating new squads/functions, remember to update the condition in [`deployments/terraform/live/workspace/access-control/README.md`](../../../modules/access-control/variables.tf#L38)

2. Tag platform squad (secondary on-call) to review the PR. The PR will then be
handled by platform squad. In the meanwhile, do not close or update the PR without
consulting platform members first.

## Removing a member

*Note: When removing a member when they leaves the company, please create a PR to remove
them by their last day of work. Otherwise, when HR admin removes their email account,
this access control system will encounter conflicts.*

Reference PR: [#5800](https://github.com/manabie-com/backend/pull/5800/files)

Steps:

1. Create a PR to `develop` branch to remove the info block of the member.
2. Tag platform squad (secondary on-call) to review the PR (similarly when adding a new member).

## Squads, functions, and roles

* Squads are task forces
    * Each squad is in charge of one or multiple bounded contexts
    * For example: `lesson`, `usermgmt`, `adobo`, etc...
    * List of squads can be found [here](../../../modules/access-control/variables.tf#L42)
* Functions are defined by skill sets and later will define the resources you can access
    * For example: `backend`, `frontend` (`web`), `mobile`, etc...
    * Function decides the resources a member can access (e.g. backend member can access to k8s infrastructure, postgresql)
    * List of functions can be found [here](../../../modules/access-control/variables.tf#L75)
* Roles:
    * `techlead`: Leader of a squad who can:
        * Decrypt/Encrypt secrets of services belong to that squad
        * Approve the adhoc workflows triggered by squad members
    * Other roles `manager`, `maintainer`, `member` do nothing much for now

## References

- [People Management for Squads (Product, Tech, Design)](https://docs.google.com/spreadsheets/d/1UE5Anm-hA3U3h0TC0qMvuvrwhtnopvBCkKvxJdrn9vA/edit#gid=150665427)