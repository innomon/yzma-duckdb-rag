package main

import (
	"sort"
	"testing"
)

// mockCommand implements Command for testing.
type mockCommand struct {
	name        string
	description string
	usage       string
}

func (m *mockCommand) Name() string                            { return m.name }
func (m *mockCommand) Description() string                     { return m.description }
func (m *mockCommand) Usage() string                           { return m.usage }
func (m *mockCommand) Run(rag *RAGSystem, args []string) error { return nil }

func TestGetCommand_Exists(t *testing.T) {
	expected := []string{"add", "delete", "list", "query", "serve"}
	for _, name := range expected {
		cmd, ok := GetCommand(name)
		if !ok {
			t.Errorf("expected command %q to be registered, but it was not", name)
			continue
		}
		if cmd.Name() != name {
			t.Errorf("expected command name %q, got %q", name, cmd.Name())
		}
	}
}

func TestGetCommand_NotExists(t *testing.T) {
	_, ok := GetCommand("nonexistent")
	if ok {
		t.Error("expected GetCommand(\"nonexistent\") to return false, got true")
	}
}

func TestListCommands(t *testing.T) {
	cmds := ListCommands()

	expected := []string{"add", "delete", "list", "query", "serve"}

	if len(cmds) < len(expected) {
		t.Fatalf("expected at least %d commands, got %d", len(expected), len(cmds))
	}

	names := make([]string, len(cmds))
	for i, cmd := range cmds {
		names[i] = cmd.Name()
	}
	sort.Strings(names)
	sort.Strings(expected)

	for _, exp := range expected {
		found := false
		for _, name := range names {
			if name == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected ListCommands to include %q, but it was missing (got %v)", exp, names)
		}
	}
}

func TestRegisterCommand(t *testing.T) {
	const testName = "mock-test-cmd"

	mock := &mockCommand{
		name:        testName,
		description: "a mock command for testing",
		usage:       "mock-test-cmd [args]",
	}

	RegisterCommand(mock)

	cmd, ok := GetCommand(testName)
	if !ok {
		t.Fatalf("expected command %q to be registered after RegisterCommand", testName)
	}
	if cmd.Name() != testName {
		t.Errorf("expected command name %q, got %q", testName, cmd.Name())
	}
	if cmd.Description() != mock.description {
		t.Errorf("expected description %q, got %q", mock.description, cmd.Description())
	}
	if cmd.Usage() != mock.usage {
		t.Errorf("expected usage %q, got %q", mock.usage, cmd.Usage())
	}

	// Clean up the global commands map.
	delete(commands, testName)

	if _, ok := GetCommand(testName); ok {
		t.Error("expected mock command to be removed after cleanup")
	}
}
