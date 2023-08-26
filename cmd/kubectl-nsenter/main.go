package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pabateman/kubectl-nsenter/internal/config"
	"github.com/pabateman/kubectl-nsenter/internal/nsenter"

	cli "github.com/urfave/cli/v2"
)

var Version = "local"

func main() {
	app := &cli.App{
		Name:     "kubectl-nsenter",
		Version:  Version,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name: "pabateman",
			},
		},
		Copyright: "© 2022 pabateman",
		HelpName:  "kubectl-nsenter",
		Usage: "kubectl plugin for pod's linux namespaces command execution " +
			"via direct node ssh connection",
		UsageText: `kubectl-nsenter [flags] [pod name] [command]

		Example:

		kubectl-nsenter -u node_user sample-pod-0 ip address

		kubectl-nsenter -u node_user -p 2222 postgres-1 tcpdump -nni any port 5432
		`,
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		HideHelpCommand:        true,
		Action:                 nsenter.Nsenter,
		Flags:                  config.Flags,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%+v: %+v", os.Args[0], err)
	}
}
