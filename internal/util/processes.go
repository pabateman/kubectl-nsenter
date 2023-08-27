package util

import (
	"fmt"
	"os"
	"os/exec"
)

func Fork(cmd []string, isInteractive bool) error {
	c := exec.Command(cmd[0], cmd[1:]...)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if isInteractive {
		c.Stdin = os.Stdin
	}

	err := c.Start()
	if err != nil {
		return fmt.Errorf("failed to execute child process %q: %v", c.Args[0], err)
	}

	err = c.Wait()
	if err != nil {
		return fmt.Errorf("error during child process execution %q: %v", c.Args[0], err)
	}
	return nil
}
