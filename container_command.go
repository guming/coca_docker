package main

import (
	"github.com/urfave/cli"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/coca_docker/container"
	"github.com/coca_docker/cgroup/subsystems"
)

var runCommand=cli.Command{
	Name:"run",
	Usage: `Create a container with namespace and cgroups limit
			coca_docker run -ti [command]`,
	Flags:[]cli.Flag{
		cli.BoolFlag{
			Name:"ti",
			Usage:"tty enabled",
		},
		cli.StringFlag{
			Name:"m",
			Usage:"memory limit",
		},
		cli.StringFlag{
			Name:"cpuset",
			Usage:"cpuset limit",
		},
		cli.StringFlag{
			Name:"cpushare",
			Usage:"cpushare limit",
		},
	},
	Action: func(context *cli.Context) error {
		if len(context.Args())<1{
			return fmt.Errorf("missing the container command.")
		}
		cmd:=context.Args()
		tty:=context.Bool("ti")
		config:=&subsystems.ResourceConfig{
			MemoryLimit:context.String("m"),
			Cpushare:context.String("cpushare"),
			Cpuset:context.String("cpuset"),
		}
		Run(cmd,tty,config)
		return nil
	},
}

var initCommand=cli.Command{
	Name:"init",
	Usage: `init container`,
	Action: func(context *cli.Context) error {
		log.Infof("init container")
		cmd:=context.Args().Get(0)
		log.Infof("init command %s", cmd)
		err:=container.RunContainerInit()
		return err
	},
}