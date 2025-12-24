package main

import "fmt"

type Command interface {
	Name() string
	Description() string
	Usage() string
	Run(rag *RAGSystem, args []string) error
}

var commands = make(map[string]Command)

func RegisterCommand(cmd Command) {
	commands[cmd.Name()] = cmd
}

func GetCommand(name string) (Command, bool) {
	cmd, ok := commands[name]
	return cmd, ok
}

func ListCommands() []Command {
	result := make([]Command, 0, len(commands))
	for _, cmd := range commands {
		result = append(result, cmd)
	}
	return result
}

func PrintCommandsHelp() {
	fmt.Println("\nCommands:")
	for _, cmd := range commands {
		fmt.Printf("  %-25s %s\n", cmd.Name(), cmd.Description())
	}
}
