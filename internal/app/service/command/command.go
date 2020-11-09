package command

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"strings"
)

type Command struct {
	Exec string
	Args []string
}

func (c *Command) Add2Arg(d string, d2 string) {
	c.Args = append(c.Args, d)

	if d2 != "" {
		c.Args = append(c.Args, d2)
	}
}

func (c *Command) Add2ArgAsSolo(d string, d2 string) {
	c.Add2Arg(fmt.Sprintf("%s=\"%s\"", d, d2), "")
}

func (c *Command) Create() *exec.Cmd {
	logrus.Debugf("Command <<<%s>>>", c.GetCommandString(true))
	return exec.Command(c.Exec, c.Args...)
}

func (c *Command) GetCommandString(clearPassword bool) string {
	cmdAsString := fmt.Sprintf("%s %s", c.Exec, strings.Join(c.Args[:], " "))

	if !clearPassword {
		return cmdAsString
	}

	return regexp.MustCompile(`(?U)-{2}password="(.+)"`).ReplaceAllString(cmdAsString, `--password="<SECRET>"`)
}

func (c *Command) Run(code int) error {
	_, err := c.Create().Output()

	if err != nil {
		logrus.Error(err)
		return fmt.Errorf("<<<%s>>> command failed. CmdCode: %d", c.GetCommandString(true), code)
	}

	return nil
}

func CreateNewCommand(exec string, args []string) *Command {
	return &Command{exec, args}
}
