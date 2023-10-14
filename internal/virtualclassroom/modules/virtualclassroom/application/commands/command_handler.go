package commands

type CommandHandler interface {
	Execute() error
}
