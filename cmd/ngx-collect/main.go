package main

import (
	"os"

	"github.com/logn/ngx-collect/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	// log.SetReportCaller(true)

	cmd.Execute()
}
