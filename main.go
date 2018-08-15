package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage=`coca docker is a simple container runtime implementation`

func main(){

	app:=cli.NewApp()
	app.Name="coca docker"
	app.Usage=usage
	app.Commands=[]cli.Command{
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		logsCommand,
		execCommand,
		stopCommand,
		removeCommand,
	}
	app.Before= func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	err:=app.Run(os.Args)
	if err!=nil{
		log.Fatal(err)
	}
}

