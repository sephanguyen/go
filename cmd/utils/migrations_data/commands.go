package migrationsdata

import (
	"context"
	"fmt"

	fileio "github.com/manabie-com/backend/internal/golibs/io"

	"github.com/spf13/cobra"
)

var (
	table           string
	batchSize       int
	migFileInput    string
	migFolderOutput string
	orgID           string
)

const (
	LessonTable     = "lessons"
	LessonsTeachers = "lessons_teachers"
)

func GetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pg_gen_migration",
		Short: "Gen data migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if orgID == "" {
				return fmt.Errorf("orgID should not be empty")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			migrationLesson := InitMigrator(MigratorConfig{fileInput: migFileInput,
				folderOutput:     migFolderOutput,
				tableName:        table,
				batchSize:        batchSize,
				totalConcurrency: 3},
				&fileio.FileUtils{})
			switch table {
			case LessonTable:
				migrationLesson.SetConverter(&LessonConverter{})
			case LessonsTeachers:
				migrationLesson.SetConverter(&LessonTeacherConverter{})
			default:
				return fmt.Errorf("table not implement")
			}
			err := migrationLesson.Run(ctx, orgID)
			return err
		},
	}

	cmd.PersistentFlags().StringVar(&table, "table", "", "Table name you want to convert")
	cmd.PersistentFlags().IntVar(&batchSize, "batch", 1000, "number row want to split")

	cmd.PersistentFlags().StringVar(&migFileInput, "migFileInput", "", "Specific file migration input")
	cmd.PersistentFlags().StringVar(&migFolderOutput, "migFolderOutput", "kec_migrations/out", "Specific folder migration output")
	cmd.PersistentFlags().StringVar(&orgID, "orgID", "", "orgID need to import")

	return cmd
}

var RootCmd = GetCmd()
