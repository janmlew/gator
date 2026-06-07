package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/janmlew/gator/internal/config"
	"github.com/janmlew/gator/internal/database"
)

// state holds the application state shared across command handlers: the
// database query layer and the config.
type state struct {
	db  *database.Queries
	cfg *config.Config
}

// command represents a parsed CLI command: its name and its arguments.
type command struct {
	name string
	args []string
}

// commands holds all the commands the CLI can handle, mapping each command
// name to its handler function.
type commands struct {
	handlers map[string]func(*state, command) error
}

// register adds a new handler function for a command name.
func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

// run executes the handler registered for the given command, if it exists.
func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}

// handlerLogin sets the current user in the config file.
// Usage: gator login <username>
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("login requires a single argument: the username")
	}

	username := cmd.args[0]
	if _, err := s.db.GetUser(context.Background(), username); err != nil {
		return fmt.Errorf("can't login as %q: user does not exist: %w", username, err)
	}

	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("user has been set to %s\n", username)
	return nil
}

// handlerRegister creates a new user in the database and sets them as the
// current user in the config.
// Usage: gator register <username>
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("register requires a single argument: the username")
	}

	name := cmd.args[0]
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("couldn't create user %q: %w", name, err)
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Printf("user %s was created\n", user.Name)
	log.Printf("user data: %+v\n", user)
	return nil
}

// handlerReset deletes all users from the database, resetting it to a blank
// state. Useful during development.
// Usage: gator reset
func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("couldn't reset users table: %w", err)
	}

	fmt.Println("database reset successfully: all users deleted")
	return nil
}
