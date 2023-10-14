package rls

import (
	"fmt"

	fileio "github.com/manabie-com/backend/internal/golibs/io"

	"github.com/spf13/cobra"
)

var (
	rlsType                     string
	table                       string
	pkey                        string
	accessPathTable             string
	accessPathTableKey          string
	databaseName                string
	permissionPrefix            string
	hasuraVersion               string
	templateVersion             string
	ownerCol                    string
	accessPathLocationCol       string
	removeOldRole               string
	writePermissionHasura       string
	addRLSToAllPermissionHasura bool
	otherTemplateFilterName     string
	mapHasuraDirectly           bool
	stgHasura                   bool
	acFolder                    string
)

const (
	PostgresType    = "pg"
	HasuraType      = "hasura"
	ViewType        = "view"
	CachingViewType = "cache-view"
	GenRole         = "gen-role"
	GenRLSType      = "rls"
	GenRLSFile      = "rls-file"
	RollbackRLSFile = "rollback-rls-file"
)

func GetCmd(f *Postgres, h *Hasura, v *GrantedView, file *FileTemplate) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pg_gen_rls",
		Short: "Gen policies and utils of row level security for Access Control",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch rlsType {
			case PostgresType:
				_, err := f.genPostgresRLS()
				return err
			case HasuraType:
				_, err := h.genRLSMetadata()
				return err
			case ViewType, CachingViewType:
				return v.genGrantedView(cmd, args)
			case GenRole:
				return h.genNewRole()
			case GenRLSType:
				_, err := f.genPostgresRLS()
				if err != nil {
					return err
				}
				_, err = h.genRLSMetadata()
				return err
			case GenRLSFile:
				return file.genFromFile()
			case RollbackRLSFile:
				return file.rollbackRLSFiles()
			default:
				return fmt.Errorf("rlsType not found")
			}
		},
	}

	cmd.PersistentFlags().StringVar(&rlsType, "rlsType", "", "Type of command which will be gen. Accepted value: pg, hasura, view, gen-role, rls.")
	cmd.PersistentFlags().StringVar(&table, "table", "", "name of table which we want to generate RLS")
	cmd.PersistentFlags().StringVar(&pkey, "pkey", "", "primary key of table which related with access_path table or column content location")
	cmd.PersistentFlags().StringVar(&accessPathTable, "accessPathTable", "", "table contain access_path")
	cmd.PersistentFlags().StringVar(&databaseName, "databaseName", "", "name of database which we want to write migration file to")
	cmd.PersistentFlags().StringVar(&permissionPrefix, "permissionPrefix", "", "prefix of permission example user.student.red -> prefix is user.student")
	cmd.PersistentFlags().StringVar(&accessPathTableKey, "accessPathTableKey", "", "accessPathTableKey key of access path table column which can be join with primary key on currently table")
	cmd.PersistentFlags().StringVar(&hasuraVersion, "hasuraVersion", "", "hasuraVersion contain 1 and 2. default is 1")
	cmd.PersistentFlags().StringVar(&templateVersion, "templateVersion", "", "templateVersion contain 1 and 4. default is 1")
	cmd.PersistentFlags().StringVar(&removeOldRole, "removeOldRole", "", "removeOldRole contain true and false. default is false. only use for gen-role")
	cmd.PersistentFlags().StringVar(&ownerCol, "ownerCol", "", "required if templateVersion is 4 contain col which is saved owners of record")
	cmd.PersistentFlags().StringVar(&accessPathLocationCol, "accessPathLocationCol", "", "in case location col in access path table is not location_id we can fill this option")
	cmd.PersistentFlags().StringVar(&writePermissionHasura, "writePermissionHasura", "", "Add INSERT/UPDATE/DELETE permission for hasura")
	cmd.PersistentFlags().BoolVar(&addRLSToAllPermissionHasura, "addRLSToAllPermissionHasura", false, "If true add RLS to all Role of hasura")
	cmd.PersistentFlags().StringVar(&otherTemplateFilterName, "otherTemplateFilterName", "", "Only use for hasura in case we use template 1 beside with template 4.")
	cmd.PersistentFlags().BoolVar(&mapHasuraDirectly, "mapHasuraDirectly", false, "Only use for hasura in case we want map directly main table to granted view")
	cmd.PersistentFlags().BoolVar(&stgHasura, "stgHasura", false, "Only use when build file tables_stg.yaml")
	cmd.PersistentFlags().StringVar(&acFolder, "acFolder", "", "Specific folder access control want to generate")

	return cmd
}

var rlsPostgres = &Postgres{
	IOUtils: &fileio.FileUtils{},
}

var rlsHasura = &Hasura{
	IOUtils: &fileio.FileUtils{},
}

var grantedView = &GrantedView{
	IOUtils: &fileio.FileUtils{},
}

var fileTemplate = &FileTemplate{
	IOUtils: &fileio.FileUtils{},
}

var RootCmd = GetCmd(rlsPostgres, rlsHasura, grantedView, fileTemplate)
