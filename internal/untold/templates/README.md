# untold

`untold` allows to embed encrypted secrets into your Go application.
This way you can store encrypted secrets in code repository, and developers can
add new secrets without a need of external secret management tool.

Encryption and decryption is possible thanks to [golang.org/x/crypto/nacl/box](https://golang.org/x/crypto/nacl/box) package.

Go version 1.16 or later is required.

## Installation & Usage

```
$ go install github.com/jordansinko/untold/cmd/untold@v0.0.3-alpha

$ untold init
WARNING: Directory name not provided, using default - "untold"
WARNING: No environment provided, using default - "development"
SUCCESS: Vault structure with "development" environment initialized.

$ tree -a
.
└── untold
    ├── .gitignore // ignore .private keys
    ├── README.md // documentation for fellow developers
    ├── development // secrets storage for development environment
    │   └── .gitkeep
    ├── development.private // development environment decryption key
    └── development.public // development environment encryption key

2 directories, 5 files

$ cd untold/

$ untold add-secret secret                                                                                                                                            2 ↵
WARNING: No environment provided, using default - "development"
Enter value for "secret" secret in "development" environment:
sup3rs3cr3tvalu3
SUCCESS: Secret "secret" for "development" environment stored.

$ cd ..

$ go mod init example

$ go get github.com/jordansinko/untold@v0.0.3-alpha

$ touch main.go
```

Set content of `main.go` to

```go
package main

import (
	"embed"
	"fmt"

	"github.com/jordansinko/untold"
)

// Let's embed folder `untold` into application.
// Note that you will have to provide path prefix if your secrets directory is not named as `untold`

//go:embed untold
var untoldFS embed.FS

// Loader will look for tags `untold` within Config
// and will search for MD5 equivalent of tag's value
// in the embedded filesystem.

type Config struct {
	Secret string `untold:"secret"` // will look for secrets/development/MD5("secret")
}

func main() {
	var config Config

	// all options are optional in this case
	err := untold.NewVault(
		untoldFS,                          // provide embedded FS
		untold.PathPrefix("untold"),       // directory where your secrets are stored (default "untold")
		untold.Environment("development"), // environment name (default "development")
		untold.EnvVariable("UNTOLD_KEY"),  // environment variable (default "UNTOLD_KEY")
	).Load(&config)
	if err != nil {
		panic(err)
	}

	fmt.Println(config.Secret)
	// output: sup3rs3cr3tvalu3
}
```

Compile and run:

```
$ go build -o example main.go
$ UNTOLD_KEY=$(cat untold/development.private) ./example
```

If environment variable `UNTOLD_KEY` is not provided, `untold` will look for `{environment_name}.private`
in embedded filesystem. So, because we have embedded `untold/development.private` key, this will also work:

```
$ go build -o example main.go
$ ./example
```

## Important

Encrypted passwords are not completely secure. You should never store your passwords
in public repositories, because bad actors can try to decrypt them.
In case of source code leak you should rotate your keys and secrets immediately.

Never store private keys in your repository. Private key should be stored someplace secure
and provided to the application within environment variable `UNTOLD_KEY`.
Only exception is your local development environment private key.

This library does not obfuscate your secrets. If you distribute binaries of your application
be aware, that encrypted secrets can be found in the binary with simple `grep -a` command.
This library was built with web services in mind.
