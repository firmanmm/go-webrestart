package restart

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

type RestartOption struct {
	ext         map[string]bool
	ProgramName string
	ProgramExt  string
	PassParam   string
	IsVerbose   bool
	Process     *os.Process
}

func (g *RestartOption) AddExt(ext []string) {
	for _, val := range ext {
		g.ext[val] = true
	}
}

func (g *RestartOption) GetExt() []string {
	res := make([]string, 1)
	for key := range g.ext {
		res = append(res, key)
	}
	return res
}

func (g *RestartOption) IsExtExist(ext string) bool {
	_, ok := g.ext[ext]
	return ok
}

func (g *RestartOption) String() string {
	return "Ext : " + strings.Join(g.GetExt(), " ") +
		" PassParam : " + g.PassParam +
		" Verbose : " + strconv.FormatBool(g.IsVerbose)
}

func NewRestartOption() *RestartOption {
	data := new(RestartOption)
	data.ext = map[string]bool{".go": true}
	data.IsVerbose = false
	data.PassParam = ""
	data.Process = nil
	if runtime.GOOS == "windows" {
		data.ProgramExt += ".exe"
	} else {
		data.ProgramExt += ".bin"
	}
	return data
}