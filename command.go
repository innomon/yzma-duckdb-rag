package main

import "fmt"

// Command is the interface that all CLI subcommands must implement.
type Command interface {
	Name() string
	Description() string
	Usage() string
	Run(rag *RAGSystem, args []string) error
}

// commands holds the registry of all available CLI commands keyed by name.
var commands = make(map[string]Command)

// RegisterCommand adds a Command to the global command registry.
func RegisterCommand(cmd Command) {
	commands[cmd.Name()] = cmd
}

// GetCommand returns the Command registered under the given name and a boolean indicating whether it was found.
func GetCommand(name string) (Command, bool) {
	cmd, ok := commands[name]
	return cmd, ok
}

// ListCommands returns a slice of all registered commands.
func ListCommands() []Command {
	result := make([]Command, 0, len(commands))
	for _, cmd := range commands {
		result = append(result, cmd)
	}
	return result
}

// PrintCommandsHelp prints a formatted list of all registered commands and their descriptions to stdout.
func PrintCommandsHelp() {
	fmt.Println("\nCommands:")
	for _, cmd := range commands {
		fmt.Printf("  %-25s %s\n", cmd.Name(), cmd.Description())
	}
}
