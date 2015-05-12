package controllers

import (
	// "github.com/astaxie/beego"
	// "io"
	// "io/ioutil"
	// "log"
	// "net/url"
	// "bufio"
	// "fmt"
	// "github.com/astaxie/beego/config"
	// "github.com/codegangsta/cli"
	// "github.com/fairlyblank/md2min"
	// "os"
	// "os/exec"
	// "path/filepath"
	// "regexp"
	"sort"
	"strconv"
	"strings"
)

//文件的名称使用数字命名，通过比较数字对文件名进行排序
type NumberName struct {
	Name string
	// Number float64
	H1, H2 int
}

func NewNumberName(name string, h1, h2 int) *NumberName {
	// integer := int64(number)
	// float := number - integer
	return &NumberName{
		Name: name,
		// Number: number,
		H1: h1,
		H2: h2,
	}
}

type NumberNameList []*NumberName

func (this NumberNameList) Len() int {
	return len(this)
}
func (this NumberNameList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
func (this NumberNameList) Less(i, j int) bool {
	if this[i].H1 < this[j].H1 {
		return true
	} else if this[i].H1 == this[j].H1 {
		return this[i].H2 < this[j].H2
	}
	// return this[i].H1 < this[j].H1 && this[i].H2 < this[j].H2
	return false
}

func (this NumberNameList) Print() {
	DebugTraceF(G_DebugLine, GetFileLocation())
	DebugTraceF("列表：", GetFileLocation())
	for _, nn := range this {
		DebugTraceF("Name: %s  H1: %d  H2: %d", nn.Name, nn.H1, nn.H2, GetFileLocation())
	}
	DebugTraceF(G_DebugLine, GetFileLocation())

}
func (this NumberNameList) ToNameList() []string {
	sort.Sort(this)
	list := []string{}
	for _, nn := range this {
		list = append(list, nn.Name)
	}
	return list
}
func (this NumberNameList) Add(name string) NumberNameList {
	DebugTraceF(G_DebugLine, GetFileLocation())
	DebugTraceF("添加名称到列表：%s", name, GetFileLocation())
	if name == "README.html" {
		this = append(this, NewNumberName(name, 0, 0))
	} else {
		indexes := strings.SplitN(strings.TrimRight(name, ".html"), ".", 2) //章节序列号的数组，3.13 => ["3","13"]
		if len(indexes) >= 2 {
			h1, errH1 := strconv.Atoi(indexes[0])
			h2, errH2 := strconv.Atoi(indexes[1])
			if errH1 == nil && errH2 == nil {
				this = append(this, NewNumberName(name, h1, h2))
			}
		}
	}

	DebugTraceF("现在有 %d 个元素", len(this), GetFileLocation())
	DebugTraceF(G_DebugLine, GetFileLocation())
	return this
}
