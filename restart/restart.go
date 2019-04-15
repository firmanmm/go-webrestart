package restart

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

//GoWebRestart provide function to detect source code change and automatically restart it
type GoWebRestart struct {
	process         *os.Process
	watcher         *fsnotify.Watcher
	OnCompileFinish func()
	OnRun           func()
	Option          *RestartOption
}

//Watch for change on specific source, edit Option
func (g *GoWebRestart) Watch() {
	g.Stop()
	watcher, _ := fsnotify.NewWatcher()
	g.watcher = watcher
	g.recursiveWatch(watcher, g.Option.Source)
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
				if difference.Seconds() < 5 {
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
	compilePath := "/tmp_" + g.Option.ProgramName + g.Option.ProgramExt
	referenceTime := time.Now()
	if g.Option.IsVerbose {
		log.Println("[I] Restarting...")
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("[ERROR] " + err.Error())
		return time.Since(referenceTime)
	}

	if _, err := os.Stat(cwd + compilePath); err == nil {
		os.Remove(cwd + compilePath)
		if g.Option.IsVerbose {
			log.Printf("[I] Cleaning Residue : " + cwd + compilePath)
		}
	}

	if err := g.Compile(cwd+compilePath, g.Option.Source); err != nil {
		log.Println("[ERROR] " + err.Error())
		return time.Since(referenceTime)
	}

	if g.process != nil {
		g.process.Kill()
		g.process.Wait()
		g.process = nil
	}

	g.swapProcess(cwd)

	if g.Option.IsVerbose {
		log.Println("Finish Restarting")
	}
	return time.Since(referenceTime)
}

//Compile current app with certain name, and with path to source code
func (g *GoWebRestart) Compile(name, path string) error {
	paramList := []string{"build", "-o", name}
	if len(g.Option.CompileTags) > 0 && g.Option.CompileTags[0] != "" {
		paramList = append(paramList, g.Option.CompileTags...)
	}
	paramList = append(paramList, path)
	cmd := exec.Command("go", paramList...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd.Wait()
	if g.OnCompileFinish != nil {
		g.OnCompileFinish()
	}

	return nil
}

func (g *GoWebRestart) swapProcess(cwd string) {
	appLocation := cwd + "/" + g.Option.ProgramName + g.Option.ProgramExt

	if _, err := os.Stat(appLocation); err == nil {
		if err = os.Remove(appLocation); err != nil {
			log.Println("[ERROR] " + err.Error())
		}
		if g.Option.IsVerbose {
			log.Printf("[I] Removing OLD : " + appLocation)
		}
	}

	if err := os.Rename(cwd+"/tmp_"+g.Option.ProgramName+g.Option.ProgramExt, appLocation); err != nil {
		log.Println("[ERROR] " + err.Error())
	}

	cmd := exec.Command(appLocation, g.Option.RunTags...)
	cmd.Dir = filepath.Dir(appLocation)
	cmd.Stdout = g.Option.Stdout
	cmd.Stderr = g.Option.Stderr

	cmd.Start()
	if g.OnRun != nil {
		g.OnRun()
	}
	g.process = cmd.Process
}

//Stop current watcher
func (g *GoWebRestart) Stop() {
	if g.watcher != nil {
		g.watcher.Close()
	}
	g.watcher = nil
}

//NewGoWebRestart instance with its option added
func NewGoWebRestart() *GoWebRestart {

	webRestart := new(GoWebRestart)
	webRestart.Option = NewRestartOption()

	return webRestart
}
