### ACCESS CONTROL FLOW

The basic idea to create, manage and grant permissions for members defined in access-control module is:

- The `access-control` module will output (by using terraform `output`) member by access levels, with data like this:
```
    members_by_access_level = {
      data = {
        low = [
          "bao.nguyen@manabie.com",
        ],
        moderate = [
          "nhan.huynh@manabie.com",
          "phi.pham@manabie.com",
        ],
        high = [
          "daniel.sont@manabie.com",
        ],
        super = [
          "tuananh.pham@manabie.com",
        ],
      }
    }
```
this data means in `data` function:
  - `nhan.huynh@manabie.com` has `low` access level.
  - `nguyen.nguyen@manabie.com` has `moderate` access level.
  - `tuananh.pham@manabie.com` has `super` access level.
  - ... and so on.

- Having a separate module named `postgresql-roles` to:
  - Create `read_only_role` that can only read data.
  - Create `read_write_role` that can read & write data.
these roles will be used in the module below.

- Having a separate module named `postgresql-grant` to:
  - Grant `SELECT` permission for `read_only_role`.
  - Grant `SELECT`, `INSERT`, `UPDATE` to `read_write_role`.

- Having a separate module named `project-roles` to:
  - Create all custom GCP roles with predefined permissions for all functions combined with all access levels.
    Suppose we have 4 functions: `data`, `backend`, `frontend` and `platform`, and 4 access levels: `low`, `moderate`, `high`, `super`,
    then this module will create these custom roles with predefined permissions:
    - `customroles.data.low`
    - `customroles.data.moderate`
    - `customroles.data.high`
    - `customroles.data.super`
    - ... and so on.

  - Define custom roles by function and access level:
  ```
    roles = [
    // DATA roles
    {
      id          = "customroles.data.low"
      title       = "Data Role for Low Access Level"
      description = ""
      base_roles  = [
         "roles/cloudsql.viewer",     
      ]
      permissions = []
    },
    {
      id          = "customroles.data.moderate"
      title       = "Data Role for Moderate Access Level"
      description = ""
      base_roles = [
        "roles/cloudsql.viewer", 
        "roles/cloudsql.client",
        "roles/cloudsql.instanceUser",
      ]
      permissions = []
    },
    ...
  ```

(The `customroles.data.low` above extends the managed `roles/cloudsql.viewer` role from GCP, while `customroles.data.moderate` extends 
`roles/cloudsql.client` and `roles/cloudsql.instanceUser` roles. We can also define individual permissions for the custom roles, using `permissions` field.)

