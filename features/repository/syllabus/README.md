# Repository test

## Why
Repository test to reduce the test cases on service layer. Cover edge cases, reduce bug potential. More detail: https://manabie.atlassian.net/wiki/spaces/TECH/pages/444859964/Refactor+testing
## How
### Overview:
### Structure: a entity will have according to a folder test
- entity (folder): including inccure entities when write the test, especially  graphql.
- utils: (folder):
    - common.go: common steps, if alot we can separate to multi files.
    - graphql.go: focus on graphql helper.
    - setup.go: setup suite. Use go's generic technique.
    - utils.go
- repository_name (folder)
- bdd.go: setup config
- learning_material_steps.go: only steps belong to LM's test cases.
- study_plan_steps.go: only steps belong to SP's test cases.
### Usage:
- Add 1 folder with `entity_name` which you want to test if not exist. 
- A test entity should have these parts (if have):
    - Complicated queries 
    - Hasura
    - View 
    - Trigger
- Setup `repository_name` test:
    - `syllabus/repository_name/steps.go`: repository test's steps
    ```go
    type StepState struct {
        BookID          string
        ...
    }

    func InitStep(s *Suite) map[string]interface{} {
        steps := map[string]interface{}{
            `^step_description`:                   s.func_implementation,
        }
        return steps
    }

    ```
    - `syllabus/repository_name/bdd.go`: setup suite config, if it's default, no need declare. For example:
        - Default: 
        ```go
        type Suite utils.Suite[StepState]
        ```
        - Customize:
        ```go
        type Suite struct {
            *StepState
            DB             database.Ext
            ZapLogger      *zap.Logger
            HasuraAdminURL string
            HasuraPassword string
        }

        func NewSuite(stepState *StepState, db database.Ext, zapLogger *zap.Logger, hasuraURL, hasuraPwd string) *Suite {
            return &Suite{
                StepState:      stepState,
                DB:             db,
                ZapLogger:      zapLogger,
                HasuraAdminURL: hasuraURL,
                HasuraPassword: hasuraPwd,
            }
        }
        ```
    - `syllabus/bdd.go`: declare repository Step, this cases `book` repo.
        ```go
        type StepState struct {
            DefaultSchoolID int32
            BookStepState   *book.StepState
        }
      ```
    - `syllabus/bdd.go`: Add recognize path if not exist + its StepState (if not -> panic), in this case: `book`.
        ```go
        func ScenarioInitializer(c *gandalf.Config) func(ctx *godog.ScenarioContext) {
	        return func(ctx *godog.ScenarioContext) {
                s := newSuite()
                initSteps(ctx, s)
                s.DefaultSchoolID = DefaultSchoolID

                ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
                    uriSplit := strings.Split(sc.Uri, ":")
                    entityName := strings.Split(uriSplit[0], "/")[2]
                    switch entityName {
                    case "book":
                        ctx = utils.StepStateToContext(ctx, s.BookStepState)
                    default:
                        ctx = utils.StepStateToContext(ctx, s.StepState)
                    }
                    if ENABLED_DEFAULT_RLS {
                        ctx = addResourcePathToCtx(ctx, s.DefaultSchoolID)
                    }
                    return ctx, nil
                })
            }
        }
        ```
    - Depend on `repository` belong to LM or SP, we add steps according to LM, SP, in this case `book` belong to LM, so in `learning_material_steps.go`:
        ```go
        func initLearningMaterialStep(s *Suite) map[string]interface{} {
            steps := map[string]interface{}{}

            // init entities's steps.
            bookStep := book.InitStep((*book.Suite)(utils.NewEntitySuite(s.StepState.BookStepState, s.DB, s.ZapLogger, s.HasuraAdminURL, s.HasuraPassword)))

            utils.AppendSteps(steps, bookStep)
            return steps
        }
        ```
    - The func call hasura should prepare the config first: if the table is new, should add track table, and the query not yet exists in file `query_collection.yaml` (eureka): https://github.com/manabie-com/backend/blob/78761a96833520cd15e191e423671a289903cb08/deployments/helm/manabie-all-in-one/charts/eureka/files/hasura/metadata/query_collections.yaml
        ```go
        if err := utils.TrackTableForHasuraQuery(
            s.HasuraAdminURL,
            s.HasuraPassword,
            "books",
        ); err != nil {
            return nil, fmt.Errorf("trackTableForHasuraQuery: %w", err)
        }

        if err := utils.CreateSelectPermissionForHasuraQuery(
            s.HasuraAdminURL,
            constant.UserGroupAdmin,
            "books",
        ); err != nil {
            return nil, fmt.Errorf("createSelectPermissionForHasuraQuery: %w", err)
        }

        rawQuery := `query BooksTitle($book_id: String!) {
            books(where: {book_id: {_eq: $book_id}}) {
            name
            }
        }`

        if err := utils.AddQueryToAllowListForHasuraQuery(s.HasuraAdminURL, s.HasuraPassword, rawQuery); err != nil {
            return StepStateToContext(ctx, stepState), fmt.Errorf("addQueryToAllowListForHasuraQuery(): %v", err)
        }
        ```

- How to run: 
    ```
    ./deployments/k8s_run_repository_test.bash features/repository/syllabus/
    ```