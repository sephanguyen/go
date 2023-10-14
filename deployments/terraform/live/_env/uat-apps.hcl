locals {
  config   = read_terragrunt_config("config.hcl").locals
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl"))

  postgresql_port_config = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/postgresql-port.hcl")

  project_id             = local.env_vars.locals.project_id
  runs_on_project_id     = local.config.runs_on_project_id
  service_account_prefix = local.config.service_account_prefix

  org = local.config.org
  env = local.env_vars.locals.env

  adhoc_db_user = "${local.env}-${local.org}-ad-hoc@${local.runs_on_project_id}.iam"

  db_prefix      = try(local.config.db_prefix, "")
  db_user_prefix = try(local.config.db_user_prefix, "")

  # Import the global service defintions
  service_definitions = yamldecode(file("${get_repo_root()}/deployments/decl/uat-defs.yaml"))

  # Import all GKE RBAC policies
  rbac_policies = read_terragrunt_config("${get_terragrunt_dir()}/../../_env/gke-rbac-roles.hcl").locals.rbac_policies

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

    # this user is used for migration database.
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
      if try(service.postgresql.createdb, false) && !contains(["nats"], service.name)
    ],

    # The following block contains other users not defined in `service_definitions`
    [
      "redash",
      local.adhoc_db_user,
    ],

    # Create hasura-v2 user
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
      if try(service.postgresql.createdb, false) && (try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false))
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
            owner       = "${local.service_account_prefix}${grant.dbname}-m@${local.project_id}.iam"
            schema      = "public"
            object_type = "table"
            privileges  = lookup(grant, "grant_delete", false) ? ["SELECT", "INSERT", "UPDATE", "DELETE"] : ["SELECT", "INSERT", "UPDATE"]
          },
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            owner       = "${local.service_account_prefix}${grant.dbname}-m@${local.project_id}.iam"
            schema      = "public"
            object_type = "sequence"
            privileges  = ["USAGE", "SELECT", "UPDATE"]
          },
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            owner       = "${local.service_account_prefix}${grant.dbname}-m@${local.project_id}.iam"
            schema      = "public"
            object_type = "function"
            privileges  = ["EXECUTE"]
          },
          {
            database    = "${local.db_prefix}${grant.dbname}"
            user        = "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
            owner       = "${local.service_account_prefix}${grant.dbname}-m@${local.project_id}.iam"
            schema      = "public"
            object_type = "schema"
            privileges  = ["USAGE"]
          },
        ]
      ]
    ],

    # Grant user migration-database all privileges in its databases. 
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = ""
          object_type = "database"
          privileges  = ["CREATE", "CONNECT", "TEMPORARY"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "schema"
          privileges  = ["CREATE", "USAGE"]
        },
      ] if try(service.postgresql.createdb, false) && !contains(["nats"], service.name)
    ],

    # Grant hasura's SA access to the database of the service whose `hasura.enabled` is true.
    # For now, grant full permissions only for draft.
    [
      for service in local.service_definitions : [
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = ""
          object_type = "database"
          privileges  = ["CREATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = service.name != "draft" ? ["SELECT", "INSERT", "UPDATE"] : ["SELECT", "INSERT", "UPDATE", "DELETE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
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
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = ""
          object_type = "database"
          privileges  = ["CREATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}_hasura_metadata"
          user        = "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
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
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = lookup(lookup(service, "kafka", {}), "grant_delete", false) ? ["SELECT", "INSERT", "UPDATE", "DELETE"] : ["SELECT", "INSERT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = "${local.service_account_prefix}kafka-connect@${local.project_id}.iam"
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
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
        owner       = "${local.service_account_prefix}bob-m@${local.project_id}.iam"
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
        owner       = "${local.service_account_prefix}bob-m@${local.project_id}.iam"
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
          user        = local.adhoc_db_user
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "table"
          privileges  = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = local.adhoc_db_user
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "sequence"
          privileges  = ["USAGE", "SELECT", "UPDATE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = local.adhoc_db_user
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "function"
          privileges  = ["EXECUTE"]
        },
        {
          database    = "${local.db_prefix}${service.name}"
          user        = local.adhoc_db_user
          owner       = "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
          schema      = "public"
          object_type = "schema"
          privileges  = ["USAGE"]
        },
        # ignore nats since it's a legacy db
      ] if try(service.postgresql.createdb, false) && service.name != "nats"
    ]
  ))

  postgresql_bypass_rls_roles = flatten(concat(
    [
      local.adhoc_db_user,
    ],
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}-h@${local.project_id}.iam"
      if try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false)
    ],
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}@${local.project_id}.iam"
      if try(service.postgresql.bypassrls, false)
    ],
    [
      for service in local.service_definitions : "${local.service_account_prefix}${service.name}-m@${local.project_id}.iam"
      if try(service.postgresql.createdb, false) && !contains(["nats"], service.name)
    ],
  ))

  postgresql_replication_roles = [
    "${local.service_account_prefix}kafka-connect@${local.project_id}.iam",
  ]

  identity_namespaces = {
    "services" : "${local.env}-${local.org}-services",
    "nats-jetstream" : "${local.env}-${local.org}-nats-jetstream",
    "machine-learning" : "${local.env}-${local.org}-machine-learning",
    "elastic" : try(local.config.elasticsearch_identity_namespace, "${local.env}-${local.org}-elastic"),
    "kafka" : try(local.config.kafka_identity_namespace, "${local.env}-${local.org}-kafka"),
    "unleash" : try(local.config.unleash_identity_namespace, "${local.env}-${local.org}-unleash"),
    "appsmith" : "${local.env}-${local.org}-appsmith",
    "frontend" : "${local.env}-${local.org}-frontend",
    "backend" : "${local.env}-${local.org}-backend",
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
        bucket_roles        = lookup(try(service.bucket_roles, {}), "${local.org}", {})

        # Service's pod need impersonate of hasura's SA to login cloud-proxy in `hasura-migration` step.
        # Service's pod need impersonate of `-m@...` SA to login cloud-proxy in `database-migration` step.
        impersonations = flatten(concat(
          try(service.hasura.enabled, false) || try(service.hasura.v2_enabled, false) ? [
            {
              name    = "${local.service_account_prefix}${service.name}-h"
              project = local.project_id
              role    = "roles/iam.serviceAccountTokenCreator"
            }
          ] : [],
          try(service.postgresql.createdb, false) && !contains(["nats"], service.name) ? [
            {
              name    = "${local.service_account_prefix}${service.name}-m"
              project = local.project_id
              role    = "roles/iam.serviceAccountTokenCreator"
            }
          ] : [],
          # Service `shamir` need to impersonate `auth-m` to connect to `auth` database to run the migration.
          # TODO(bao): remove this once we replace `shamir` to `auth` service later.
          service.name == "shamir" ? [
            {
              name    = "${local.service_account_prefix}auth-m"
              project = local.project_id
              role    = "roles/iam.serviceAccountTokenCreator"
            },
          ] : [],
        ))
      } if !try(service.disable_iam, false) && !contains(["usermgmt"], service.name)
    ],

    # create serviceaccount for migration database job.
    [
      for service in local.service_definitions :
      {
        name    = "${local.service_account_prefix}${service.name}-m"
        project = local.project_id
        roles = {
          "${local.project_id}" = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ],
          "${local.runs_on_project_id}" = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
        }
        identity_namespaces = [
          local.identity_namespaces["services"],
          local.identity_namespaces["backend"],
        ]
      } if try(service.postgresql.createdb, false) && !contains(["nats"], service.name)
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
            "roles/identityplatform.admin",
            "roles/iam.serviceAccountTokenCreator",
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
          ]
          "${local.runs_on_project_id}" = [
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/cloudprofiler.agent",
            "projects/${local.runs_on_project_id}/roles/CustomRole447",
          ]
        }
        bucket_roles        = lookup(try(local.service_definitions[index(local.service_definitions.*.name, "usermgmt")].bucket_roles, {}), "${local.org}", {})
        identity_namespaces = [
          local.identity_namespaces["services"],
          local.identity_namespaces["backend"],
        ]
        impersonations = local.org != "jprep" ? [
          {
            # Impersonate the "${local.service_account_prefix}${local.org}-usermgmt" below.
            name    = "${local.service_account_prefix}${local.org}-usermgmt"
            project = local.project_id
            role    = "roles/iam.serviceAccountTokenCreator"
          }
        ] : []
      },
      {
        # JPREP doesn't need this service account
        name = local.org != "jprep" ? "${local.service_account_prefix}${local.org}-usermgmt" : null

        # Note: this is using local.project_id, while prod is using
        # local.runs_on_project_id, that's because for UAT Manabie,
        # both Firebase project and Identity Platform project are
        # using the the same uat-manabie project.
        project             = local.project_id
        roles               = {}
        identity_namespaces = []
      },
      {
        # service account for ad-hoc tasks
        name    = "${local.env}-${local.org}-ad-hoc"
        project = local.runs_on_project_id
        roles = {
          "${local.runs_on_project_id}" = [
            "roles/container.clusterViewer",
            "roles/cloudsql.client",
            "roles/cloudsql.instanceUser",
            "roles/logging.logWriter",
          ],
          "student-coach-e1e95" = [
            "roles/artifactregistry.reader", # allow adhoc to pull custom images to run cloud build
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

  gke_rbac_enabled = try(local.config.gke_rbac_enabled, false)
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
      ]
    },
    {
      kind      = "Group"
      group     = "tech-func-backend@manabie.com"
      role_kind = "ClusterRole"
      role_name = "custom-admin"
      # Cluster role "custom-admin" is defined in the platforms module.
      # It has the same permissions as cluster-admin, except that it
      # doesn't have permission to delete some resources, such as
      # deployments, statefulsets, secrets...etc.
      namespaces = [
        local.identity_namespaces["services"],
        local.identity_namespaces["backend"],
        local.identity_namespaces["nats-jetstream"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["kafka"],
        local.identity_namespaces["unleash"],
        local.identity_namespaces["appsmith"],
      ]
    },
    {
      kind      = "Group"
      group     = "tech-squad-platform@manabie.com"
      role_kind = "ClusterRole"
      role_name = "custom-admin"
      # Cluster role "custom-admin" is defined in the platforms module.
      # It has the same permissions as cluster-admin, except that it
      # doesn't have permission to delete some resources, such as
      # deployments, statefulsets, secrets...etc.
      namespaces = [
        local.identity_namespaces["services"],
        local.identity_namespaces["backend"],
        local.identity_namespaces["nats-jetstream"],
        local.identity_namespaces["elastic"],
        local.identity_namespaces["kafka"],
        local.identity_namespaces["unleash"],
        local.identity_namespaces["appsmith"],
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
        local.identity_namespaces["appsmith"],
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

  platforms_gke_dep        = try(local.config.platforms_gke_dep, "../platforms")
  platforms_kms_dep        = try(local.config.platforms_kms_dep, "../platforms")
  platforms_storage_dep    = try(local.config.platforms_storage_dep, "../platforms")
  platforms_postgresql_dep = try(local.config.platforms_postgresql_dep, "../platforms")
}

dependency "platforms_gke" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_gke_dep}"
}

dependency "platforms_kms" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_kms_dep}"
}