- Having another module named `project-grant` to:
  -  Take the `members_by_access_level` output from `access-control` module above.
  - Define a role mapping by function and level:
  ```
    role_by_access_level = {
    data = {
      low = {
        stag = {
          can_read_databases = true
          custom_roles = [
            "customroles.data.moderate",
          ]
        }
        uat = {
          custom_roles = [
            "customroles.data.low",
          ]
        }
      }
      moderate = {
        stag = {
          can_read_databases  = true
          can_write_databases = true
          custom_roles = [
            "customroles.data.high",
          ]
        }
        uat = {
          can_read_databases = true
          custom_roles = [
            "customroles.data.moderate",
          ]
        }
        prod = {
          custom_roles = [
            "customroles.data.low",
          ]
        }
      }
      high = {
        stag = {
          can_read_databases  = true
          can_write_databases = true
          custom_roles = [
            "customroles.data.high",
          ]
        }
        uat = {
          can_read_databases  = true
          can_write_databases = true
          custom_roles = [
            "customroles.data.high",
          ]
        }
        prod = {
          can_read_databases = true
          custom_roles = [
            "customroles.data.moderate",
          ]
        }
      }
      super = {
        stag = {
          can_read_databases  = true
          can_write_databases = true
          custom_roles = [
            "customroles.data.super",
          ]
        }
        uat = {
          can_read_databases  = true
          can_write_databases = true
          custom_roles = [
            "customroles.data.super",
          ]
        }
        prod = {
          can_read_databases  = true
          can_write_databases = true
          custom_roles = [
            "customroles.data.super",
          ]
        }
      }
    }
  }
(The mapping above means `data` function with `low` access level can read from databases on `staging`, and also has the `customroles.data.moderate` role on `staging` also. But it doesn't have permission to read from databases on `uat`, and only has the `customroles.data.low` role on `uat` also. Same for other access level.)


  - Build a terraform list like this (for `staging` env):
  ```
   member_function_level_role = [
  {
    "custom_role_id" = "customroles.data.high"
    "function" = "data"
    "level" = "high"
    "member" = "daniel.sont@manabie.com"
  },
  {
    "custom_role_id" = "customroles.data.moderate"
    "function" = "data"
    "level" = "low"
    "member" = "bao.nguyen@manabie.com"
  },
  {
    "custom_role_id" = "customroles.data.high"
    "function" = "data"
    "level" = "moderate"
    "member" = "nhan.huynh@manabie.com"
  },
  {
    "custom_role_id" = "customroles.data.high"
    "function" = "data"
    "level" = "moderate"
    "member" = "phi.pham@manabie.com"
  },
  {
    "custom_role_id" = "customroles.data.super"
    "function" = "data"
    "level" = "super"
    "member" = "tuananh.pham@manabie.com"
  },
]
  ```
(As you can see, member `bao.nguyen@manabie.com` has access level `low` in `data` function; read from the `role_by_access_level` above,
we can see that he has `customroles.data.moderate` role on `staging`. Same for other members.)

  - Build 2 Postgresql role lists:
  ```
  postgresql_read_only_users = toset([
  "bao.nguyen@manabie.com",
])
postgresql_read_write_users = toset([
  "daniel.sont@manabie.com",
  "nhan.huynh@manabie.com",
  "phi.pham@manabie.com",
  "tuananh.pham@manabie.com",
])
  ```
(Again, member ` bao.nguyen@manabie.com` can only read from databases on `staging`, while `nhan.huynh@manabie.com` has access
level `moderate` in `data` function; read from the `role_by_access_level` above, he can write to databases on `staging` also. Same for other members.)

---

The above approach has a limitation, is that by using `dependency` between module, the dependant module can't run the plan until
its parent module applies first, because terraform `output` requires terraform apply so it can modify the state. See [this](https://github.com/gruntwork-io/terragrunt/issues/720#issuecomment-497888756)
and [this](https://terragrunt.gruntwork.io/docs/features/execute-terraform-commands-on-multiple-modules-at-once/#unapplied-dependency-and-mock-outputs)
for more details.

So in order to make this work, whenever the `access-control` project changes, we need to:
  1. Run plan for the `access-control` project.
  2. Run apply for the `access-control` project.
  3. Run plan for all of the dependant projects.
  4. Run apply for all of the dependant projects.

The `access-control-workflow.sh` script does the 3) and 4) step above.

We also created a custom access-control Atlantis workflow. That workflow works like all other workflows, except it 
will run that `access-control-workflow.sh` script after the `access-control` module is done applying its changes.
See [this](https://github.com/manabie-com/backend/blob/develop/deployments/helm/platforms/atlantis/production-values.yaml#L101) for more details.

Since there's quite a lot of dependant projects, so it also speed up the process
by running all the commands in background.

---

### Summary

The access control have several modules:

- `access-control` module: define members along with their access level.
- `project-roles` module: create GCP custom roles for backend/data/mobile/platform/web functions.
- `project-grant` module: grant GCP custom roles (created by `project-roles` module) to each member, depend on the access level.
- `postgresql-roles` module: create Postgres custom roles (in all Postgres instances):
    - read_only_role
    - read_write_role
    - bypass_rls_role
    - replication_role
    - Add dev emails to Postgres instance. Depend on access level, only devs with high access level will be added to Production databases.
- `postgresql-grant` module: grant Postgres custom roles (created by `postgresql-roles` module) privileges (`SELECT`, `INSERT`, `UPDATE`...etc) to all existing databases.
- `access-control-postgresql` module: grant members (created by `access-control` module) to `read_only_role` or `read_write_role`, depends on theirs access level.
