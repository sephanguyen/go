package bootstrap

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/multierr"
)

var rootCmd = &cobra.Command{
	Use: "gserver",
}

var jobCmd = &cobra.Command{
	Use: "gjob",
}

// RunServer runs the server with the underlying root cobra.Command.
// It is generally used in a main function. To integrate the underlying
// root command to another command, use AddCommand instead.
func RunServer(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

// RunServer runs the server with the underlying root cobra.Command.
// It is generally used in a main function. To integrate the underlying
// root command to another command, use AddCommand instead.
func RunJob(ctx context.Context) error {
	return jobCmd.ExecuteContext(ctx)
}

// AddCommand adds the underlying root cobra.Command to cmd,
// so that the underlying root cobra.Command can be run using the cmd.
func AddCommand(cmd *cobra.Command) {
	cmd.AddCommand(rootCmd)
	cmd.AddCommand(jobCmd)
}

type RegisterOpt[T any] struct {
	http    HTTPServicer[T]
	grpc    GRPCServicer[T]
	nats    NatsServicer[T]
	kafka   KafkaServicer[T]
	monitor MonitorServicer[T]

	cmd *cobra.Command
}

func newRegisterOpt[T any]() *RegisterOpt[T] {
	return &RegisterOpt[T]{cmd: &cobra.Command{}}
}

// WithGRPC should be used to register a GRPC service.
func WithGRPC[T any](s GRPCServicer[T]) *RegisterOpt[T] {
	return newRegisterOpt[T]().WithGRPC(s)
}

// WithHTTP should be used to register a HTTP service.
func WithHTTP[T any](s HTTPServicer[T]) *RegisterOpt[T] {
	return newRegisterOpt[T]().WithHTTP(s)
}

// WithNatsServicer should be used to register a service that
// publishes/subscribes to NATS Jetstream.
func WithNatsServicer[T any](s NatsServicer[T]) *RegisterOpt[T] {
	return newRegisterOpt[T]().WithNatsServicer(s)
}

// WithKafkaServicer should be used to register a service that
// publishes/subscribes to Kafka.
func WithKafkaServicer[T any](s KafkaServicer[T]) *RegisterOpt[T] {
	return newRegisterOpt[T]().WithKafkaServicer(s)
}

// WithMonitorServicer should be used to register a service that
// exposes monitoring metrics and traces.
func WithMonitorServicer[T any](s MonitorServicer[T]) *RegisterOpt[T] {
	return newRegisterOpt[T]().WithMonitorServicer(s)
}

// WithGRPC should be used to register a GRPC service.
// This is a method similar to bootstrap.WithGRPC that can be used for method chaining.
func (o *RegisterOpt[T]) WithGRPC(s GRPCServicer[T]) *RegisterOpt[T] {
	o.grpc = s
	return o
}

// WithHTTP should be used to register a service that publishes/subscribes
// to NATS Jetstream.
// This is a method similar to bootstrap.WithHTTP that can be used for method chaining.
func (o *RegisterOpt[T]) WithHTTP(s HTTPServicer[T]) *RegisterOpt[T] {
	o.http = s
	return o
}

// WithNatsServicer should be used to register a service that
// publishes/subscribes to NATS Jetstream.
// This is a method similar to bootstrap.WithNatsServicer that can be used for method chaining.
func (o *RegisterOpt[T]) WithNatsServicer(s NatsServicer[T]) *RegisterOpt[T] {
	o.nats = s
	return o
}

// WithKafkaServicer should be used to register a service that
// publishes/subscribes to Kafka.
// This is a method similar to bootstrap.WithKafkaServicer that can be used for method chaining.
func (o *RegisterOpt[T]) WithKafkaServicer(s KafkaServicer[T]) *RegisterOpt[T] {
	o.kafka = s
	return o
}

// WithMonitorServicer should be used to register a service that
// exposes monitoring metrics and traces.
// This is a method similar to bootstrap.WithMonitorServicer that can be used for method chaining.
func (o *RegisterOpt[T]) WithMonitorServicer(s MonitorServicer[T]) *RegisterOpt[T] {
	o.monitor = s
	return o
}

// FlagStringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (o *RegisterOpt[T]) FlagStringVar(p *string, name string, value string, usage string) *RegisterOpt[T] {
	o.cmd.Flags().StringVar(p, name, value, usage)
	return o
}

// FlagBoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func (o *RegisterOpt[T]) FlagBoolVar(p *bool, name string, value bool, usage string) *RegisterOpt[T] {
	o.cmd.Flags().BoolVar(p, name, value, usage)
	return o
}

// FlagIntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func (o *RegisterOpt[T]) FlagIntVar(p *int, name string, value int, usage string) *RegisterOpt[T] {
	o.cmd.Flags().IntVar(p, name, value, usage)
	return o
}

// MarkFlagRequired instructs the various shell completion implementations to
// prioritize the named flag when performing completion,
// and causes your command to report an error if invoked without the flag.
func (o *RegisterOpt[T]) MarkFlagRequired(flagname string) *RegisterOpt[T] {
	if err := o.cmd.MarkFlagRequired(flagname); err != nil {
		panic(fmt.Errorf("failed to mark flag as required: %w", err))
	}
	return o
}

// Register finalizes and registers the current service to the root cobra.Command.
// The service can be invoked with `go run ./cmd/server/main.go gserver <service-name>`.
func (o *RegisterOpt[T]) Register(s BaseServicer[T]) {
	cmd := o.cmd
	cmd.Use = s.ServerName()
	cmd.Short = "Start " + s.ServerName() + " server"
	cmd.Args = cobra.NoArgs
	cmd.SilenceUsage = true
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		bst := newBootstrapper[T]()
		return bst.run(cmd, s)
	}
	if err := setCommonServerFlags(cmd); err != nil {
		panic(fmt.Errorf("failed to set server common flags: %w", err))
	}
	rootCmd.AddCommand(cmd)
}

const (
	commonConfigPathFlag = "commonConfigPath"
	configPathFlag       = "configPath"
	secretPathFlag       = "secretsPath"
)

func setCommonServerFlags(cmd *cobra.Command) error {
	cmd.Flags().String(commonConfigPathFlag, "", "path to the common config yaml file")
	cmd.Flags().String(configPathFlag, "", "path to the config yaml file")
	cmd.Flags().String(secretPathFlag, "", "path to the encrypted secret file")
	return multierr.Combine(
		cmd.MarkFlagFilename(commonConfigPathFlag),
		cmd.MarkFlagRequired(commonConfigPathFlag),
		cmd.MarkFlagFilename(configPathFlag),
		cmd.MarkFlagRequired(configPathFlag),
		cmd.MarkFlagFilename(secretPathFlag),
		cmd.MarkFlagRequired(secretPathFlag),
	)
}

type JobFunc[T any] func(ctx context.Context, c T, rsc *Resources) error

// JobBuilder please add more [Type]Var function if you need, now we only support a few
type JobBuilder[t any] struct {
	name string
	cmd  *cobra.Command
}

func (b *JobBuilder[T]) StringVar(p *string, name string, value string, usage string) *JobBuilder[T] {
	b.cmd.Flags().StringVar(p, name, value, usage)
	return b
}

func (b *JobBuilder[T]) BoolVar(p *bool, name string, value bool, usage string) *JobBuilder[T] {
	b.cmd.Flags().BoolVar(p, name, value, usage)
	return b
}

func (b *JobBuilder[T]) IntVar(p *int, name string, value int, usage string) *JobBuilder[T] {
	b.cmd.Flags().IntVar(p, name, value, usage)
	return b
}

func (b *JobBuilder[T]) BytesBase64Var(p *[]byte, name string, value []byte, usage string) *JobBuilder[T] {
	b.cmd.Flags().BytesBase64Var(p, name, value, usage)
	return b
}

// Desc adds a short description (cobra.Command.Short) for this job.
func (b *JobBuilder[T]) Desc(s string) *JobBuilder[T] {
	b.cmd.Short = s
	return b
}

// DescLong adds a long description (cobra.Command.Long) for this job.
func (b *JobBuilder[T]) DescLong(s string) *JobBuilder[T] {
	b.cmd.Short = s
	return b
}

// RegisterJob a job callable from gserver, returning flagset allowing configure custom runtime flag.
//
// Note that the context argument of JobFunc s already handles signal processing
// (e.g. SIGTERM...) so you don't need to handle it again.
func RegisterJob[T any](name string, s JobFunc[T]) *JobBuilder[T] {
	cmd := &cobra.Command{
		Use:   name,
		Short: "Start " + name + " job",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bst := newBootstrapper[T]()
			return bst.runJob(cmd, s)
		},
		SilenceUsage: true,
	}
	if err := setCommonServerFlags(cmd); err != nil {
		panic(fmt.Errorf("failed to register server: %w", err))
	}
	jobCmd.AddCommand(cmd)
	return &JobBuilder[T]{name: name, cmd: cmd}
}
