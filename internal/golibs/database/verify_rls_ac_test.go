package database

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/cmd/utils/rls"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const tableStageJSONStr = `{
	"filename": "accesscontrol/mastermgmt/archtechture.ac_test_template_11_4.yaml",
	"service": "mastermgmt",
	"revision": 2,
	"table_name": "ac_test_template_11_4",
	"created_at": "2022-10-13T12:22:22.795753037+07:00",
	"updated_at": "2022-10-13T12:22:22.795753066+07:00",
	"stages": [
		{
			"template": "1.1",
			"hasura": null,
			"postgres": null,
			"accessPathTable": null,
			"locationCol": null,
			"permissionPrefix": null,
			"permissions": null,
			"ownerCol": null,
			"use_custom_policy": true,
			"hasura_policies": null,
			"postgres_policies": [
				{
					"name": "rls_ac_test_template_11_4_select_location",
					"using": "true \u003c= (\n  select\t\t\t\n    true\n  from\n          granted_permissions p\n  join ac_test_template_11_4_access_paths usp on\n          usp.location_id = p.location_id\n  where\n    p.user_id = current_setting('app.user_id')\n    and p.permission_name = 'accesscontrol.ac_test_template_11_4.read'\n    and usp.\"ac_test_template_11_4_id\" = ac_test_template_11_4.ac_test_template_11_4_id\n  limit 1\n  )\n",
					"with_check": "",
					"for": "select"
				},
				{
					"name": "rls_ac_test_template_11_4_insert_location",
					"using": "",
					"with_check": "(\n  1 = 1\n)\n",
					"for": "insert"
				},
				{
					"name": "rls_ac_test_template_11_4_update_location",
					"using": "true \u003c= (\n  select\t\t\t\n    true\n  from\n          granted_permissions p\n  join ac_test_template_11_4_access_paths usp on\n          usp.location_id = p.location_id\n  where\n    p.user_id = current_setting('app.user_id')\n    and p.permission_name = 'accesscontrol.ac_test_template_11_4.write'\n    and usp.\"ac_test_template_11_4_id\" = ac_test_template_11_4.ac_test_template_11_4_id\n  limit 1\n  )\n",
					"with_check": "true \u003c= (\n  select\t\t\t\n    true\n  from\n          granted_permissions p\n  join ac_test_template_11_4_access_paths usp on\n          usp.location_id = p.location_id\n  where\n    p.user_id = current_setting('app.user_id')\n    and p.permission_name = 'accesscontrol.ac_test_template_11_4.write'\n    and usp.\"ac_test_template_11_4_id\" = ac_test_template_11_4.ac_test_template_11_4_id\n  limit 1\n  )\n",
					"for": "update"
				},
				{
					"name": "rls_ac_test_template_11_4_delete_location",
					"using": "true \u003c= (\n  select\t\t\t\n    true\n  from\n          granted_permissions p\n  join ac_test_template_11_4_access_paths usp on\n          usp.location_id = p.location_id\n  where\n    p.user_id = current_setting('app.user_id')\n    and p.permission_name = 'accesscontrol.ac_test_template_11_4.write'\n    and usp.\"ac_test_template_11_4_id\" = ac_test_template_11_4.ac_test_template_11_4_id\n  limit 1\n  )\n",
					"with_check": "",
					"for": "delete"
				},
				{
					"name": "rls_ac_test_template_11_4_permission_v4",
					"using": "current_setting('app.user_id') = owners\n",
					"with_check": "",
					"for": "all"
				}
			]
		}
	]
}`

const acTestTemplate114SchemaJSONStr = `{
	"schema": [
		{
			"column_name": "ac_test_template_11_4_id",
			"data_type": "text",
			"column_default": null,
			"is_nullable": "NO"
		},
		{
			"column_name": "created_at",
			"data_type": "timestamp with time zone",
			"column_default": null,
			"is_nullable": "NO"
		},
		{
			"column_name": "deleted_at",
			"data_type": "timestamp with time zone",
			"column_default": null,
			"is_nullable": "YES"
		},
		{
			"column_name": "name",
			"data_type": "text",
			"column_default": null,
			"is_nullable": "YES"
		},
		{
			"column_name": "owners",
			"data_type": "text",
			"column_default": null,
			"is_nullable": "NO"
		},
		{
			"column_name": "resource_path",
			"data_type": "text",
			"column_default": "autofillresourcepath()",
			"is_nullable": "NO"
		},
		{
			"column_name": "updated_at",
			"data_type": "timestamp with time zone",
			"column_default": null,
			"is_nullable": "NO"
		}
	],
	"policies": [
		{
			"tablename": "ac_test_template_11_4",
			"policyname": "rls_ac_test_template_11_4_delete_location",
			"qual": "(true \u003c= ( SELECT true AS bool\n   FROM (granted_permissions p\n     JOIN ac_test_template_11_4_access_paths usp ON ((usp.location_id = p.location_id)))\n  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_name = 'accesscontrol.ac_test_template_11_4.write'::text) AND (usp.ac_test_template_11_4_id = ac_test_template_11_4.ac_test_template_11_4_id))\n LIMIT 1))",
			"with_check": null,
			"relrowsecurity": true,
			"relforcerowsecurity": true,
			"permissive": "PERMISSIVE",
			"roles": {
				"Elements": [
					"public"
				],
				"Dimensions": [
					{
						"Length": 1,
						"LowerBound": 1
					}
				],
				"Status": 2
			}
		},
		{
			"tablename": "ac_test_template_11_4",
			"policyname": "rls_ac_test_template_11_4_insert_location",
			"qual": null,
			"with_check": "(1 = 1)",
			"relrowsecurity": true,
			"relforcerowsecurity": true,
			"permissive": "PERMISSIVE",
			"roles": {
				"Elements": [
					"public"
				],
				"Dimensions": [
					{
						"Length": 1,
						"LowerBound": 1
					}
				],
				"Status": 2
			}
		},
		{
			"tablename": "ac_test_template_11_4",
			"policyname": "rls_ac_test_template_11_4_permission_v4",
			"qual": "(current_setting('app.user_id'::text) = owners)",
			"with_check": "(current_setting('app.user_id'::text) = owners)",
			"relrowsecurity": true,
			"relforcerowsecurity": true,
			"permissive": "PERMISSIVE",
			"roles": {
				"Elements": [
					"public"
				],
				"Dimensions": [
					{
						"Length": 1,
						"LowerBound": 1
					}
				],
				"Status": 2
			}
		},
		{
			"tablename": "ac_test_template_11_4",
			"policyname": "rls_ac_test_template_11_4_restrictive",
			"qual": "permission_check(resource_path, 'ac_test_template_11_4'::text)",
			"with_check": "permission_check(resource_path, 'ac_test_template_11_4'::text)",
			"relrowsecurity": true,
			"relforcerowsecurity": true,
			"permissive": "RESTRICTIVE",
			"roles": {
				"Elements": [
					"public"
				],
				"Dimensions": [
					{
						"Length": 1,
						"LowerBound": 1
					}
				],
				"Status": 2
			}
		},
		{
			"tablename": "ac_test_template_11_4",
			"policyname": "rls_ac_test_template_11_4_select_location",
			"qual": "(true \u003c= ( SELECT true AS bool\n   FROM (granted_permissions p\n     JOIN ac_test_template_11_4_access_paths usp ON ((usp.location_id = p.location_id)))\n  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_name = 'accesscontrol.ac_test_template_11_4.read'::text) AND (usp.ac_test_template_11_4_id = ac_test_template_11_4.ac_test_template_11_4_id))\n LIMIT 1))",
			"with_check": null,
			"relrowsecurity": true,
			"relforcerowsecurity": true,
			"permissive": "PERMISSIVE",
			"roles": {
				"Elements": [
					"public"
				],
				"Dimensions": [
					{
						"Length": 1,
						"LowerBound": 1
					}
				],
				"Status": 2
			}
		},
		{
			"tablename": "ac_test_template_11_4",
			"policyname": "rls_ac_test_template_11_4_update_location",
			"qual": "(true \u003c= ( SELECT true AS bool\n   FROM (granted_permissions p\n     JOIN ac_test_template_11_4_access_paths usp ON ((usp.location_id = p.location_id)))\n  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_name = 'accesscontrol.ac_test_template_11_4.write'::text) AND (usp.ac_test_template_11_4_id = ac_test_template_11_4.ac_test_template_11_4_id))\n LIMIT 1))",
			"with_check": "(true \u003c= ( SELECT true AS bool\n   FROM (granted_permissions p\n     JOIN ac_test_template_11_4_access_paths usp ON ((usp.location_id = p.location_id)))\n  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_name = 'accesscontrol.ac_test_template_11_4.write'::text) AND (usp.ac_test_template_11_4_id = ac_test_template_11_4.ac_test_template_11_4_id))\n LIMIT 1))",
			"relrowsecurity": true,
			"relforcerowsecurity": true,
			"permissive": "PERMISSIVE",
			"roles": {
				"Elements": [
					"public"
				],
				"Dimensions": [
					{
						"Length": 1,
						"LowerBound": 1
					}
				],
				"Status": 2
			}
		}
	],
	"constraint": [
		{
			"constraint_name": "pk__ac_test_template_11_4",
			"column_name": "ac_test_template_11_4_id",
			"constraint_type": "PRIMARY KEY"
		}
	],
	"table_name": "ac_test_template_11_4"
}`

const metadataYAMLStr = `- table:
    schema: public
    name: ac_hasura_test_template_1
  object_relationships:
  - name: ac_hasura_test_template_1_location_permission
    using:
      manual_configuration:
        remote_table:
          schema: public
          name: granted_permissions
        column_mapping:
          location_id: location_id
  insert_permissions:
  - role: USER_GROUP_ADMIN
    permission:
      check:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.write
        - resource_path:
            _eq: X-Hasura-Resource-Path
      set:
        resource_path: x-hasura-Resource-Path
      columns:
      - ac_hasura_test_template_1_id
      - created_at
      - deleted_at
      - location_id
      - name
      - updated_at
  select_permissions:
  - role: USER_GROUP_ADMIN
    permission:
      columns:
      - ac_hasura_test_template_1_id
      - created_at
      - deleted_at
      - location_id
      - name
      - updated_at
      - resource_path
      filter:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.read
        - resource_path:
            _eq: X-Hasura-Resource-Path
      allow_aggregations: true
  - role: MANABIE
    permission:
      columns:
      - ac_hasura_test_template_1_id
      - created_at
      - deleted_at
      - location_id
      - name
      - updated_at
      - resource_path
      filter:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.read
        - resource_path:
            _eq: X-Hasura-Resource-Path
      allow_aggregations: true
  update_permissions:
  - role: USER_GROUP_ADMIN
    permission:
      check:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.write
        - resource_path:
            _eq: X-Hasura-Resource-Path
      columns:
      - ac_hasura_test_template_1_id
      - created_at
      - deleted_at
      - location_id
      - name
      - updated_at
      filter:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.write
        - resource_path:
            _eq: X-Hasura-Resource-Path
  delete_permissions:
  - role: USER_GROUP_ADMIN
    permission:
      check:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.write
        - resource_path:
            _eq: X-Hasura-Resource-Path
      filter:
        _and:
        - ac_hasura_test_template_1_location_permission:
            _and:
            - user_id:
                _eq: X-Hasura-User-Id
            - permission_name:
                _eq: accesscontrol.b.write
        - resource_path:
            _eq: X-Hasura-Resource-Path