dependency "platforms_storage" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_storage_dep}"
}

dependency "platforms_postgresql" {
  config_path = "${get_terragrunt_dir()}/${local.platforms_postgresql_dep}"
}

inputs = {
  postgresql_instance_port = local.postgresql_port_config.locals.postgresql_instance_port
  postgresql = {
    project_id = dependency.platforms_postgresql.outputs.postgresql_project
    instance   = dependency.platforms_postgresql.outputs.postgresql_instance
    databases  = local.databases
    users      = local.users
  }
  postgresql_user_permissions  = local.postgresql_user_permissions
  postgresql_bypass_rls_roles  = local.postgresql_bypass_rls_roles
  postgresql_replication_roles = local.postgresql_replication_roles
  postgresql_statement_timeout = local.postgresql_statement_timeout
  adhoc = {
    grant_enabled = true
    dbuser        = local.adhoc_db_user
  }

  gke_endpoint           = try(dependency.platforms_gke.outputs.gke_endpoint, "")
  gke_ca_cert            = try(dependency.platforms_gke.outputs.gke_ca_cert, "")
  gke_identity_namespace = try(dependency.platforms_gke.outputs.gke_identity_namespace, "")
  service_accounts       = local.service_accounts

  kms_keys = local.kms_key_name != "" ? {
    "${local.kms_key_name}" = merge(
      local.kms_key_props,
      {
        key_ring = try(dependency.platforms_kms.outputs.kms_key_ring, "")
      },
    )
  } : {}

  cloudconvert = {
    service_account = "${local.service_account_prefix}cloudconvert"
    bucket          = dependency.platforms_storage.outputs.backend_bucket
  }

  gke_rbac = {
    enabled  = local.gke_rbac_enabled
    policies = local.gke_rbac_policies
  }

  rbac_roles = {
    enabled = local.gke_rbac_enabled
    # See explanation in gke-rbac-roles.hcl.
    policies = {
      for level, envs in local.rbac_policies :
      "${level}" => {
        for env, policies in envs :
        "${env}" => [
          for policy in policies : merge(
            policy,
            {
              # On UAT we are using the same "elastic" namespace for both
              # Manabie and JPREP, so we need to exclude that namespace for
              # Manabie (or JPREP, it doesn't matter, just one of them),
              # otherwise this will try to create the same role binding to
              # that "elastic" namespace, which will cause duplicated error.
              namespaces = local.org == "manabie" ? [
                for ns in policy.namespaces : local.identity_namespaces[ns]
                if ns != "elastic"
              ] : [for ns in policy.namespaces : local.identity_namespaces[ns]]
            }
          )
        ]
        if env == local.env
      }
    }
  }
}
