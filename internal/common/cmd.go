package common

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(cmd string) error {
	return RunCommandWithInput(cmd, os.Stdin)
}

func RunCommandWithInput(cmd string, input io.Reader) error {
	parts := strings.Split(cmd, " ")

	c := exec.Command(parts[0], parts[1:]...)
	c.Stdin = input
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	fmt.Println(c.String())

	return c.Run()
}

// runBashCommand runs a bash command, which allows the expansion of globs (*) and other bash features
func RunBashCommand(cmd string) error {
	c := exec.Command("bash", "-c", cmd)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err != nil {
		return fmt.Errorf("error running bash command: %w", err)
	}

	fmt.Println(c.String())

	return nil
}