`

const acTestHasuraTemplate1JSONstr = `{
	"filename": "accesscontrol/mastermgmt/archtechture.ac_hasura_test_template_1.yaml",
	"service": "mastermgmt",
	"revision": 1,
	"table_name": "ac_hasura_test_template_1",
	"created_at": "2022-10-13T12:22:22.795120409+07:00",
	"updated_at": "2022-10-13T12:22:22.795120468+07:00",
	"stages": [
		{
			"template": "1",
			"hasura": {
				"stage_dir": "deployments/helm/manabie-all-in-one/charts/mastermgmt/files/hasura/metadata/tables.yaml",
				"permissions": [
					"SELECT",
					"INSERT",
					"UPDATE",
					"DELETE"
				],
				"relationship": "ac_hasura_test_template_1_location_permission",
				"first_level_query": "",
				"hasura_policies": {
					"select_permission": [
						{
							"name": "USER_GROUP_ADMIN",
							"filter": {
								"ac_hasura_test_template_1_location_permission": {
									"_and": [
										{
											"user_id": {
												"_eq": "X-Hasura-User-Id"
											}
										},
										{
											"permission_name": {
												"_eq": "accesscontrol.b.read"
											}
										}
									]
								}
							}
						},
						{
							"name": "MANABIE",
							"filter": {
								"ac_hasura_test_template_1_location_permission": {
									"_and": [
										{
											"user_id": {
												"_eq": "X-Hasura-User-Id"
											}
										},
										{
											"permission_name": {
												"_eq": "accesscontrol.b.read"
											}
										}
									]
								}
							}
						}
					],
					"insert_permission": [
						{
							"name": "USER_GROUP_ADMIN",
							"check": {
								"ac_hasura_test_template_1_location_permission": {
									"_and": [
										{
											"user_id": {
												"_eq": "X-Hasura-User-Id"
											}
										},
										{
											"permission_name": {
												"_eq": "accesscontrol.b.write"
											}
										}
									]
								}
							}
						}
					],
					"delete_permission": [
						{
							"name": "USER_GROUP_ADMIN",
							"check": {
								"ac_hasura_test_template_1_location_permission": {
									"_and": [
										{
											"user_id": {
												"_eq": "X-Hasura-User-Id"
											}
										},
										{
											"permission_name": {
												"_eq": "accesscontrol.b.write"
											}
										}
									]
								}
							}
						}
					],
					"update_permission": [
						{
							"name": "USER_GROUP_ADMIN",
							"filter": {
								"ac_hasura_test_template_1_location_permission": {
									"_and": [
										{
											"user_id": {
												"_eq": "X-Hasura-User-Id"
											}
										},
										{
											"permission_name": {
												"_eq": "accesscontrol.b.write"
											}
										}
									]
								}
							},
							"check": {
								"ac_hasura_test_template_1_location_permission": {
									"_and": [
										{
											"user_id": {
												"_eq": "X-Hasura-User-Id"
											}
										},
										{
											"permission_name": {
												"_eq": "accesscontrol.b.write"
											}
										}
									]
								}
							}
						}
					]
				}
			},
			"postgres": null,
			"accessPathTable": null,
			"locationCol": "location_id",
			"permissionPrefix": "accesscontrol.b",
			"permissions": {
				"postgres": null,
				"hasura": []
			},
			"ownerCol": null,
			"use_custom_policy": null,
			"hasura_policies": null,
			"postgres_policies": null
		}
	]
}`

func TestPostgresAC(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.Nil(VerifyPostgresAC())
}

func TestHasuraAC(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.Nil(VerifyHasuraAC())
}

func removeRestrictivePolicy(table tableSchema) []*tablePolicy {
	foundIndex := -1
	for i, policy := range table.Policies {
		if policy.Permissive.String == RESTRICTIVE {
			foundIndex = i
		}
	}
	if foundIndex == -1 {
		return table.Policies
	}
	return append(table.Policies[:foundIndex], table.Policies[foundIndex+1:]...)
}

func removeAllPermissivePolicy(table tableSchema) []*tablePolicy {
	policies := []*tablePolicy{}
	for _, policy := range table.Policies {
		if policy.Permissive.String != PERMISSIVE {
			policies = append(policies, policy)
		}
	}

	return policies
}

func removeOnePolicy[T any](slice []T) []T {
	return append(slice[:1], slice[1+1:]...)
}

func cloneData[T any](data *T) *T {
	src := []T{*data}
	des := []T{}
	copy(des, src)
	return &des[0]
}

func createWrongCondition() interface{} {
	return map[string]map[string]string{
		"_and": {
			"test": "1",
		},
	}
}

func TestVerifyPostgresAC(t *testing.T) {
	svc := "mastermgmt"
	var tableStage = &rls.FileStage{}
	err := json.Unmarshal([]byte(tableStageJSONStr), &tableStage)
	assert.NoError(t, err)

	var acTestTemplate114Schema = &tableSchema{}

	t.Run("Should return success when compare stage and database tracing", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)
		assert.NoError(t, err)
	})

	t.Run("Should return error when missing restrictive data", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies = removeRestrictivePolicy(*acTestTemplate114Schema)

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)
		assert.Error(t, err)
		assert.Equal(t, "table ac_test_template_11_4 in service mastermgmt missing restrictive rls policy", err.Error())
	})

	t.Run("Should return error when don't have any permissive policy", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies = removeAllPermissivePolicy(*acTestTemplate114Schema)

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "table ac_test_template_11_4 in service mastermgmt missing permissive rls policy", err.Error())
	})

	t.Run("Should return error when policy name is wrong", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies[0].PolicyName = pgtype.Text{String: "wrong_policy_name"}

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "policy wrong_policy_name on table ac_test_template_11_4 in service mastermgmt is unexpected policy", err.Error())
	})

	t.Run("Should return error when force row security is not enable", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies[0].Relforcerowsecurity = pgtype.Bool{Bool: false}

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "please force row level security for table ac_test_template_11_4 in service mastermgmt", err.Error())
	})

	t.Run("Should return error when row security is not enable", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies[0].RelrowSecurity = pgtype.Bool{Bool: false}

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "row security is not enable for table ac_test_template_11_4 in service mastermgmt", err.Error())
	})

	t.Run("Should return error when policy is not public type", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies[0].Roles.Set([]string{"bob"})

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "policy for table ac_test_template_11_4 in service mastermgmt is not granted to public", err.Error())
	})

	t.Run("Should return error when content USING of policy is not correct", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		// removed check permission name
		acTestTemplate114Schema.Policies[0].Qual = pgtype.Text{String: "(true \u003c= ( SELECT true AS bool\n   FROM (granted_permissions p\n     JOIN ac_test_template_11_4_access_paths usp ON ((usp.location_id = p.location_id)))\n  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (usp.ac_test_template_11_4_id = ac_test_template_11_4.ac_test_template_11_4_id))\n LIMIT 1))"}

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "policy rls_ac_test_template_11_4_delete_location on table ac_test_template_11_4 in service mastermgmt have content using is not correct", err.Error())
	})

	t.Run("Should return error when content WITH CHECK of policy is not correct", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		// removed check permission name
		acTestTemplate114Schema.Policies[1].WithCheck = pgtype.Text{String: "select * from a"}

		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "policy rls_ac_test_template_11_4_insert_location on table ac_test_template_11_4 in service mastermgmt have content with check is not correct", err.Error())
	})

	t.Run("Should return error when policy is missing one", func(t *testing.T) {
		err = json.Unmarshal([]byte(acTestTemplate114SchemaJSONStr), &acTestTemplate114Schema)
		assert.NoError(t, err)

		acTestTemplate114Schema.Policies = removeOnePolicy(acTestTemplate114Schema.Policies)
		err = VerifyPostgresRls(svc, acTestTemplate114Schema, *tableStage)

		assert.Error(t, err)
		assert.Equal(t, "table ac_test_template_11_4 in service mastermgmt missing rls_ac_test_template_11_4_insert_location policy", err.Error())
	})
}

func TestVerifyHasuraAC(t *testing.T) {
	var tableStage = &rls.FileStage{}
	err := json.Unmarshal([]byte(acTestHasuraTemplate1JSONstr), &tableStage)
	assert.NoError(t, err)

	tableMetadata := []rls.HasuraTable{}

	t.Run("Should return success when compare stage and hasura metadata", func(t *testing.T) {
		err = yaml.Unmarshal([]byte(metadataYAMLStr), &tableMetadata)
		assert.NoError(t, err)

		err = verifyTableHasuraAC(*tableStage, tableMetadata[0])
		assert.NoError(t, err)
	})

	t.Run("Should return error when missing SelectPermission", func(t *testing.T) {
		err = yaml.Unmarshal([]byte(metadataYAMLStr), &tableMetadata)
		assert.NoError(t, err)

		tableMetadata[0].SelectPermissions = nil

		err = verifyTableHasuraAC(*tableStage, tableMetadata[0])
		assert.Error(t, err)
		assert.Equal(t, "role (select) on table ac_hasura_test_template_1 missing select permission", err.Error())
	})

	t.Run("Should return error when missing on role permission", func(t *testing.T) {
		err = yaml.Unmarshal([]byte(metadataYAMLStr), &tableMetadata)
		assert.NoError(t, err)

		table := tableMetadata[0]
		selectPer := removeOnePolicy(*table.SelectPermissions)
		table.SelectPermissions = &selectPer
		err = verifyTableHasuraAC(*tableStage, table)

		assert.Error(t, err)
		assert.Equal(t, "role (delete) on table ac_hasura_test_template_1 missing permission: MANABIE", err.Error())
	})

	t.Run("Should return error when wrong filter", func(t *testing.T) {
		err = yaml.Unmarshal([]byte(metadataYAMLStr), &tableMetadata)
		assert.NoError(t, err)

		table := tableMetadata[0]
		selectPer := *table.SelectPermissions
		filter := createWrongCondition()
		selectPer[0].Permission.Filter = &filter
		err = verifyTableHasuraAC(*tableStage, table)

		assert.Error(t, err)
		assert.Equal(t, "role (select) USER_GROUP_ADMIN on table ac_hasura_test_template_1 have content of filter is not correct", err.Error())
	})

	t.Run("Should return error when wrong filter", func(t *testing.T) {
		err = yaml.Unmarshal([]byte(metadataYAMLStr), &tableMetadata)
		assert.NoError(t, err)

		table := tableMetadata[0]
		insertPer := *table.InsertPermissions
		filter := createWrongCondition()
		insertPer[0].Permission.Check = &filter
		err = verifyTableHasuraAC(*tableStage, table)

		assert.Error(t, err)
		assert.Equal(t, "role (insert) USER_GROUP_ADMIN on table ac_hasura_test_template_1 have content of check is not correct", err.Error())
	})

}
