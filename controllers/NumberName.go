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
	"path/filepath"
	// "regexp"
	// "fmt"
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

func (this NumberNameList) SplitByH1() NumberNameGroupList {
	this.Sort()
	list := NumberNameGroupList{}
	for _, nn := range this {
		list = list.AddNumberName(nn)
	}
	return list
}

func (this NumberNameList) Print() {
	DebugTraceF(G_DebugLine)
	DebugTraceF("列表：")
	for _, nn := range this {
		DebugTraceF("Name: %10s		H1: %3d		H2: %3d", nn.Name, nn.H1, nn.H2)
	}
	DebugTraceF(G_DebugLine)
}
func (this NumberNameList) ToNameList() []string {
	// sort.Sort(this)
	this.Sort()
	list := []string{}
	for _, nn := range this {
		list = append(list, nn.Name)
	}
	return list
}
func (this NumberNameList) Sort() {
	sort.Sort(this)
}
func (this NumberNameList) Add(name string) NumberNameList {
	// DebugTraceF(G_DebugLine)
	// DebugTraceF("添加名称到列表：%s", name)
	nameWithoutExt := strings.TrimRight(name, filepath.Ext(name))
	if nameWithoutExt == "README" {
		this = append(this, NewNumberName(name, 0, 0))
	} else {
		// if filepath.Ext(name) == ".html" {
		// 	nameWithoutExt = strings.TrimRight(name, ".html")
		// } else if filepath.Ext(name) == ".md" {

		// }
		indexes := strings.SplitN(nameWithoutExt, ".", 2) //章节序列号的数组，3.13 => ["3","13"]
		if len(indexes) >= 1 {
			h1, errH1 := strconv.Atoi(indexes[0])
			if errH1 != nil {
				return this
			}
			if len(indexes) >= 2 {
				h2, errH2 := strconv.Atoi(indexes[1])
				if errH2 == nil {
					this = append(this, NewNumberName(name, h1, h2))
				}
			} else {
				this = append(this, NewNumberName(name, h1, 0))
			}
		}
	}

	// DebugTraceF("现在有 %d 个元素", len(this))
	// DebugTraceF(G_DebugLine)
	return this
}
