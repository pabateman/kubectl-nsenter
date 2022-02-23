package main

import (
	"fmt"
	"github.com/pabateman/kubectl-nsenter/internal/nsenter"
	"github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"

	"os"
	"os/user"
	"time"
)

func getSystemUser() *user.User {
	systemUser, _ := user.Current()
	return systemUser
}

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
				Value:    getSystemUser().Username,
				Required: false,
			},
			&cli.StringFlag{
				Name:     "ssh-auth-sock",
				Usage:    "sets ssh-agent socket",
				EnvVars:  []string{"SSH_AUTH_SOCK"},
				Required: false,
			},
			&cli.StringFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Usage:    "sets ssh port",
				Value:    "22",
				Required: false,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%+v: %+v", os.Args[0], err)
	}
}
