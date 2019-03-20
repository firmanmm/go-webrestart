package restart

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

//RestartOption provide option for GoWebRestart
type RestartOption struct {
	ext         map[string]bool
	Source      string
	ProgramName string
	ProgramExt  string
	RunTags     []string
	CompileTags []string
	IsVerbose   bool
}

//AddExt in ".ext" format. Example ".go .exe .html"
func (g *RestartOption) AddExt(ext []string) {
	for _, val := range ext {
		g.ext[val] = true
	}
}

//GetExt will return all ext
func (g *RestartOption) GetExt() []string {
	res := make([]string, 1)
	for key := range g.ext {
		res = append(res, key)
	}
	return res
}

//IsExtExist check if you want to check, accept ".ext"
func (g *RestartOption) IsExtExist(ext string) bool {
	_, ok := g.ext[ext]
	return ok
}

func (g *RestartOption) String() string {
	return "Ext : " + strings.Join(g.GetExt(), " ") +
		" Verbose : " + strconv.FormatBool(g.IsVerbose)
}

//NewRestartOption constructor
func NewRestartOption() *RestartOption {
	data := new(RestartOption)
	data.ext = map[string]bool{".go": true}
	data.IsVerbose = false
	data.CompileTags = make([]string, 0)
	cwd, _ := os.Getwd()
	data.Source = cwd
	if runtime.GOOS == "windows" {
		data.ProgramExt += ".exe"
	} else {
		data.ProgramExt += ".bin"
	}
	return data
}
