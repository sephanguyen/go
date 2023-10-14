### Handler function signature
The [handler functions](bob/bdd_steps.go#L26-L27) should follow this structure:

```go
func (s *suite) handlerName(ctx context.Context, ....) (context.Context, error) {
```
Later on when Go have generic, we can rewrite the function and check this during compile time. For now, a helper function `helper.BuildRegexpMapV2` checking the signature using reflection during runtime.  
The reason for this signature described below.
### Godog context chain
For more detail https://github.com/cucumber/godog/pull/409  
Given an example scenario like this:
```gherkin
Feature: eat godogs
  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining
```
And old way of implementing the test:
```go
var Godogs int

func thereAreGodogs(available int) error {
	Godogs = available
	return nil
}

func iEat(num int) error {
	if Godogs < num {
		return fmt.Errorf("you cannot eat %d godogs, there are %d available", num, Godogs)
	}
	Godogs -= num
	return nil
}

func thereShouldBeRemaining(remaining int) error {
	if Godogs != remaining {
		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, Godogs)
	}
	return nil
}

func InitializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() { Godogs = 0 })
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		Godogs = 0 // clean the state before every scenario

		return ctx, nil
	})

	sc.Step(`^there are (\d+) godogs$`, thereAreGodogs)
	sc.Step(`^I eat (\d+)$`, iEat)
	sc.Step(`^there should be (\d+) remaining$`, thereShouldBeRemaining)
}
```

This work fine for a small example, but `Godogs` is a global variable, nothing can be good when using a global variable when your project is more than a dozen of files.  
How can we make sure no other step violently read and modify the `Godogs` variable? Similar question, how to run scenarios in concurrency and they're all trying to read and modify a same variable?  
Now, compare with new way of handling:
```go
func thereAreGodogs(ctx context.Context, available int) (context.Context, error) {
	state := StateFromContext(ctx)
	state.Godogs = available
	return StateToContext(ctx, state), nil
}

func iEat(ctx context.Context, num int) (context.Context, error) {
	state := StateFromContext(ctx)
	if state.Godogs < num {
		return ctx, fmt.Errorf("you cannot eat %d godogs, there are %d available", num, Godogs)
	}
	state.Godogs -= num
	return StateToContext(ctx, state), nil
}

func thereShouldBeRemaining(ctx context.Context, remaining int) (context.Context, error) {
	state := StateFromContext(ctx)
	if state.Godogs != remaining {
		return ctx, fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, Godogs)
	}
	return ctx, nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		state := State{}
		state.Godogs = 0
		ctx, _ = context.WithTimeOut(StateToContext(ctx, state), 15*time.Second)
		return ctx, nil
	})

	sc.Step(`^there are (\d+) godogs$`, thereAreGodogs)
	sc.Step(`^I eat (\d+)$`, iEat)
	sc.Step(`^there should be (\d+) remaining$`, thereShouldBeRemaining)
}
```
We can see this style is more cumbersome but context now like an isolated thread-safe store, data can no longer be modified in another goroutine.  
Since the context is propagated by godog runner (note the context return), we're now can control the whole scenario execution, the above example will canceled after 15 seconds, including all API calls or DB queries.  

### Hints for writing test
#### Get the state when you need
Consider an example that you're writing `yourNewFancyFunc`:
```go
func someOneElseFuncFoo(ctx context.Context) (context.Context, error) {
	// do some magic, nasty things
	state := StateFromContext(ctx)
	state.MagicNumber = 42
	return ctx, nil
}

// yourNewFancyFunc expect to increase MagicNumber to 50
func yourNewFancyFunc(ctx context.Context) (context.Context, error) {
	state := StateFromContext(ctx)
	if ctx, err := someOneElseFuncFoo(ctx); err !=nil {
		return ctx, err
	}

	state.MagicNumber += 8
	return StateToContext(ctx, state), nil
}
```
For some reason you expecting the `state.MagicNumber == 50` but the wrong place of state getter causing issue.  
The suggestion is get the state when you really need it:
```go
func yourNewFancyFunc(ctx context.Context) (context.Context, error) {
	if ctx, err := someOneElseFuncFoo(ctx); err !=nil {
		return ctx, err
	}

	state := StateFromContext(ctx)
	state.MagicNumber += 8
	return StateToContext(ctx, state), nil
}
```
#### Put the state back to context if you're not sure
Using the same example but with the `someOneElseFuncFoo`, actually we have an error in the first place.  
I modified state data but didn't put it back and `state` is just a copy of data, to make sure it work all the time, we need to put it back:
```go
func someOneElseFuncFoo(ctx context.Context) (context.Context, error) {
	// do some magic, nasty things
	state := StateFromContext(ctx)
	state.MagicNumber = 42
	return StateToContext(ctx, state), nil
}
```

