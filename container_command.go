package main

import (
	"github.com/urfave/cli"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/coca_docker/container"
)

var runCommand=cli.Command{
	Name:"runCommand",
	Usage: `Create a container with namespace and cgroups limit
			coca_docker run -ti [command]`,
	Flags:[]cli.Flag{
		cli.BoolFlag{
			Name:"ti",
			Usage:"tty enabled",
		},
	},
	Action: func(context cli.Context) error {
		if len(context.Args())<1{
			return fmt.Errorf("missing the container command.")
		}
		cmd:=context.Args().Get(0)
		tty:=context.Bool("tti")
		Run(cmd,tty)
		return nil
	},
}

var initCommand=cli.Command{
	Name:"init",
	Usage: `init container`,
	Action: func(context cli.Context) error {
		log.Infof("init container")
		cmd:=context.Args().Get(0)
		log.Infof("initcommand %s", cmd)
		err:=container.RunContainerInit(cmd,nil)
		return err
	},
}