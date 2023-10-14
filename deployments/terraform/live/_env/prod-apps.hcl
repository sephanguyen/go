locals {
  config   = read_terragrunt_config("config.hcl").locals
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))

  postgresql_port_config = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/postgresql-port.hcl")

  project_id             = local.env_vars.locals.project_id
  runs_on_project_id     = local.config.runs_on_project_id
  service_account_prefix = local.config.service_account_prefix

  org = local.config.org
  env = local.env_vars.locals.env

  db_prefix      = try(local.config.db_prefix, "")
  db_user_prefix = try(local.config.db_user_prefix, "")

  # Import the global service defintions
  service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/prod-defs.yaml"))

  # `databases` block defines the logical databases to be created
  # in every database instance.
  databases = concat(
    # Create a database for each service
    [
      for service in local.service_definitions : "${local.db_prefix}${service.name}"
      if try(service.postgresql.createdb, false)
    ],

    # Create a metadata database for hasura-v2
    [
      for service in local.service_definitions : "${local.db_prefix}${service.name}_hasura_metadata"
      if try(service.postgresql.createdb, false) && try(service.hasura.v2_enabled, false)
    ],

    # Other databases
    ["${local.db_prefix}unleashv2"],
  )

  # `users` block defines the database users to be created.
  users = flatten(concat(
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
      if !try(service.disable_iam, false)
    ],

    # The following block contains other users not defined in `service_definitions`
    [
      "${local.db_user_prefix}hasura",
      "redash",
      "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam",
    ],

    # Legacy hasura user. GA-only since it's only used in prod.ga.
    # See: https://manabie.slack.com/archives/C01TG8A97ME/p1686738802593919?thread_ts=1686735050.266879&cid=C01TG8A97ME
    local.org == "ga" && local.env == "prod" ? ["hasura"] : [],

    # Create hasura-v2 user
    [
      for service in local.service_definitions : [
        "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam",
      ] if try(service.postgresql.createdb, false) && (try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false))
    ]
  ))

  # `postgresql_statement_timeout` block defines timeout per roles.
  postgresql_statement_timeout = flatten(concat(
    [
      for service in local.service_definitions : [
        {
          user              = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
          statement_timeout = lookup(service.postgresql.statement_timeout, "timeout", "120s")
        }
      ] if !try(service.disable_iam, false) && try(service.postgresql.statement_timeout.enabled, false)
    ],

    # Create hasura-v2 user
    [
      for service in local.service_definitions : [
        {
          user              = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          statement_timeout = lookup(service.postgresql.statement_timeout, "timeout", "120s")
        }
      ] if try(service.postgresql.createdb, false) && (try(service.hasura.enabled, false) && try(service.postgresql.statement_timeout.enabled, false) || try(service.hasura.v2_enabled, false)) && try(service.postgresql.statement_timeout.enabled, false)
    ],

    [
      {
        user              = "${local.db_user_prefix}hasura"
        statement_timeout = "30s"
      }
    ],
  ))

  # `postgresql_user_permissions` defines the database permissions or privileges
  # to be granted to the database users.
  # Either use `service_definitions` or manipulate this array to add or remove grants.
  postgresql_user_permissions = flatten(concat(
    # Grant each service access the database (basic permissions).
    # If service.postgresql.grants[].grant_delete is true, DELETE privilege is also granted.
    [
      for service in local.service_definitions : [
        for grant in try(service.postgresql.grants, []) : [
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            schema      = "public"
            object_type = "table"
            privileges  = lookup(grant, "grant_delete", false) ? ["SELECT", "INSERT", "UPDATE", "DELETE"] : ["SELECT", "INSERT", "UPDATE"]
          },
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            schema      = "public"
            object_type = "sequence"
            privileges  = ["USAGE", "SELECT", "UPDATE"]
          },
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            schema      = "public"
            object_type = "function"
            privileges  = ["EXECUTE"]
          },
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            schema      = "public"
            object_type = "schema"
            privileges  = ["USAGE"]
          },
        ]
      ]
    ],

    # Grant hasura access to the database of the service whose `hasura.enabled` is true.
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.db_user_prefix}hasura"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.db_user_prefix}hasura"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.db_user_prefix}hasura"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.db_user_prefix}hasura"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        },
      ] if try(service.hasura.enabled, false)
    ],

    # Due to legacy reasons, prod.ga still uses `hasura` account, instead of `ga_hasura`.
    # Thus, this block grants permissions for `hasura` that `ga_hasura` would have
    # to make both accounts work.
    # See: https://manabie.slack.com/archives/C01TG8A97ME/p1686738802593919?thread_ts=1686735050.266879&cid=C01TG8A97ME
    #
    # This block should be identical to the previous block, except for `user` field.
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "hasura"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "hasura"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "hasura"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "hasura"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        },
      ] if try(service.hasura.enabled, false) && local.org == "ga" && local.env == "prod"
    ],

    # Grant hasura's SA access to the database of the service whose `hasura.enabled` is true.
    # For now, grant full permissions only for draft.
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = service.name != "draft" ? ["SELECT", "INSERT", "UPDATE"] : ["SELECT", "INSERT", "UPDATE", "DELETE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        }
      ] if try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false)
    ],

    # Grant hasura-v2 access to the database
    [
      for service in local.service_definitions : [
        # Access to the metadata database
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = ""
          object_type = "database"
          privileges  = ["CREATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        }
      ] if try(service.postgresql.createdb, false) && try(service.hasura.v2_enabled, false)
    ],

    # Grant kafka-connector access to the database of the service whose `kafka_enabled` is true.
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}kafka-connect@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = lookup(lookup(service, "kafka", {}), "grant_delete", false) ? ["SELECT", "INSERT", "UPDATE", "DELETE"] : ["SELECT", "INSERT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}kafka-connect@${local.project_id}.iam"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        },
      ] if try(service.kafka.enabled, false)
    ],

    # The following block contains the non-traditional permissions
    # which cannot be formulated in the above manner.
    [
      # Grant shamir read-only access to only certain tables in bob
      {
        database    = "${local.db_prefix}bob"
        user        = "${local.service_account_prefix}shamir@${local.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
        objects = [
          "users",
          "users_groups",
          "teachers",
          "school_admins",
          "students",
          "parents",
          "organization_auths",
          "api_keypair",
          "organizations",
        ]
      },
      {
        database    = "${local.db_prefix}bob"
        user        = "${local.service_account_prefix}shamir@${local.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },

      # Grant redash read-only access to certain databases
      {
        database    = "${local.db_prefix}bob"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}bob"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}eureka"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}eureka"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}tom"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}tom"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}fatima"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}fatima"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}invoicemgmt"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}invoicemgmt"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}calendar"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}calendar"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}entryexitmgmt"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}entryexitmgmt"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}lessonmgmt"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}lessonmgmt"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}timesheet"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}timesheet"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      {
        database    = "${local.db_prefix}mastermgmt"
        user        = "redash"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT"]
      },
      {
        database    = "${local.db_prefix}mastermgmt"
        user        = "redash"
        schema      = "public"
        object_type = "schema"
        privileges  = ["USAGE"]
      },
      # Grant unleash all privileges in its database.
      # Extra permission for tables/sequences/functions are also added,
      # because many objects are previously created using `postgres` account.
      {
        database    = "${local.db_prefix}unleash"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = ""
        object_type = "database"
        privileges  = ["CREATE", "CONNECT", "TEMPORARY"]
      },
      {
        database    = "${local.db_prefix}unleash"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"]
      },
      {
        database    = "${local.db_prefix}unleash"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "sequence"
        privileges  = ["USAGE", "SELECT", "UPDATE"]
      },
      {
        database    = "${local.db_prefix}unleash"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "function"
        privileges  = ["EXECUTE"]
      },
      {
        database    = "${local.db_prefix}unleash"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["CREATE", "USAGE"]
      },

      # Grant unleash all privileges in its database unleashv2.
      # Extra permission for tables/sequences/functions are also added,
      # because many objects are previously created using `postgres` account.
      {
        database    = "${local.db_prefix}unleashv2"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = ""
        object_type = "database"
        privileges  = ["CREATE", "CONNECT", "TEMPORARY"]
      },
      {
        database    = "${local.db_prefix}unleashv2"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "table"
        privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"]
      },
      {
        database    = "${local.db_prefix}unleashv2"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "sequence"
        privileges  = ["USAGE", "SELECT", "UPDATE"]
      },
      {
        database    = "${local.db_prefix}unleashv2"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "function"
        privileges  = ["EXECUTE"]
      },
      {
        database    = "${local.db_prefix}unleashv2"
        user        = "${local.service_account_prefix}unleash@${local.project_id}.iam"
        schema      = "public"
        object_type = "schema"
        privileges  = ["CREATE", "USAGE"]
      },
    ],

    # Grant ad-hoc serviceaccount access to all service databases
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        },
      ] if try(service.postgresql.createdb, false) && service.name != "nats" # ignore nats since it's a legacy db
    ]
  ))

  postgresql_bypass_rls_roles = flatten(concat(
    [
      "${local.db_user_prefix}hasura",
      "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam",
    ],
    [
      for service in local.service_definitions : [
        "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam",
      ] if try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false)
    ],
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
      if try(service.postgresql.bypassrls, false)
    ]
  ))

  postgresql_replication_roles = [
    "${local.service_account_prefix}kafka-connect@${local.project_id}.iam",
  ]

  identity_namespaces = {
    "services" : "${local.env}-${local.org}-services",
    "backend" : "${local.env}-${local.org}-backend",
    "nats-jetstream" : "${local.env}-${local.org}-nats-jetstream",
    "machine-learning" : "${local.env}-${local.org}-machine-learning",
    "elastic" : try(local.config.elasticsearch_identity_namespace, "${local.env}-${local.org}-elastic"),
    "kafka" : try(local.config.kafka_identity_namespace, "${local.env}-${local.org}-kafka"),
    "unleash" : try(local.config.unleash_identity_namespace, "${local.env}-${local.org}-unleash"),
    "preproduction-services" : "dorp-${local.org}-services",
    "preproduction-backend" : "dorp-${local.org}-backend",
    "preproduction-nats-jetstream" : "dorp-${local.org}-nats-jetstream",
    "preproduction-elastic" : "dorp-${local.org}-elastic",
    "preproduction-kafka" : "dorp-${local.org}-kafka",
    "preproduction-unleash" : "dorp-${local.org}-unleash",
    "data-warehouse" : try(local.config.data_warehouse_identity_namespace, "${local.env}-${local.org}-data-warehouse"),
    "frontend" : "${local.env}-${local.org}-frontend",
    "preproduction-frontend" : "dorp-${local.org}-frontend",
    "appsmith" : "${local.env}-${local.org}-appsmith",
    "preproduction-data-warehouse" : "dorp-${local.org}-data-warehouse",
  }

  service_accounts = flatten(concat(
    [
      for service in local.service_definitions :
      {
        name    = "${local.service_account_prefix}${service.name}"
        project = local.project_id
        roles = local.project_id == local.runs_on_project_id ? {
          "${local.project_id}" = distinct(concat(
            [
              for role in service.iam_roles : replace(role, "local.runs_on_project_id", "${local.runs_on_project_id}")
            ],
            [
              for role in service.run_on_project_iam_roles : replace(role, "local.runs_on_project_id", "${local.runs_on_project_id}")
            ]
          ))
          } : {
          "${local.project_id}" = [
            for role in service.iam_roles : replace(role, "local.runs_on_project_id", "${local.runs_on_project_id}")
          ],
          "${local.runs_on_project_id}" = [
            for role in service.run_on_project_iam_roles : replace(role, "local.runs_on_project_id", "${local.runs_on_project_id}")
          ]
        }
        identity_namespaces = [for identity_namespace in service.identity_namespaces : "${local.identity_namespaces[identity_namespace]}"]
        bucket_roles        = try(service.bucket_roles["${local.org}"], {})
        # Service's pod need impersonate of hasura's SA to login cloud-proxy in `hasura-migration` step.
        impersonations = flatten(concat(
          try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false) ? [
            {
              name    = "${local.service_account_prefix}${service.name}-h"
              project = local.project_id
              role    = "roles/iam.serviceAccountTokenCreator"
            }
          ] : [],

          # Service `shamir` need to impersonate `auth-m` to connect to `auth` database to run the migration.
          # TODO(bao): remove this once we replace `shamir` to `auth` service later.
          service == "shamir" ? [
            {
              name    = "${local.service_account_prefix}auth-m"
              project = local.project_id
              role    = "roles/iam.serviceAccountTokenCreator"
            }
          ] : [],
        ))
      } if !try(service.disable_iam, false) && !contains(["usermgmt"], service.name)
    ],
    [
      {
        name    = "${local.service_account_prefix}kafka"
        project = local.project_id
        roles = local.project_id == local.runs_on_project_id ? {
          "${local.project_id}" = [
            "roles/cloudsql.client",
            "roles/bigquery.dataOwner",
          ],
          } : {
          "${local.project_id}" = [
            "roles/cloudsql.client",
          ],
          "${local.runs_on_project_id}" = [
            "roles/cloudsql.client",
            "roles/bigquery.dataOwner",
          ]
        }
        identity_namespaces = [
          local.identity_namespaces["kafka"],
          local.identity_namespaces["preproduction-kafka"],
        ]
      },
      {
        name    = "${local.service_account_prefix}usermgmt"
        project = local.project_id
        roles = local.project_id == local.runs_on_project_id ? {
          "${local.project_id}" = [
            "roles/firebaseauth.admin",
            "roles/iam.serviceAccountTokenCreator",
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/cloudprofiler.agent",
          ]
          } : {
          "${local.project_id}" = [
            "roles/firebaseauth.admin",
            "roles/iam.serviceAccountTokenCreator",
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
          "${local.runs_on_project_id}" = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/cloudprofiler.agent",
            "roles/iam.serviceAccountTokenCreator",
            "projects/${local.runs_on_project_id}/roles/CustomRole447",
          ]
        }
        bucket_roles        = lookup(try(local.service_definitions[index(local.service_definitions.*.name, "usermgmt")].bucket_roles, {}), "${local.org}", {})
        identity_namespaces = [
          local.identity_namespaces["services"],
          local.identity_namespaces["backend"],
          local.identity_namespaces["preproduction-services"],
          local.identity_namespaces["preproduction-backend"],
        ]
        impersonations = local.org != "jprep" && local.org != "tokyo" ? [
          {
            # Impersonate the "${local.service_account_prefix}${local.org}-usermgmt" below.
            name    = "${local.service_account_prefix}${local.org}-usermgmt"
            project = local.runs_on_project_id
            role    = "roles/iam.serviceAccountTokenCreator"
          }
        ] : []
      },
      {
        # JPREP doesn't need this service account
        name = local.org != "jprep" && local.org != "tokyo" ? "${local.service_account_prefix}${local.org}-usermgmt" : null

        # Note: this is using local.runs_on_project_id, while Staging and UAT are
        # using local.project_id, that's because for Production, Firebase project
        # and Identity Platform project are different. That is, all partner's
        # Firebase project is the same with its GCP project, while Identity Platform
        # project is always student-coach-e1e95, and it is the same for all of them.
        #
        # So to make this usermgmt service account be able to create custom token for
        # Identity Platform in that student-coach-e1e95 project, it must belongs to
        # that project also.
        #
        # See https://docs.google.com/document/d/1BFkzkADGPnAXWnoccYgeN1O-Tr94LjlBi1mMJzia2MA/edit#heading=h.oerfmxd5lar1
        # for more details.
        project             = local.runs_on_project_id
        roles               = {}
        identity_namespaces = []
      },
      {
        # service account for ad-hoc tasks
        name    = "${local.env}-${local.org}-ad-hoc"
        project = local.runs_on_project_id
        roles = {
          "${local.runs_on_project_id}" = [
            "roles/artifactregistry.reader", # allow adhoc to pull custom images to run cloud build. it should be granted in learner project.
            "roles/container.clusterViewer",
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/logging.logWriter",
          ],
        }
        identity_namespaces = []
      },
    ],

    # Create SA for hasura v1 & v2
    flatten([
      for service in local.service_definitions : [
        {
          name    = "${local.service_account_prefix}${service.name}-h"
          project = local.project_id
          roles = {
            "${local.project_id}" = [
              "roles/cloudsql.client",
              "roles/cloudsql.instanceUser",
              "roles/alloydb.client",
            ],
            "${local.runs_on_project_id}" = [
              "roles/cloudsql.client",
              "roles/cloudsql.instanceUser",
              "roles/alloydb.client",
            ]
          }
          identity_namespaces = [
            local.identity_namespaces["services"],
            local.identity_namespaces["backend"],
          ]
        },
      ] if try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false)
    ])
  ))

  gke_rbac_enabled = try(local.config.gke_rbac_enabled, true)
  gke_rbac_policies = local.gke_rbac_enabled ? [
    {
      kind      = "Group"
      group     = "tech-func-cse@manabie.com"
      role_kind = "ClusterRole"
      role_name = "view"
      namespaces = [
        local.identity_namespaces["services"],
        local.identity_namespaces["backend"],
        local.identity_namespaces["nats-jetstream"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["kafka"],
        local.identity_namespaces["unleash"],
        local.identity_namespaces["appsmith"],
        local.identity_namespaces["data-warehouse"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-func-backend@manabie.com"
      role_kind = "ClusterRole"
      role_name = "view"
      namespaces = [
        local.identity_namespaces["services"],
        local.identity_namespaces["backend"],
        local.identity_namespaces["nats-jetstream"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["kafka"],
        local.identity_namespaces["unleash"],
        local.identity_namespaces["appsmith"],
        local.identity_namespaces["data-warehouse"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-platform@manabie.com"
      role_kind = "ClusterRole"
      role_name = "view"
      namespaces = [
        local.identity_namespaces["services"],
        local.identity_namespaces["backend"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["unleash"],
        local.identity_namespaces["appsmith"],
        local.identity_namespaces["data-warehouse"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-platform@manabie.com"
      role_kind = "ClusterRole"
      role_name = "cluster-admin"
      namespaces = [
        local.identity_namespaces["preproduction-nats-jetstream"],
      ]
    },
    # This "custom-admin" role has almost the same permissions as
    # predefined "admin" role, except that it doesn't have permission
    # to delete some resources, such as deployments, statefulsets, secrets...etc.
    # We can use this role to prevent accidental deletion of resources.
    {
      kind      = "Group"
      group     = "tech-squad-platform@manabie.com"
      role_kind = "ClusterRole"
      role_name = "custom-admin"
      namespaces = [
        local.identity_namespaces["nats-jetstream"],
        local.identity_namespaces["kafka"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-platform@manabie.com"
      role_kind = "ClusterRole"
      role_name = "pods-port-forwarder"
      namespaces = [
        local.identity_namespaces["preproduction-nats-jetstream"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["preproduction-elastic"],
        local.identity_namespaces["preproduction-kafka"],
        local.identity_namespaces["data-warehouse"],
        local.identity_namespaces["preproduction-data-warehouse"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-architecture@manabie.com"
      role_kind = "ClusterRole"
      role_name = "cluster-admin"
      namespaces = [
        local.identity_namespaces["appsmith"],
        local.identity_namespaces["preproduction-elastic"],
        local.identity_namespaces["preproduction-kafka"],
        local.identity_namespaces["preproduction-data-warehouse"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-architecture@manabie.com"
      role_kind = "ClusterRole"
      role_name = "custom-admin"
      namespaces = [
        local.identity_namespaces["elastic"],
        local.identity_namespaces["data-warehouse"],
        local.identity_namespaces["kafka"],
        local.identity_namespaces["nats-jetstream"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-automation@manabie.com"
      role_kind = "ClusterRole"
      role_name = "cluster-admin"
      namespaces = [
        local.identity_namespaces["unleash"],
        local.identity_namespaces["preproduction-unleash"],
      ]
    },
    {
      kind      = "User"
      group     = "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam.gserviceaccount.com"
      role_kind = "ClusterRole"
      role_name = "cluster-admin"
      namespaces = [
        local.identity_namespaces["services"],
        local.identity_namespaces["backend"],
        local.identity_namespaces["nats-jetstream"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["kafka"],
        local.identity_namespaces["unleash"],
        local.identity_namespaces["preproduction-services"],
        local.identity_namespaces["preproduction-backend"],
        local.identity_namespaces["preproduction-nats-jetstream"],
        local.identity_namespaces["preproduction-elastic"],
        local.identity_namespaces["preproduction-kafka"],
        local.identity_namespaces["data-warehouse"],
      ]
    },
  ] : []

  kms_key_name = try(local.config.kms_key_name, "")
  kms_key_props = {
    rotation_period = try(local.config.kms_key_rotation_period, "2592000s")
    owner           = "group:dev.infras@manabie.com"
    encrypter       = "${local.service_account_prefix}${local.org}-kms-encrypter"
    decrypters = [
      for sa in local.service_accounts : {
        service_account_project = sa.project
        service_account_name    = sa.name
      }
    ]
  }

  platforms_gke        = try(local.config.platforms_gke, "../platforms")
  platforms_kms        = try(local.config.platforms_kms, "../platforms")
  platforms_storage    = try(local.config.platforms_storage, "../platforms")
  platforms_postgresql = try(local.config.platforms_postgresql, "../platforms")
}

dependency "gke" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_gke}"
}

dependency "kms" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_kms}"
}

dependency "storage" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_storage}"
}

dependency "postgresql" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_postgresql}"
}

inputs = {
  postgresql_instance_port = local.postgresql_port_config.locals.postgresql_instance_port
  postgresql = {
    project_id = dependency.postgresql.outputs.postgresql_project
    instance   = dependency.postgresql.outputs.postgresql_instance
    databases  = local.databases
    users      = local.users
  }
  postgresql_user_permissions  = local.postgresql_user_permissions
  postgresql_bypass_rls_roles  = local.postgresql_bypass_rls_roles
  postgresql_replication_roles = local.postgresql_replication_roles
  postgresql_statement_timeout = local.postgresql_statement_timeout

  gke_endpoint           = try(dependency.gke.outputs.gke_endpoint, "")
  gke_ca_cert            = try(dependency.gke.outputs.gke_ca_cert, "")
  gke_identity_namespace = try(dependency.gke.outputs.gke_identity_namespace, "")
  service_accounts       = local.service_accounts

  gke_rbac = {
    enabled  = local.gke_rbac_enabled
    policies = local.gke_rbac_policies
  }

  kms_keys = local.kms_key_name != "" ? {
    "${local.kms_key_name}" = merge(
      local.kms_key_props,
      {
        key_ring = try(dependency.kms.outputs.kms_key_ring, "")
      },
    )
  } : {}

  # We don't have to set cloudconvert resource if the dependency platforms doesn't have any bucket
  cloudconvert = try(dependency.storage.outputs.backend_bucket, "") == "" ? null : {
    service_account = "${local.service_account_prefix}cloudconvert"
    bucket          = dependency.storage.outputs.backend_bucket
  }
}
