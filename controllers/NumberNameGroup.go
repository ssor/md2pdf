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
	"fmt"
	// "sort"
	// "strconv"
	// "strings"
)

type NumberNameGroup struct {
	H1    int
	Names []string
}

func NewNumberNameGroup(h1 int, name string) *NumberNameGroup {
	return &NumberNameGroup{
		H1:    h1,
		Names: []string{name},
	}
}
func (this *NumberNameGroup) Print() {
	DebugTrace(fmt.Sprintf("标题: %3d    文件列表: %v", this.H1, this.Names))
}
func (this *NumberNameGroup) AddName(name string) {
	for _, s := range this.Names {
		if s == name {
			return
		}
	}
	this.Names = append(this.Names, name)
}

type NumberNameGroupList []*NumberNameGroup

func (this NumberNameGroupList) Print() {
	DebugTrace(G_DebugLine)
	for _, nng := range this {
		nng.Print()
	}
	DebugTrace(G_DebugLine)
}

func (this NumberNameGroupList) Find(h1 int) *NumberNameGroup {
	for _, nng := range this {
		if nng.H1 == h1 {
			return nng
		}
	}
	return nil
}
func (this NumberNameGroupList) AddNumberName(nn *NumberName) NumberNameGroupList {
	nng := this.Find(nn.H1)
	if nng == nil {
		this = append(this, NewNumberNameGroup(nn.H1, nn.Name))
	} else {
		nng.AddName(nn.Name)
	}
	return this
}
