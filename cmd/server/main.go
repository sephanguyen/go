package main

import (
	"context"
	"fmt"
	"log"
	"os"

	_ "github.com/manabie-com/backend/cmd/server/auth"
	_ "github.com/manabie-com/backend/cmd/server/bob"
	_ "github.com/manabie-com/backend/cmd/server/calendar"
	_ "github.com/manabie-com/backend/cmd/server/conversationmgmt"
	_ "github.com/manabie-com/backend/cmd/server/discount"
	_ "github.com/manabie-com/backend/cmd/server/draft"
	_ "github.com/manabie-com/backend/cmd/server/enigma"
	_ "github.com/manabie-com/backend/cmd/server/entryexitmgmt"
	_ "github.com/manabie-com/backend/cmd/server/eureka"
	"github.com/manabie-com/backend/cmd/server/fatima"
	"github.com/manabie-com/backend/cmd/server/fink"
	_ "github.com/manabie-com/backend/cmd/server/hephaestus"
	_ "github.com/manabie-com/backend/cmd/server/invoicemgmt"
	_ "github.com/manabie-com/backend/cmd/server/jerry"
	_ "github.com/manabie-com/backend/cmd/server/jerry2"
	_ "github.com/manabie-com/backend/cmd/server/lessonmgmt"
	"github.com/manabie-com/backend/cmd/server/lessonmgmt/job"
	_ "github.com/manabie-com/backend/cmd/server/mastermgmt"
	_ "github.com/manabie-com/backend/cmd/server/notificationmgmt"
	_ "github.com/manabie-com/backend/cmd/server/payment"
	_ "github.com/manabie-com/backend/cmd/server/rls_scan"
	_ "github.com/manabie-com/backend/cmd/server/shamir"
	_ "github.com/manabie-com/backend/cmd/server/spike"
	_ "github.com/manabie-com/backend/cmd/server/timesheet"
	_ "github.com/manabie-com/backend/cmd/server/tom"
	"github.com/manabie-com/backend/cmd/server/usermgmt"
	_ "github.com/manabie-com/backend/cmd/server/virtualclassroom"
	_ "github.com/manabie-com/backend/cmd/server/yasuo"
	_ "github.com/manabie-com/backend/cmd/server/zeus"
	bobCfg "github.com/manabie-com/backend/internal/bob/configurations"
	fatimaCfg "github.com/manabie-com/backend/internal/fatima/configurations"
	finkCfg "github.com/manabie-com/backend/internal/fink/configurations"
	_ "github.com/manabie-com/backend/internal/golibs/automaxprocs"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	lessonmgmtCfg "github.com/manabie-com/backend/internal/lessonmgmt/configurations"
	tomCfg "github.com/manabie-com/backend/internal/tom/configurations"
	tomstress "github.com/manabie-com/backend/internal/tom/infra/stress"
	usermgmtCfg "github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"cloud.google.com/go/profiler"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

var rootCmd = &cobra.Command{}

func makeRootCmd() {
	var (
		configPath       string
		commonConfigPath string
		secretsPath      string

		// only batch job commands use those
		schoolID, schoolName string

		// only used for migrate student enrollment status
		newStatus, originStatus, resourcePath string

		// only used for migrate resource_path in eureka db
		bobConfigPath       string
		bobCommonConfigPath string
		bobSecretPath       string

		createdAt, organizationID string

		// only used for create user_group
		userGroupName, roles, locationIDs string

		// only used for assign usergroup to specific staff
		orgID, userGroupID, userIDsSequence string

		// only used for migrate current_grade to grade_id
		gradePartnerIDs string

		// only used for migrate user phone number / locations
		userType string

		enableCreatePublicationInLocal bool

		// only for migrate kec data
		bucketName, objectName string
	)

	tomStressConf := tomstress.StagingStressConfig{}
	cmdTomStressTest := &cobra.Command{
		Use:   "tom_stress_test",
		Short: "Ping stress test for tom",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &tomCfg.Config{
				Common: configs.CommonConfig{
					Name: "tom_stress_test",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg,
			)

			tomstress.RunStagingStressTest(cmd.Context(), cfg, &tomStressConf)
			return nil
		},
	}
	tomstress.BindCobra(cmdTomStressTest.Flags(), &tomStressConf)

	cmdFinkUpsertStreams := &cobra.Command{
		Use:   "upsert_streams",
		Short: "This is a pre-hook of manabie-all-in-one to create streams of nats-jetstream",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &finkCfg.Config{
				Common: configs.CommonConfig{
					Name: "upsert_streams",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg,
			)

			fink.RunUpsertStreams(cmd.Context(), cfg)
			return nil
		},
	}

	cmdUserMgmtMigrateUsersFromFirebase := &cobra.Command{
		Use:   "usermgmt_migrate_users_from_firebase",
		Short: "Migrate users from firebase",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg,
			)
			startProfiler(cmd, &cfg.Common)
			return usermgmt.MigrateUsersFromFirebase(cmd.Context(), cfg, organizationID)
		},
	}

	cmdUserMgmtMigrateUsersFromFirebase.Flags().StringVar(&organizationID, "organizationID", "", "Increase grade of students which were matched with organization_id")

	cmdFatimaMigrateStudentPackagesToStudentPackageAccessPath := &cobra.Command{
		Use:   "fatima_migrate_student_packages_to_student_package_access_path",
		Short: "Migrate student_package to student_package_access_path",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// this is used for querying orgs
			bobCfg := &bobCfg.Config{
				Common: configs.CommonConfig{
					Name: "fatima_migrate_student_subscriptions",
				},
			}
			fatimaCfg := &fatimaCfg.Config{
				Common: configs.CommonConfig{
					Name: "fatima_migrate_student_subscriptions",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				fatimaCfg,
			)
			configs.MustLoadConfig(cmd.Context(),
				bobCommonConfigPath,
				bobConfigPath,
				bobSecretPath,
				bobCfg,
			)
			startProfiler(cmd, &bobCfg.Common)
			fatima.RunMigrateStudentPackagesToStudentPackageAccessPath(cmd.Context(), fatimaCfg, bobCfg)
			return nil
		},
	}

	cmdUsermgmtMigrationDeleteStudentLocationOrg := &cobra.Command{
		Use:   "usermgmt_migrate_delete_student_location_org",
		Short: "Migrate delete student_location org",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrateDeleteStudentLocationOrg(cmd.Context(), config)
			return nil
		},
	}

	cmdUsermgmtIncreaseGradeOfStudents := &cobra.Command{
		Use:   "usermgmt_increase_grade_of_students",
		Short: "Increase grade of students",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunIncreaseGradeOfStudents(cmd.Context(), config, createdAt, organizationID)
			return nil
		},
	}
	cmdUsermgmtIncreaseGradeOfStudents.Flags().StringVar(&createdAt, "createdAt", "", "Increase grade of students which were created before created_at")
	cmdUsermgmtIncreaseGradeOfStudents.Flags().StringVar(&organizationID, "organizationID", "", "Increase grade of students which were matched with organization_id")

	cmdUsermgmtMigrateCurrentGradeToGradeID := &cobra.Command{
		Use:   "usermgmt_migrate_current_grade_to_grade_id",
		Short: "Increase grade of students",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrateCurrentGradeToGradeID(cmd.Context(), config, gradePartnerIDs, organizationID)
			return nil
		},
	}
	cmdUsermgmtMigrateCurrentGradeToGradeID.Flags().StringVar(&gradePartnerIDs, "gradePartnerIDs", "", "Migrate current_grade to grade_id matched with gradePartnerIDs")
	cmdUsermgmtMigrateCurrentGradeToGradeID.Flags().StringVar(&organizationID, "organizationID", "", "Migrate current_grade to grade_id matched with organization_id")

	cmdUsermgmtMigrationStudentEnrollmentOriginalStatus := &cobra.Command{
		Use:   "usermgmt_migrate_student_enrollment_original_status",
		Short: "Migrate student_enrollment original status",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrateStudentEnrollmentOriginalStatus(cmd.Context(), config, newStatus, originStatus, resourcePath)
			return nil
		},
	}

	cmdUsermgmtMigrationStudentEnrollmentOriginalStatus.Flags().StringVar(&newStatus, "newStatus", "", "migrate for student enrollment")
	cmdUsermgmtMigrationStudentEnrollmentOriginalStatus.Flags().StringVar(&originStatus, "originStatus", "", "migrate for student enrollment")
	cmdUsermgmtMigrationStudentEnrollmentOriginalStatus.Flags().StringVar(&resourcePath, "resourcePath", "", "migrate for student enrollment")

	cmdUsermgmtRunMigrationAddDefaultUserGroupForStudentParent := &cobra.Command{
		Use:   "usermgmt_migrate_add_default_usergroup_for_student_parent",
		Short: "Migration Add Default UserGroup For Student & Parent",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrationAddDefaultUserGroupForStudentParent(cmd.Context(), config)
			return nil
		},
	}

	cmdUsermgmtMigrationCreateUserGroup := &cobra.Command{
		Use:   "usermgmt_migrate_create_user_group",
		Short: "Migrate create user group",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrateCreateUserGroup(cmd.Context(), config, userGroupName, roles, locationIDs, organizationID)
			return nil
		},
	}

	cmdUsermgmtMigrationCreateUserGroup.Flags().StringVar(&userGroupName, "userGroupName", "", "migrate create user group")
	cmdUsermgmtMigrationCreateUserGroup.Flags().StringVar(&roles, "roles", "", "migrate create user group")
	cmdUsermgmtMigrationCreateUserGroup.Flags().StringVar(&locationIDs, "locationIDs", "", "migrate create user group")
	cmdUsermgmtMigrationCreateUserGroup.Flags().StringVar(&organizationID, "organizationID", "", "migrate create user group")

	cmdMigrateStudentSubscriptions := &cobra.Command{
		Use:   "fatima_migrate_student_subscriptions",
		Short: "Migrate data from fatima.student_packages to bob.lesson_student_subscriptions.",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// this is used for querying orgs
			bobCfg := &bobCfg.Config{
				Common: configs.CommonConfig{
					Name: "fatima_migrate_student_subscriptions",
				},
			}
			fatimaCfg := &fatimaCfg.Config{
				Common: configs.CommonConfig{
					Name: "fatima_migrate_student_subscriptions",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				fatimaCfg,
			)
			configs.MustLoadConfig(cmd.Context(),
				bobCommonConfigPath,
				bobConfigPath,
				bobSecretPath,
				bobCfg,
			)
			startProfiler(cmd, &bobCfg.Common)
			fatima.RunMigrateStudentSubscriptions(cmd.Context(), bobCfg, fatimaCfg)
			return nil
		},
	}
	cmdMigrateStudentSubscriptions.Flags().StringVar(&schoolID, "schoolID", "", "migrate for specific school")
	cmdMigrateStudentSubscriptions.Flags().StringVar(&schoolName, "schoolName", "", "migrate for specific school")

	cmdUsermgmtMigrationSetCurrentSchoolByGrade := &cobra.Command{
		Use:   "usermgmt_migrate_set_current_school_by_grade",
		Short: "Migrate set current school by grade",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrateSetCurrentSchoolByGrade(cmd.Context(), config)
			return nil
		},
	}

	cmdMigrateStaffLocations := &cobra.Command{
		Use:   "usermgmt_migrate_locations_for_users",
		Short: "Migrate Locations For Users",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrateLocationsForUsers(cmd.Context(), config, organizationID, locationIDs, userType)
			return nil
		},
	}

	cmdMigrateStaffLocations.Flags().StringVar(&organizationID, "organizationID", "", "Migrate Locations For Users With organizationID")
	cmdMigrateStaffLocations.Flags().StringVar(&locationIDs, "locationIDs", "", "Migrate Locations For Users With locationIDs")
	cmdMigrateStaffLocations.Flags().StringVar(&userType, "userType", "", "Migrate Locations For Users With userType")

	cmdUpdateDataLessonReport := &cobra.Command{
		Use:   "sync_lesson_report_data",
		Short: "job update lesson data report",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &lessonmgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "sync_lesson_report_data",
				},
			}
			configs.MustLoadConfig(
				cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg)
			startProfiler(cmd, &cfg.Common)

			job := job.InitJob(cmd.Context(), cfg, &job.ConfigJob{TotalJobConcurrency: 1, Limit: 1000}, job.LESSON_REPORT_EXECUTOR)
			return job.Run(cmd.Context())
		},
	}

	cmdUpdateDataLessonMembers := &cobra.Command{
		Use:   "sync_lesson_members_data",
		Short: "job update lesson members data",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &lessonmgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "sync_lesson_members_data",
				},
			}
			configs.MustLoadConfig(
				cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg)
			startProfiler(cmd, &cfg.Common)

			job := job.InitJob(cmd.Context(), cfg, &job.ConfigJob{TotalJobConcurrency: 5, Limit: 200}, job.LESSON_MEMBERS_EXECUTOR)
			return job.Run(cmd.Context())
		},
	}

	cmdUpdateDataLessonTeachers := &cobra.Command{
		Use:   "sync_lesson_teachers_data",
		Short: "job update lesson teacher data",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &lessonmgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "sync_lesson_teachers_data",
				},
			}
			configs.MustLoadConfig(
				cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg)
			startProfiler(cmd, &cfg.Common)

			job := job.InitJob(cmd.Context(), cfg, &job.ConfigJob{TotalJobConcurrency: 5, Limit: 1000}, job.LESSON_TEACHERS_EXECUTOR)
			return job.Run(cmd.Context())
		},
	}

	cmdUpdateDataStudentSubscriptions := &cobra.Command{
		Use:   "sync_lesson_student_subscriptions_data",
		Short: "job update lesson teacher data",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &lessonmgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "sync_lesson_student_subscriptions_data",
				},
			}
			configs.MustLoadConfig(
				cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfg)
			startProfiler(cmd, &cfg.Common)

			job := job.InitJob(cmd.Context(), cfg, &job.ConfigJob{TotalJobConcurrency: 5, Limit: 1000}, job.LESSON_STUDENT_SUBSCRIPTIONS_EXECUTOR)
			return job.Run(cmd.Context())
		},
	}

	cmdUsermgmtMigrateStudentFullNameToLastNameAndFirstName := &cobra.Command{
		Use:   "usermgmt_migrate_student_full_name_to_last_name_and_first_name",
		Short: "Migrate Full Name to Last Name",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfbBob := &bobCfg.Config{}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfbBob,
			)
			startProfiler(cmd, &cfbBob.Common)
			usermgmt.RunMigrateStudentFullNameToLastNameAndFirstName(cmd.Context(), cfbBob)
			return nil
		},
	}

	cmdUsermgmtMigrateAssignUsergroupToSpecificStaff := &cobra.Command{
		Use:   "usermgmt_migrate_assign_user_group_to_specific_staff",
		Short: "migrate assign user_group to specific staff",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &usermgmtCfg.Config{
				Common: configs.CommonConfig{
					Name: "usermgmt",
				},
			}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				config,
			)
			startProfiler(cmd, &config.Common)
			usermgmt.RunMigrationAssignUsergroupToSpecificStaff(cmd.Context(), config, orgID, userGroupID, userIDsSequence)
			return nil
		},
	}

	cmdUsermgmtMigrateAssignUsergroupToSpecificStaff.Flags().StringVar(&orgID, "organizationID", "", "migrate for assign usergroup to specific staff")
	cmdUsermgmtMigrateAssignUsergroupToSpecificStaff.Flags().StringVar(&userGroupID, "userGroupID", "", "migrate for assign usergroup to specific staff")
	cmdUsermgmtMigrateAssignUsergroupToSpecificStaff.Flags().StringVar(&userIDsSequence, "userIDsSequence", "", "migrate for assign usergroup to specific staff")

	cmdUsermgmtMigrateUserPhoneNumber := &cobra.Command{
		Use:   "usermgmt_migrate_user_phone_number",
		Short: "migrate user phone number",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfbBob := &bobCfg.Config{}
			configs.MustLoadConfig(cmd.Context(),
				commonConfigPath,
				configPath,
				secretsPath,
				cfbBob,
			)
			startProfiler(cmd, &cfbBob.Common)
			usermgmt.RunMigrateUserPhoneNumber(cmd.Context(), cfbBob, orgID, userType)
			return nil
		},
	}

	cmdUsermgmtMigrateUserPhoneNumber.Flags().StringVar(&orgID, "organizationID", "", "migrate for assign user phone number to specific staff or student")
	cmdUsermgmtMigrateUserPhoneNumber.Flags().StringVar(&userType, "userType", "", "migrate for assign user phone number to specific staff or student")
	var err error
	rootCmd.Use = "server"

	rootCmd.PersistentFlags().StringVar(
		&commonConfigPath,
		"commonConfigPath",
		"",
		"path to common configuration file, usually used for configuration",
	)
	err = multierr.Append(err, rootCmd.MarkPersistentFlagRequired("commonConfigPath"))

	rootCmd.PersistentFlags().StringVar(
		&configPath,
		"configPath",
		"",
		"path to configuration file, usually used for configuration",
	)
	err = multierr.Append(err, rootCmd.MarkPersistentFlagRequired("configPath"))

	rootCmd.PersistentFlags().StringVar(
		&secretsPath,
		"secretsPath",
		"",
		"path to encrypted secrets file, usually used for secrets data",
	)
	err = multierr.Append(err, rootCmd.MarkPersistentFlagRequired("secretsPath"))

	rootCmd.PersistentFlags().StringVar(
		&bobConfigPath,
		"bobConfigPath",
		"",
		"path to bob configuration file, usually used for bob configuration",
	)
	rootCmd.PersistentFlags().StringVar(
		&bobSecretPath,
		"bobSecretPath",
		"",
		"path to bob encrypted secrets file, usually used for bob secrets data",
	)
	rootCmd.PersistentFlags().StringVar(
		&bobCommonConfigPath,
		"bobCommonConfigPath",
		"",
		"path to bob common configuration file, usually used for bob configuration",
	)

	rootCmd.PersistentFlags().BoolVar(
		&enableCreatePublicationInLocal,
		"enableCreatePublicationInLocal",
		false,
		"Only use this local environment, in other env we have to manage this upfront",
	)

	rootCmd.PersistentFlags().StringVar(
		&bucketName,
		"bucketName",
		"",
		"Bucket name migrate enrollment status",
	)

	rootCmd.PersistentFlags().StringVar(
		&objectName,
		"objectName",
		"",
		"object name migrate enrollment status",
	)

	rootCmd.PersistentFlags().StringVar(
		&organizationID,
		"organizationID",
		"",
		"organization id",
	)

	rootCmd.AddCommand(
		cmdUserMgmtMigrateUsersFromFirebase,
		cmdUsermgmtIncreaseGradeOfStudents,
		cmdUsermgmtMigrationDeleteStudentLocationOrg,
		cmdUsermgmtMigrationStudentEnrollmentOriginalStatus,
		cmdUsermgmtMigrateAssignUsergroupToSpecificStaff,
		cmdUsermgmtMigrateUserPhoneNumber,
		cmdUsermgmtRunMigrationAddDefaultUserGroupForStudentParent,
		cmdUsermgmtMigrationCreateUserGroup,
		cmdMigrateStaffLocations,
		cmdMigrateStudentSubscriptions,
		cmdFatimaMigrateStudentPackagesToStudentPackageAccessPath,
		cmdTomStressTest,
		cmdUsermgmtMigrateStudentFullNameToLastNameAndFirstName,
		cmdFinkUpsertStreams,
		cmdUsermgmtMigrateCurrentGradeToGradeID,
		cmdUpdateDataLessonReport,
		cmdUpdateDataLessonMembers,
		cmdUsermgmtMigrationSetCurrentSchoolByGrade,
		cmdUpdateDataLessonTeachers,
		cmdUpdateDataStudentSubscriptions,
	)
	bootstrap.AddCommand(rootCmd)

	if err != nil {
		log.Fatalf("failed to set up root cobra.Command: %s", err)
	}
}

func main() {
	makeRootCmd()

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}

func startProfiler(_ *cobra.Command, cfg *configs.CommonConfig) {
	// Profiler initialization, best done as early as possible.
	err := profiler.Start(profiler.Config{
		ProjectID:      cfg.GoogleCloudProject,
		Service:        fmt.Sprintf("%s-%s-%s", cfg.Environment, cfg.Organization, cfg.Name),
		ServiceVersion: cfg.ImageTag,
	})
	if err != nil {
		log.Println("startProfiler error", err)
	}
}
