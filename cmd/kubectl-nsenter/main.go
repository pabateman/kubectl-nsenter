package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"

	"os"
	"time"
)

func main() {
	app := &cli.App{
		Name:     "kubectl-nsenter",
		Version:  "0.1.0",
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
		UsageText:              "kubectl-nsenter [pod name] [flags] [command]",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		HideHelpCommand:        true,
		Action:                 nsenter,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "kubeconfig",
				Usage:       "kubernetes client config path",
				EnvVars:     []string{"KUBECONFIG"},
				Value:       fmt.Sprintf("%s/.kube/config", homedir.HomeDir()),
				Required:    false,
				DefaultText: "$HOME/.kube/config",
			},
			&cli.StringFlag{
				Name:     "container",
				Aliases:  []string{"c"},
				Usage:    "use namespace of specified container. By default first container will taken",
				Value:    "",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "context",
				Usage:    "override current context from kubeconfig",
				Value:    "",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "namespace",
				Aliases:  []string{"n"},
				Usage:    "override namespace of current context from kubeconfig",
				Value:    "",
				Required: false,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%v:\n%+v", os.Args[0], err)
	}
}
