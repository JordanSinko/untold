package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"github.com/jordansinko/untold/internal/secret"
	"github.com/jordansinko/untold/internal/untold"
	"github.com/jordansinko/untold/internal/vault"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	subcommands.Register(untold.NewInitCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")

	subcommands.Register(vault.NewCreateCommand(), "vault management")
	subcommands.Register(vault.NewRotateCommand(), "vault management")

	subcommands.Register(secret.NewAddCommand(), "secrets")
	subcommands.Register(secret.NewShowCommand(), "secrets")
	subcommands.Register(secret.NewChangeCommand(), "secrets")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
