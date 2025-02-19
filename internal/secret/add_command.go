package secret

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/jordansinko/untold"
	"github.com/jordansinko/untold/internal/cli"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

type addCmd struct {
	environment string
}

func NewAddCommand() subcommands.Command { return &addCmd{environment: untold.DefaultEnvironment} }

func (a *addCmd) Name() string { return "add-secret" }

func (a *addCmd) Synopsis() string { return "add new secret" }

func (a *addCmd) Usage() string {
	return `untold add-secret [-env={environment}] <secret_name>:
  Add new secret.
`
}

func (a *addCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.environment, "env", a.environment, "set environment")
}

func (a *addCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	name := f.Arg(0)
	if name == "" {
		cli.Errorf("argument \"name\" is required")
		a.Usage()

		return subcommands.ExitUsageError
	}

	environment := a.environment
	if environment == "" || environment == untold.DefaultEnvironment {
		environment = untold.DefaultEnvironment
		cli.Warnf("No environment provided, using default - %q", environment)
	}

	md5Hash := md5.Sum([]byte(name))

	if _, err := os.Stat(filepath.Join(environment, hex.EncodeToString(md5Hash[:]))); !os.IsNotExist(err) {
		cli.Errorf("secret %q for %q environment already exists", name, environment)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environment); os.IsNotExist(err) {
		cli.Errorf("directory for %q environment not found", environment)

		return subcommands.ExitFailure
	}

	if _, err := os.Stat(environment + ".public"); os.IsNotExist(err) {
		cli.Errorf("public key for %q environment not found", environment)

		return subcommands.ExitFailure
	}

	base64EncodedPublicKey, err := os.ReadFile(environment + ".public")
	if err != nil {
		cli.Wrapf(err, "read public key for %q environment", environment)

		return subcommands.ExitFailure
	}

	var publicKey [32]byte
	publicKey, err = untold.DecodeBase64Key(base64EncodedPublicKey)
	if err != nil {
		cli.Wrapf(err, "decode base64 encoded public key for %q environment", environment)

		return subcommands.ExitFailure
	}

	fmt.Printf("Enter value for %q secret in %q environment:\n", name, environment)

	var value string
	_, err = fmt.Scanln(&value)
	if err != nil {
		cli.Wrapf(err, "read user input")

		return subcommands.ExitFailure
	}

	encryptedValue, err := box.SealAnonymous(nil, []byte(value), &publicKey, rand.Reader)
	if err != nil {
		cli.Wrapf(err, "encrypt user input")

		return subcommands.ExitFailure
	}

	err = os.WriteFile(filepath.Join(environment, hex.EncodeToString(md5Hash[:])), untold.Base64Encode(encryptedValue), 0644)
	if err != nil {
		cli.Wrapf(err, "write secret %q for %q environment to file", name, environment)

		return subcommands.ExitFailure
	}

	cli.Successf("Secret %q for %q environment stored.", name, environment)

	return subcommands.ExitSuccess
}
