package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

//GoWebRestart provide function to detect source code change and automatically restart it
type GoWebRestart struct {
	Option *RestartOption
}

//Watch for change on specific dir
func (g *GoWebRestart) Watch(dir string) {
	watcher, _ := fsnotify.NewWatcher()
	g.recursiveWatch(watcher, dir)
	go g.watchForChange(watcher)
}

func (g *GoWebRestart) watchForChange(watcher *fsnotify.Watcher) {
	referenceTime := time.Now()
	tolerance := g.restartService()
	if g.Option.IsVerbose {
		log.Println(fmt.Sprintf("[I] Setting Tolerance beetween build time to %f seconds(s)", tolerance.Seconds()))
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
					log.Println("[ERROR] " + err.Error())
					break
				}
				defer file.Close()
				if fileInfo, _ := file.Stat(); fileInfo.IsDir() {
					g.recursiveWatch(watcher, event.Name)
				}
				break
			case fsnotify.Write:
				difference := time.Since(referenceTime)
				referenceTime = time.Now()
				if difference.Seconds() < 1+tolerance.Seconds() {
					break
				}
				ext := filepath.Ext(event.Name)
				if g.Option.IsExtExist(ext) {
					g.restartService()
				}

			default:
				break
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				break
			}
			log.Println(err.Error())
		}
	}
}

func (g *GoWebRestart) recursiveWatch(watcher *fsnotify.Watcher, directory string) {
	watcher.Add(directory)
	if g.Option.IsVerbose {
		log.Println("Watching " + directory)
	}
	file, err := os.Open(directory)
	if err != nil {
		log.Println("[FATAL] " + err.Error())
		return
	}
	fileInfos, err := file.Readdir(0)
	file.Close()
	if err != nil {
		log.Println("[FATAL] " + err.Error())
		return
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			g.recursiveWatch(watcher, directory+"/"+fileInfo.Name())
		}
	}
}

func (g *GoWebRestart) restartService() time.Duration {
	referenceTime := time.Now()
	if g.Option.IsVerbose {
		log.Println("[I] Restarting...")
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("[ERROR] " + err.Error())
		return time.Since(referenceTime)
	}

	if _, err := os.Stat(cwd + "/tmp_" + g.Option.ProgramName + g.Option.ProgramExt); err == nil {
		os.Remove(cwd + "/tmp_" + g.Option.ProgramName + g.Option.ProgramExt)
		if g.Option.IsVerbose {
			log.Printf("[I] Cleaning Residue : " + cwd + "/tmp_" + g.Option.ProgramName + g.Option.ProgramExt)
		}
	}

	if err := g.compile(cwd); err != nil {
		log.Println("[ERROR] " + err.Error())
		return time.Since(referenceTime)
	}

	if g.Option.Process != nil {
		g.Option.Process.Kill()
		g.Option.Process.Wait()
		g.Option.Process = nil
	}

	g.swapProcess(cwd)

	if g.Option.IsVerbose {
		log.Println("Finish Restarting")
	}
	return time.Since(referenceTime)
}

func (g *GoWebRestart) compile(cwd string) error {
	paramList := []string{"build", "-o", "tmp_" + g.Option.ProgramName + g.Option.ProgramExt}
	if len(g.Option.PassParam) > 0 {
		paramList = append(paramList, strings.Split(g.Option.PassParam, " ")...)
	}

	if g.Option.IsVerbose {
		log.Println("[I] Building ")
	}

	cmd := exec.Command("go", paramList...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd.Wait()
	if _, err := os.Stat(cwd + "/tmp_" + g.Option.ProgramName + g.Option.ProgramExt); err != nil {
		return err
	}
	return nil
}

func (g *GoWebRestart) swapProcess(cwd string) {
	if _, err := os.Stat(cwd + "/" + g.Option.ProgramName + g.Option.ProgramExt); err == nil {
		if err = os.Remove(cwd + "/" + g.Option.ProgramName + g.Option.ProgramExt); err != nil {
			log.Println("[ERROR] " + err.Error())
		}
		if g.Option.IsVerbose {
			log.Printf("[I] Removing OLD : " + cwd + "/" + g.Option.ProgramName + g.Option.ProgramExt)
		}
	}

	if err := os.Rename(cwd+"/tmp_"+g.Option.ProgramName+g.Option.ProgramExt, cwd+"/"+g.Option.ProgramName+g.Option.ProgramExt); err != nil {
		log.Println("[ERROR] " + err.Error())
	}

	cmd := exec.Command(cwd + "/" + g.Option.ProgramName + g.Option.ProgramExt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	g.Option.Process = cmd.Process
}

//NewGoWebRestart instance with its option added
func NewGoWebRestart() *GoWebRestart {

	webRestart := new(GoWebRestart)
	webRestart.Option = NewRestartOption()

	return webRestart
}
