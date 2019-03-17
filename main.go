package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	log.Println("Started Go-Restart")
	option := parseParameter(os.Args[1:])
	prepareSignalHandling(option)
	cwd, _ := os.Getwd()
	webRestart := new(GoWebRestart)
	webRestart.Watch(cwd)

	dummy := make(chan bool)
	<-dummy
}

func parseParameter(param []string) *RestartOption {
	data := NewRestartOption()
	for i := 0; i < len(param); i++ {
		switch param[i] {
		case "-e":
			data.AddExt(parseExtension(param[i+1:]))
			break
		case "-p":
			data.PassParam = param[i+1]
			break
		case "-v":
			data.IsVerbose = true
		default:
			break
		}
	}
	return data
}

func parseExtension(param []string) []string {
	end := 0
	for i := 0; i < len(param); i++ {
		end = i
		if param[i][0] == '-' {
			break
		}
	}
	return param[0 : end+1]
}

func prepareSignalHandling(option *RestartOption) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Println("[I] Exiting...")
			log.Println(sig)
			os.Exit(0)
		}
	}()

}
