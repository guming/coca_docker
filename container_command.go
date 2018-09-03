package main

import (
	"github.com/urfave/cli"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/coca_docker/container"
	"github.com/coca_docker/cgroup/subsystems"
	"os"
	"github.com/coca_docker/network"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

var runCommand=cli.Command{
	Name:"run",
	Usage: `Create a container with namespace and cgroups limit
			coca_docker run -ti [command]`,
	Flags:[]cli.Flag {
		cli.BoolFlag {
			Name:"ti",
			Usage:"tty enabled",
		},
		cli.BoolFlag {
			Name:"d",
			Usage:"detach container",
		},
		cli.StringFlag {
			Name:"m",
			Usage:"memory limit",
		},
		cli.StringFlag {
			Name:"cpuset",
			Usage:"cpuset limit",
		},
		cli.StringFlag {
			Name:"cpushare",
			Usage:"cpushare limit",
		},
		cli.StringFlag {
			Name:"v",
			Usage:"volume map",
		},
		cli.StringFlag {
			Name:"name",
			Usage:"container name",
		},
		cli.StringSliceFlag {
			Name:"e",
			Usage:"set env",
		},
		cli.StringFlag{
			Name:"p",
			Usage:"port mapping",
		},
		cli.StringFlag{
			Name:"net",
			Usage:"container network",
		},
	},
	Action: func(context *cli.Context) error {

		if len(context.Args())<1{
			return fmt.Errorf("missing the container command.")
		}

		var cmdArray []string
		imageName:=context.Args().Get(0)
		log.Infof("image name is %s",imageName)
		for _, arg := range context.Args().Tail() {
			log.Infof("context arg is %s",arg)
			cmdArray = append(cmdArray, arg)
		}
		tty:=context.Bool("ti")
		detach:=context.Bool("d")
		if tty && detach {
			return fmt.Errorf("tty and detach are both true.")
		}
		config:=&subsystems.ResourceConfig{
			MemoryLimit:context.String("m"),
			Cpushare:context.String("cpushare"),
			Cpuset:context.String("cpuset"),
		}
		volume:=context.String("v")

		containerName:=context.String("name")
		envSlice:=context.StringSlice("e")

		network := context.String("net")
		portmapping := context.StringSlice("p")

		Run(cmdArray,tty,config,volume,detach,containerName,imageName,envSlice,network,portmapping)

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

var commitCommand=cli.Command{
	Name:"commit",
	Usage: `commit a container into image`,
	Action: func(context *cli.Context) error {
		if len(context.Args())<2{
			return fmt.Errorf("missing the container command.")
		}
		containerName:=context.Args().Get(0)
		imageName := context.Args().Get(1)
		commitContainer(imageName,containerName)
		return nil
	},
}

var listCommand=cli.Command{
	Name:"ps",
	Usage:"list containers",
	Action: func(context *cli.Context) error{
		listContainer()
		return nil
	},
}

var logsCommand=cli.Command {
	Name:"logs",
	Usage:"container logs output",
	Action: func(context *cli.Context) error {
		if len(context.Args())<1{
			return fmt.Errorf("missing the container command.")
		}
		containerName:=context.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand=cli.Command {
	Name:"exec",
	Usage:"exec container",
	Action: func(context *cli.Context) error {

		if os.Getenv(ENV_EXEC_PID) != "" {
			log.Infof("pid callback pid %d", os.Getgid())
			return nil
		}

		if len(context.Args())<2{
			return fmt.Errorf("missing the container command.")
		}
		containerName:=context.Args().Get(0)

		var cmdArray []string
		for _, arg := range context.Args().Tail() {
			log.Infof("context arg is %s",arg)
			cmdArray = append(cmdArray, arg)
		}
		execContainer(containerName,cmdArray)

		return nil
	},
}

var stopCommand=cli.Command{
	Name:"stop",
	Usage:"stop the container",
	Action: func(context *cli.Context) error {

		if len(context.Args())<1{
			return fmt.Errorf("missing the container command.")
		}
		containerName:=context.Args().Get(0)
		stopContainer(containerName)
		return nil
	},
}

var removeCommand=cli.Command{
	Name:"rm",
	Usage:"remove the container",
	Action: func(context *cli.Context) error {

		if len(context.Args())<1{
			return fmt.Errorf("missing the container command.")
		}
		containerName:=context.Args().Get(0)
		removeContainer(containerName)

		return nil
	},
}


var networkCommand=cli.Command{
	Name:"network",
	Usage:"container network",
	Subcommands:[]cli.Command{
		{
			Name:"create",
			Usage:"create a container network",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name:"driver",
					Usage:"network driver",
				},
				cli.StringFlag{
					Name:"subnet",
					Usage:"subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {

				if len(context.Args())<1{
					return fmt.Errorf("missing the container command.")
				}
				network.Init()
				dirver:=context.String("driver")
				subnet_cidr:=context.String("subnet")
				networkName:=context.Args().Get(0)
				if err:=network.CreateNetwork(dirver,networkName,subnet_cidr);err!=nil{
					log.Errorf("create network error %v",err)
				}
				return nil
			},
		},
		{
			Name: "list",
			Usage: "list container network",
			Action:func(context *cli.Context) error {
				network.Init()
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:"remove",
			Usage:"remove the container network",
			Action: func(context *cli.Context) error {

				if len(context.Args())<1{
					return fmt.Errorf("missing the container command.")
				}
				network.Init()
				networkName:=context.Args().Get(0)
				if err:=network.DeleteNetwork(networkName);err!=nil{
					log.Errorf("remove network error %v",err)
				}
				return nil
			},
		},
	},

}

