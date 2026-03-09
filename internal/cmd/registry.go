// Package cmd provides infrastructure for registering and managing TUI commands.
// Commands are organized as separate applications that can be launched via the main hostr CLI.
package cmd

// Registry manages available commands.
// This allows adding new commands without modifying the main entry point.
type Registry struct {
	commands map[string]Command
}

// Command represents a single TUI command or sub-application.
type Command interface {
	// Name returns the display name of the command.
	Name() string

	// Description returns a brief description of what the command does.
	Description() string

	// Execute runs the command and returns any error.
	// The context can be used for graceful shutdown.
	Execute() error
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(name string, cmd Command) {
	r.commands[name] = cmd
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// List returns all registered commands.
func (r *Registry) List() map[string]Command {
	return r.commands
}

// Default returns the default command (sysinfo).
func (r *Registry) Default() (Command, bool) {
	return r.Get("sysinfo")
}
