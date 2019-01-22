package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	log.Println("Started Gin-Restart")
	option := parseParameter(os.Args[1:])
	prepareSignalHandling(option)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		watcher.Close()
		panic(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		watcher.Close()
		panic(err)
	}

	recursiveWatch(option, watcher, cwd)
	defer watcher.Close()

	dummy := make(chan bool)
	if fileInfo, err := os.Stat(cwd); err == nil {
		option.ProgramName = fileInfo.Name()
	}

	go watchForChange(&sync.Mutex{}, option, watcher)
	<-dummy
}

func watchForChange(mutex *sync.Mutex, option *RestartOption, watcher *fsnotify.Watcher) {
	referenceTime := time.Now()
	tolerance := restartService(option)
	if option.IsVerbose {
		printSuccess(fmt.Sprintf("Setting Tolerance beetween build time to %f seconds(s)", tolerance.Seconds()))
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Create:
				file, err := os.Open(event.Name)
				if err != nil {
					printError(err)
					break
				}
				defer file.Close()
				if fileInfo, _ := file.Stat(); fileInfo.IsDir() {
					recursiveWatch(option, watcher, event.Name)
				}
				break
			case fsnotify.Write:
				difference := time.Since(referenceTime)
				referenceTime = time.Now()
				if option.IsVerbose {
					log.Printf("Operation took %v second(s)", difference.Seconds())
				}
				if difference.Seconds() < 1+tolerance.Seconds() {
					break
				}
				ext := filepath.Ext(event.Name)
				if option.IsExtExist(ext) {
					restartService(option)
				}

			default:
				break
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			printError(err)
		}
	}
}

func recursiveWatch(option *RestartOption, watcher *fsnotify.Watcher, directory string) {
	watcher.Add(directory)
	if option.IsVerbose {
		printSuccess("Watching " + directory)
	}
	file, err := os.Open(directory)
	if err != nil {
		log.Fatalln(err)
	}
	fileInfos, err := file.Readdir(0)
	file.Close()
	if err != nil {
		log.Fatalln(err)
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			recursiveWatch(option, watcher, directory+"/"+fileInfo.Name())
		}
	}
}

func parseParameter(param []string) *RestartOption {
	data := NewGinRestartOption()
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
	return param[1:end]
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

func restartService(option *RestartOption) time.Duration {
	referenceTime := time.Now()
	log.Println("[I] Restarting...")
	cwd, err := os.Getwd()
	if err != nil {
		printError(err)
		return time.Since(referenceTime)
	}

	var cmd *exec.Cmd
	paramList := []string{"build", "-o", "tmp_" + option.ProgramName}
	if len(option.PassParam) > 0 {
		paramList = append(paramList, strings.Split(option.PassParam, " ")...)
	}

	if option.IsVerbose {
		log.Println("[I] Building ")
	}

	cmd = exec.Command("go", paramList...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		printError(err)
		return time.Since(referenceTime)
	}
	if _, err := os.Stat(cwd + "/tmp_" + option.ProgramName); err != nil {
		printError(err)
		return time.Since(referenceTime)
	}
	if option.IsVerbose {
		printSuccess("Build OK")
	}

	if option.Process != nil {
		option.Process.Kill()
		option.Process = nil
	}

	if _, err := os.Stat(cwd + "/" + option.ProgramName); err == nil {
		os.Remove(cwd + "/" + option.ProgramName)
	}

	os.Rename(cwd+"/tmp_"+option.ProgramName, cwd+"/"+option.ProgramName)

	cmd = exec.Command(cwd + "/" + option.ProgramName)
	cmd.Stdout = os.Stdout
	cmd.Start()
	option.Process = cmd.Process

	printSuccess("Finish Restarting")
	return time.Since(referenceTime)
}

func printSuccess(info interface{}) {
	log.Printf("\033[0;32m[S] %v\033[0m", info)
}

func printError(err interface{}) {
	log.Printf("\033[0;31m[E] %v\033[0m", err)
}
