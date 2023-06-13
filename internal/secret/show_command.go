package secret

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"github.com/google/subcommands"
	"github.com/jordansinko/untold"
	"github.com/jordansinko/untold/internal/cli"
	"golang.org/x/crypto/nacl/box"
	"os"
	"path/filepath"
)

type showCmd struct {
	environment, privateKey string
}

func NewShowCommand() subcommands.Command { return &showCmd{environment: untold.DefaultEnvironment} }

func (s *showCmd) Name() string { return "show-secret" }

func (s *showCmd) Synopsis() string { return "show secret's value" }

func (s *showCmd) Usage() string {
	return `untold show-secret [-env={environment}] [-key={decryption_key}] <secret_name>:
  Show decrypted secret value.
`
}

func (s *showCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.environment, "env", s.environment, "set environment")
	f.StringVar(&s.privateKey, "key", s.privateKey, "provide decryption key")
}

func (s *showCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	name := f.Arg(0)
	if name == "" {
		cli.Errorf("argument \"name\" is required")
		s.Usage()

		return subcommands.ExitUsageError
	}

	environment := s.environment
	if environment == "" || environment == untold.DefaultEnvironment {
		environment = untold.DefaultEnvironment
		cli.Warnf("No environment provided, using default - %q", environment)
	}

	md5Hash := md5.Sum([]byte(name))
	base64EncodedPrivateKey := []byte(s.privateKey)

	if _, err := os.Stat(filepath.Join(environment, hex.EncodeToString(md5Hash[:]))); os.IsNotExist(err) {
		cli.Errorf("secret %q for %q environment not found", name, environment)

		return subcommands.ExitUsageError
	}

	if _, err := os.Stat(environment + ".public"); os.IsNotExist(err) {
		cli.Errorf("public key for %q environment not found", environment)

		return subcommands.ExitFailure
	}

	if len(base64EncodedPrivateKey) == 0 {
		if _, err := os.Stat(environment + ".private"); os.IsNotExist(err) {
			cli.Errorf("private key for %q environment not found", environment)

			return subcommands.ExitFailure
		}
	}

	base64EncodedPublicKey, err := os.ReadFile(environment + ".public")
	if err != nil {
		cli.Wrapf(err, "read public key for %q environment", environment)

		return subcommands.ExitFailure
	}

	if len(base64EncodedPrivateKey) == 0 {
		base64EncodedPrivateKey, err = os.ReadFile(environment + ".private")
		if err != nil {
			cli.Wrapf(err, "read private key for %q environment", environment)

			return subcommands.ExitFailure
		}
	}

	base64EncodedContent, err := os.ReadFile(filepath.Join(environment, hex.EncodeToString(md5Hash[:])))
	if err != nil {
		cli.Wrapf(err, "read secret %q for %q environment", name, environment)

		return subcommands.ExitFailure
	}

	var publicKey, privateKey [32]byte
	publicKey, err = untold.DecodeBase64Key(base64EncodedPublicKey)
	if err != nil {
		cli.Wrapf(err, "decode base64 encoded public key for %q environment", environment)

		return subcommands.ExitFailure
	}

	privateKey, err = untold.DecodeBase64Key(base64EncodedPrivateKey)
	if err != nil {
		cli.Wrapf(err, "decode base64 encoded private key for %q environment", environment)

		return subcommands.ExitFailure
	}

	decodedContent, err := untold.Base64Decode(base64EncodedContent)
	if err != nil {
		cli.Wrapf(err, "decode secret %q key for %q environment", name, environment)

		return subcommands.ExitFailure
	}

	decryptedValue, ok := box.OpenAnonymous(nil, decodedContent, &publicKey, &privateKey)
	if !ok {
		cli.Errorf("can not decrypt secret %q", name)

		return subcommands.ExitFailure
	}

	cli.Successf("Secret's %q value is: %s", name, decryptedValue)

	return subcommands.ExitSuccess
}
