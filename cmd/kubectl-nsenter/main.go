package main

import (
	"github.com/pabateman/kubectl-nsenter/internal/nsenter"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"

	"fmt"
	"os"
	"time"
)

func main() {
	app := &cli.App{
		Name:     "kubectl-nsenter",
		Version:  "v0.1.0",
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
		Action:                 nsenter.Nsenter,
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
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"u"},
				Usage:    "set username for ssh connection to node",
				EnvVars:  []string{"USER"},
				Required: false,
			},
			&cli.BoolFlag{
				Name:    "password",
				Aliases: []string{"s"},
				Usage:   "force ask for node password prompt",
				Value:   false,
			},
			&cli.StringFlag{
				Name:        "ssh-auth-sock",
				Usage:       "sets ssh-agent socket",
				EnvVars:     []string{"SSH_AUTH_SOCK"},
				DefaultText: "current shell auth sock",
				Required:    false,
			},
			&cli.StringFlag{
				Name:     "host",
				Usage:    "override node ip",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Usage:    "sets ssh port",
				Value:    "22",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     "ns",
				Usage:    "define container's pid linux namespaces to enter. sends transparently to nsenter cmd",
				Value:    cli.NewStringSlice("n"),
				Required: false,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%+v: %+v", os.Args[0], err)
	}
}
