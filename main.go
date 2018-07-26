package main

import "github.com/urfave/cli"

const usage=`coca docker is a simple container runtime implementation`

func main(){

	app:=cli.NewApp()
	app.Name="coca docker"
	app.Usage=usage


}

