// Command gophercram packs a software repository into a single AI-friendly
// document. See `gophercram --help` for the full option list, or the docs
// directory for an in-depth user guide.
package main

import (
	"os"

	"github.com/rajathjn/GopherCram/internal/cli"
)

func main() {
	code := cli.Run(cli.RunOptions{
		Argv:   os.Args[1:],
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	os.Exit(int(code))
}
