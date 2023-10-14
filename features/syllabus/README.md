# Syllabus service test

## Why

Ensure your logic interactive with another service correctly, close with uat/prod env. More detail: [Refactor testing](https://manabie.atlassian.net/wiki/spaces/TECH/pages/444859964/Refactor+testing)

## How

### Overview:

### Structure: a service will have according to a folder test

- entity (folder): define commonly used entities.
- utils: (folder):
  - setup.go: setup suite. Use go's generic technique.
  - Other files: Contains helpers for commonly used steps such as auth, validate status code, ...
- service_name (folder)/feature_name (**if big**): flexible.
- bdd.go: setup config
- learning_material_steps.go: only steps belong to LM's test cases.
- study_plan_steps.go: only steps belong to SP's test cases.

### Usage:

- Add 1 folder with `service_name/feature_name` which you want to test if not exist. (Recommend using `make auto-gen-bdd` for quick setup)
- NOTE: `make auto-gen-bdd` only for gen simple test cases, you have to add additional test cases to ensure your feature.
- Setup `service_name` test:

  - `syllabus/service_name/steps.go`: service test's steps

    ```go
    type StepState struct {
        Response                     interface{}
        Request                      interface{}
        ResponseErr                  error
        BookID                       string
        TopicIDs                     []string
        ChapterIDs                   []string
        Token                        string
        SchoolAdmin                  entity.SchoolAdmin
        Student                      entity.Student

        AssignmentID          string
        ...
    }

    func InitStep(s *Suite) map[string]interface{} {
        steps := map[string]interface{}{
            `^step_description`:                   s.func_implementation,
        }
        return steps
    }

    ```

  - `syllabus/service_name/bdd.go`: setup suite config, if it's default, no need declare. For example:

    - Default:

    ```go
    type Suite utils.Suite[StepState]
    ```

    - Customize:

    ```go
    type Suite struct {
        *StepState
        *common.Connections
        ZapLogger  *zap.Logger
        Cfg        *common.Config
        AuthHelper *AuthHelper
    }

    func NewSuite(stepState *StepState, connections *common.Connections, zapLogger *zap.Logger, cfg *common.Config, authHelper *AuthHelper) *Suite {
        return &Suite{
            StepState:      stepState,
            Connections: connections,
            ZapLogger:   zapLogger,
            Cfg:         cfg,
            AuthHelper:  authHelper,
        }
    }
    ```

  - `syllabus/bdd.go`: declare service Step, this cases `assignment` service.
    ```go
    type StepState struct {
        AssignmentStepState   *assignment.StepState
    }
    ```
  - `syllabus/bdd.go`: Add recognize path if not exist + its StepState (if not -> panic), in this case: `assignment`.

    ```go
    func ScenarioInitializer(c *common.Config) func(ctx *godog.ScenarioContext) {
    return func(ctx *godog.ScenarioContext) {
    	s := newSuite(c)
    	initSteps(ctx, s)

    	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    		ctx = InitSyllabusState(ctx, s)
    		uriSplit := strings.Split(sc.Uri, ":")
    		entityName := strings.Split(uriSplit[0], "/")[1]
    		switch entityName {
    		case "assignment":
                    ctx = utils.StepStateToContext(ctx, s.AssignmentStepState)
            default:
                    ctx = utils.StepStateToContext(ctx, s.StepState)
    		}

    		claim := interceptors.CustomClaims{
    			Manabie: &interceptors.ManabieClaims{
    				ResourcePath: strconv.Itoa(ManabieSchool),
    				DefaultRole:  constants.RoleSchoolAdmin,
    				UserGroup:    entities.UserGroupSchoolAdmin,
    			},
    		}
    		ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

    		return ctx, nil
    	})
    }
    ```

  - Depend on `service` belong to LM or SP, we add steps according to LM, SP, in this case `assignment` belong to LM, so in `learning_material_steps.go`:

    ```go
    func initLearningMaterialStep(s *Suite) map[string]interface{} {
        steps := map[string]interface{}{}

        // init entities's steps.
        assignmentStep := assignment.InitStep((*assignment.Suite)(utils.NewEntitySuite(s.StepState.AssignmentStepState, s.Connections, s.ZapLogger, s.Cfg, s.AuthHelper)))

        utils.AppendSteps(steps, assignmentStep)
        return steps
    }
    ```

- <mark>Note when implementing step:</mark> For commonly used steps (like `a signed in`, `a valid book content`, `return status code`, ...) you must add the prefix `service_name` to the step to distinguish the steps that are similar in different service, avoid overwrite.

  ```gherkin
  Scenario Outline: examples
    Given <assignment>a signed in "<role>"
    When do something
    Then <assignment>returns "<msg>" status code
  ```

- How to run:
  ```bash
  ./deployments/k8s_bdd_test.bash features/syllabus/
  ```
