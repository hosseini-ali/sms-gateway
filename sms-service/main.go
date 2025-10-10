package main

import (
	"fmt"
	"os"

	"notif/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
