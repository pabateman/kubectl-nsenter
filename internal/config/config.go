package config

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	KubeConfig         string
	KubeContext        string
	Namespace          string
	PodName            string
	Container          string
	Command            []string
	SshUser            string
	SshRequirePassword bool
	SshSocketPath      string
	SshHost            string
	SshPort            string
	LinuxNs            []string
}

const (
	argKubeconfig  = "kubeconfig"
	argContainer   = "container"
	argContext     = "context"
	argNamespace   = "namespace"
	argUser        = "user"
	argPassword    = "password"
	argSshAuthSock = "ssh-auth-sock"
	argHost        = "host"
	argPort        = "port"
	argNs          = "ns"
)

var (
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        argKubeconfig,
			Usage:       "kubernetes client config path",
			EnvVars:     []string{"KUBECONFIG"},
			Value:       fmt.Sprintf("%s/.kube/config", homedir.HomeDir()),
			Required:    false,
			DefaultText: "$HOME/.kube/config",
		},
		&cli.StringFlag{
			Name:     argContainer,
			Aliases:  []string{"c"},
			Usage:    "use namespace of specified container. By default first running container will taken",
			Value:    "",
			Required: false,
		},
		&cli.StringFlag{
			Name:     argContext,
			Usage:    "override current context from kubeconfig",
			Value:    "",
			Required: false,
		},
		&cli.StringFlag{
			Name:     argNamespace,
			Aliases:  []string{"n"},
			Usage:    "override namespace of current context from kubeconfig",
			Value:    "",
			Required: false,
		},
		&cli.StringFlag{
			Name:     argUser,
			Aliases:  []string{"u"},
			Usage:    "set username for ssh connection to node",
			EnvVars:  []string{"USER"},
			Required: false,
		},
		&cli.BoolFlag{
			Name:    argPassword,
			Aliases: []string{"s"},
			Usage:   "force ask for node password prompt",
			Value:   false,
		},
		&cli.StringFlag{
			Name:        argSshAuthSock,
			Usage:       "sets ssh-agent socket",
			EnvVars:     []string{"SSH_AUTH_SOCK"},
			DefaultText: "current shell auth sock",
			Required:    false,
		},
		&cli.StringFlag{
			Name:     argHost,
			Usage:    "override node ip",
			Required: false,
		},
		&cli.StringFlag{
			Name:     argPort,
			Aliases:  []string{"p"},
			Usage:    "sets ssh port",
			Value:    "22",
			Required: false,
		},
		&cli.StringSliceFlag{
			Name:     argNs,
			Usage:    "define container's pid linux namespaces to enter. Sends transparently to nsenter cmd",
			Value:    cli.NewStringSlice("n"),
			Required: false,
		},
	}
	stringFlags = []string{argKubeconfig, argContainer, argContext, argNamespace, argUser, argSshAuthSock, argHost, argPort}
)

func NewConfig(clictx *cli.Context) (Config, error) {
	podName := clictx.Args().First()
	if podName == "" {
		return Config{}, errorWithCliHelp(clictx, "you must specify pod name!")
	}

	command := clictx.Args().Tail()
	if len(command) == 0 {
		return Config{}, errorWithCliHelp(clictx, "you must provide a command!")
	}

	err := validateStringFlagsNonEmpty(clictx, stringFlags)
	if err != nil {
		return Config{}, err
	}

	return Config{
		KubeConfig:         clictx.String(argKubeconfig),
		KubeContext:        clictx.String(argContext),
		Namespace:          clictx.String(argNamespace),
		PodName:            podName,
		Container:          clictx.String(argContainer),
		Command:            command,
		SshUser:            clictx.String(argUser),
		SshSocketPath:      clictx.String(argSshAuthSock),
		SshRequirePassword: clictx.Bool(argPassword),
		SshHost:            clictx.String(argHost),
		SshPort:            clictx.String(argPort),
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

func errorWithCliHelp(clictx *cli.Context, msg string) error {
	err := cli.ShowAppHelp(clictx)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s\n", msg)
}

func errorWithCliHelpf(clictx *cli.Context, format string, a ...any) error {
	return errorWithCliHelp(clictx, fmt.Sprintf(format, a...))
}
