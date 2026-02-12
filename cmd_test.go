package main

import (
	"strings"
	"testing"
)

func TestAddCommand_MissingArgs(t *testing.T) {
	cmd := &AddCommand{}
	err := cmd.Run(nil, []string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "usage") {
		t.Fatalf("expected error containing 'usage', got: %s", err.Error())
	}
}

func TestAddCommand_OneArg(t *testing.T) {
	cmd := &AddCommand{}
	err := cmd.Run(nil, []string{"doc1"})
	if err == nil {
		t.Fatal("expected error for only one arg")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "usage") {
		t.Fatalf("expected error containing 'usage', got: %s", err.Error())
	}
}

func TestAddCommand_Name(t *testing.T) {
	cmd := &AddCommand{}
	if cmd.Name() != "add" {
		t.Fatalf("expected name 'add', got: %s", cmd.Name())
	}
}

func TestAddCommand_Description(t *testing.T) {
	cmd := &AddCommand{}
	if cmd.Description() == "" {
		t.Fatal("expected non-empty description")
	}
}

func TestDeleteCommand_MissingArgs(t *testing.T) {
	cmd := &DeleteCommand{}
	err := cmd.Run(nil, []string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "usage") {
		t.Fatalf("expected error containing 'usage', got: %s", err.Error())
	}
}

func TestDeleteCommand_Name(t *testing.T) {
	cmd := &DeleteCommand{}
	if cmd.Name() != "delete" {
		t.Fatalf("expected name 'delete', got: %s", cmd.Name())
	}
}

func TestQueryCommand_MissingArgs(t *testing.T) {
	cmd := &QueryCommand{}
	err := cmd.Run(nil, []string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "usage") {
		t.Fatalf("expected error containing 'usage', got: %s", err.Error())
	}
}

func TestQueryCommand_Name(t *testing.T) {
	cmd := &QueryCommand{}
	if cmd.Name() != "query" {
		t.Fatalf("expected name 'query', got: %s", cmd.Name())
	}
}

func TestServeCommand_Name(t *testing.T) {
	cmd := &ServeCommand{}
	if cmd.Name() != "serve" {
		t.Fatalf("expected name 'serve', got: %s", cmd.Name())
	}
}

func TestListCommand_Name(t *testing.T) {
	cmd := &ListCommand{}
	if cmd.Name() != "list" {
		t.Fatalf("expected name 'list', got: %s", cmd.Name())
	}
}
