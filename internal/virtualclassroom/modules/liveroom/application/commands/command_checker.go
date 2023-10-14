package commands

type CommandChecker interface {
	Check(command ModifyStateCommand) error
}
