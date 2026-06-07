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

// handlerUsers prints all registered users, marking the currently logged-in
// user with "(current)".
// Usage: gator users
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list users: %w", err)
	}

	for _, user := range users {
		line := "* " + user.Name
		if user.Name == s.cfg.CurrentUserName {
			line += " (current)"
		}
		fmt.Println(line)
	}
	return nil
}

// handlerAgg fetches a single RSS feed and prints it. Placeholder for the
// future long-running aggregator service.
// Usage: gator agg
func handlerAgg(s *state, cmd command) error {
	const feedURL = "https://www.wagslane.dev/index.xml"

	feed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("couldn't fetch feed %q: %w", feedURL, err)
	}

	fmt.Printf("%+v\n", *feed)
	return nil
}

// handlerAddFeed creates a new feed owned by the currently logged-in user.
// Usage: gator addfeed <name> <url>
func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return errors.New("addfeed requires two arguments: name and url")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("couldn't get current user: %w", err)
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	fmt.Printf("%+v\n", feed)
	return nil
}

// handlerFeeds prints every feed in the database along with the user who added
// it.
// Usage: gator feeds
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't list feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("* %s\n", feed.Name)
		fmt.Printf("  url:   %s\n", feed.Url)
		fmt.Printf("  added by: %s\n", feed.UserName)
	}
	return nil
}
