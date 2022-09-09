package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/GreptimeTeam/greptimedb-operator/cmd/operator/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	command := app.NewOperatorCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
