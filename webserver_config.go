package main

import (
	"flag"
	"fmt"
	"github.com/dmotylev/nutrition"
	"github.com/jelinden/uutispuroReader/service"
)

var env = flag.String("env", "test", "environment")

func init() {
	flag.Parse()
	fmt.Println("loading configuration for environment " + *env)
	err := nutrition.Env("WEBSERVER_").File("config_" + *env + ".cfg").Feed(&service.Conf)

	if err != nil {
		panic("Unable to read properties:\n" + err.Error())
	}
}
