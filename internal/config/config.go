package config

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"
)

const (
	argKubeconfig  = "kubeconfig"
	argContainer   = "container"
	argContext     = "context"
	argNamespace   = "namespace"
	argUser        = "user"
	argPassword    = "password"
	argSSHAuthSock = "ssh-auth-sock"
	argSSHOpts     = "ssh-opt"
	argHost        = "host"
	argPort        = "port"
	argNs          = "ns"
	argInteractive = "interactive"
	argTTY         = "tty"
	argUseNodeName = "use-node-name"
)

var (
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        argKubeconfig,
			Usage:       "kubernetes client config path",
			EnvVars:     []string{"KUBECONFIG"},
			Value:       fmt.Sprintf("%s/.kube/config", homedir.HomeDir()),
			DefaultText: "$HOME/.kube/config",
		},
		&cli.StringFlag{
			Name:    argContainer,
			Aliases: []string{"c"},
			Usage:   "use namespace of specified container. By default first running container will taken",
			Value:   "",
		},
		&cli.StringFlag{
			Name:  argContext,
			Usage: "override current context from kubeconfig",
			Value: "",
		},
		&cli.StringFlag{
			Name:    argNamespace,
			Aliases: []string{"n"},
			Usage:   "override namespace of current context from kubeconfig",
			Value:   "",
		},
		&cli.StringFlag{
			Name:    argUser,
			Aliases: []string{"u"},
			Usage:   "set username for ssh connection to node",
		},
		&cli.BoolFlag{
			Name:    argPassword,
			Aliases: []string{"s"},
			Usage:   "force ask for node password prompt",
		},
		&cli.StringFlag{
			Name:        argSSHAuthSock,
			Usage:       "sets ssh-agent socket",
			EnvVars:     []string{"SSH_AUTH_SOCK"},
			DefaultText: "current shell auth sock",
		},
		&cli.StringFlag{
			Name:  argHost,
			Usage: "override node ip",
		},
		&cli.StringFlag{
			Name:    argPort,
			Aliases: []string{"p"},
			Usage:   "sets ssh port",
		},
		&cli.StringSliceFlag{
			Name:  argNs,
			Usage: "define container's pid linux namespaces to enter. Sends transparently to nsenter cmd",
			Value: cli.NewStringSlice("n"),
		},
		&cli.BoolFlag{
			Name:    argInteractive,
			Aliases: []string{"i"},
			Usage:   "keep ssh session stdin",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    argTTY,
			Aliases: []string{"t"},
			Usage:   "allocate pseudo-TTY for ssh session",
			Value:   false,
		},
		&cli.StringSliceFlag{
			Name:    argSSHOpts,
			Aliases: []string{"o"},
			Usage:   "same as -o for ssh client",
		},
		&cli.BoolFlag{
			Name:    argUseNodeName,
			Aliases: []string{"j"},
			Usage:   "use kubernetes node name to connect with ssh. Useful with ssh configs",
			EnvVars: []string{"KUBECTL_NSENTER_USE_NODE_NAME"},
			Value:   true,
		},
	}
	stringFlags = []string{argKubeconfig, argContainer, argContext, argNamespace, argUser, argSSHAuthSock, argHost, argPort}
)

type Config struct {
	KubeConfig         string
	KubeContext        string
	Namespace          string
	PodName            string
	Container          string
	Command            []string
	SSHUser            string
	SSHRequirePassword bool
	SSHSocketPath      string
	SSHHost            string
	SSHPort            string
	SSHOpts            []string
	Interactive        bool
	TTY                bool
	UseNodeName        bool
	LinuxNs            []string
}

func NewConfig(clictx *cli.Context) (*Config, error) {
	podName := clictx.Args().First()
	if podName == "" {
		return nil, errorWithCliHelp(clictx, "you must specify pod name!")
	}

	command := clictx.Args().Tail()
	if len(command) == 0 {
		return nil, errorWithCliHelp(clictx, "you must provide a command!")
	}

	err := validateStringFlagsNonEmpty(clictx, stringFlags)
	if err != nil {
		return nil, err
	}

	return &Config{
		KubeConfig:         clictx.String(argKubeconfig),
		KubeContext:        clictx.String(argContext),
		Namespace:          clictx.String(argNamespace),
		PodName:            podName,
		Container:          clictx.String(argContainer),
		Command:            command,
		SSHUser:            clictx.String(argUser),
		SSHSocketPath:      clictx.String(argSSHAuthSock),
		SSHRequirePassword: clictx.Bool(argPassword),
		SSHHost:            clictx.String(argHost),
		SSHPort:            clictx.String(argPort),
		SSHOpts:            clictx.StringSlice(argSSHOpts),
		Interactive:        clictx.Bool(argInteractive),
		TTY:                clictx.Bool(argTTY),
		UseNodeName:        clictx.Bool(argUseNodeName),
		LinuxNs:            clictx.StringSlice(argNs),
	}, nil
}

func validateStringFlagsNonEmpty(clictx *cli.Context, flags []string) error {
	for _, flag := range flags {
		if clictx.IsSet(flag) {
			if clictx.String(flag) == "" {
				return errorWithCliHelpf(clictx, "option --%s must not be empty", flag)
			}
		}
	}
	return nil
}

func errorWithCliHelp(clictx *cli.Context, a any) error {
	err := cli.ShowAppHelp(clictx)
	if err != nil {
		return err
	}
	//nolint:staticcheck
	return fmt.Errorf("%s\n", a)
}

func errorWithCliHelpf(clictx *cli.Context, format string, a ...any) error {
	return errorWithCliHelp(clictx, fmt.Sprintf(format, a...))
}
