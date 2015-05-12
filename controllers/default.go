package controllers

import (
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	// "log"
	// "net/url"
	"bufio"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/codegangsta/cli"
	"github.com/fairlyblank/md2min"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	// "sort"
	// "strconv"
	"strings"
)

var (
	iniconf       config.ConfigContainer = nil
	G_dirDest                            = ""
	G_dirSrc                             = ""
	G_pdfFileName                        = "yiguo.pdf"
)

func init() {
	initConfig()
	go initCli()
}

func initCli() {
	cliApp := cli.NewApp()
	cliApp.Name = ""
	cliApp.Usage = "设置系统运行参数"
	cliApp.Version = "1.0.1"
	cliApp.Email = "ssor@qq.com"
	cliApp.Commands = []cli.Command{
		{
			Name:        "ShowHtmlFile",
			ShortName:   "show",
			Usage:       "设置需要转换成PDF文件的html文件列表",
			Description: "注意文件排列顺序",
			Action: func(c *cli.Context) {
				dir := G_dirDest + c.Args().First()
				DebugInfoF("html目录为 %s", dir, GetFileLocation())
				list, err := GetHtmlList(dir)
				if err != nil {
					DebugSysF("获取文件列表失败：%s", err.Error(), GetFileLocation())
				} else {
					DebugInfoF("文件列表 (%d)：", len(list), GetFileLocation())
					for _, name := range list {
						DebugTraceF(name, GetFileLocation())
					}
				}
			},
		}, {
			Name:        "toHtml",
			ShortName:   "html",
			Usage:       "将MD文件转换成html文档",
			Description: "生成的文档目录结构与原文件目录一致，生成的文档将会放置到对应的相同目录，图片文件直接复制到对应目录",
			Action: func(c *cli.Context) {
				if err := ConvertMdFilesToHtml(G_dirSrc, G_dirDest); err != nil {
					DebugSysF(err.Error(), GetFileLocation())
				}
			},
		}, {
			Name:        "outputPDF",
			ShortName:   "pdf",
			Usage:       "输出pdf文件",
			Description: "将之前的html列表转换成一个PDF文档输出",
			Action: func(c *cli.Context) {
				// createVersionInfoFile()
				// OutputVersionFile()
				dir := G_dirDest + c.Args().First()

				DebugInfoF("html目录为 %s", dir, GetFileLocation())
				list, err := GetHtmlList(dir)
				if err != nil {
					DebugSysF("获取文件列表失败：%s", err.Error(), GetFileLocation())
				} else {
					DebugInfoF("文件列表 (%d)：", len(list), GetFileLocation())
					for _, name := range list {
						DebugTraceF(name, GetFileLocation())
					}
					OutputPDF(G_pdfFileName, list)
				}
			},
		},
	}
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Println("等待输入。。。")

			data, _, _ := reader.ReadLine()
			command := string(data)
			cliApp.Run(strings.Split(command, " "))
		}
	}()
	// app.Run(os.Args)
}

func initConfig() {
	var err error
	iniconf, err = config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		beego.Error(err.Error())
	} else {
		// //忽略特定文件
		// ignoreFiles := iniconf.Strings("ignoredFiles")

		// temp := []string{}
		// for _, file := range ignoreFiles {
		// 	if len(file) > 0 {
		// 		temp = append(temp, file)
		// 	}
		// }
		// ignoreFiles = temp
		// if len(ignoreFiles) > 0 {
		// 	IgnoreFileNameList = append(IgnoreFileNameList, ignoreFiles...)
		// 	beego.Info(fmt.Sprintf("现有 %d 个忽略的文件", len(IgnoreFileNameList)))
		// 	beego.Info("过滤文件名称如下：")
		// 	for _, keyword := range ignoreFiles {
		// 		beego.Debug(keyword)
		// 	}
		// } else {
		// 	beego.Info("没有需要忽略的文件")
		// }

		dirSrc := iniconf.String("srcDir")
		if len(dirSrc) > 0 {
			G_dirSrc = dirSrc
		}
		DebugInfo("源目录：" + G_dirSrc)

		dirDest := iniconf.String("destDir")
		if len(dirDest) > 0 {
			G_dirDest = dirDest
		}
		DebugInfo("输出目录: " + G_dirDest)
	}
}
func OutputPDF(pdfName string, htmlNameList []string) {
	htmlPathList := []string{}
	for _, name := range htmlNameList {
		htmlPathList = append(htmlPathList, G_dirDest+name)
	}
	htmlPathList = append(htmlPathList, pdfName)

	DebugInfoF("开始转换 ...", GetFileLocation())
	cmd := exec.Command("wkhtmltopdf", htmlPathList...)
	err := cmd.Run()
	if err != nil {
		// log.Fatal(err)
		DebugMust("转换PDF出错：" + err.Error() + GetFileLocation())
		return
	} else {
		DebugInfoF("转换完成", GetFileLocation())
	}
}

func GetHtmlList(dir string) ([]string, error) {
	list := NumberNameList{}

	walkFn := func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matched, _ := regexp.MatchString(`\/\.\w+`, fullPath); matched == true {
			return nil
		}
		if info.IsDir() == false {
			if filepath.Ext(info.Name()) == ".html" {
				list = list.Add(info.Name())
			}
		}
		return nil
	}
	if err := filepath.Walk(dir, walkFn); err != nil {
		return nil, err
	}
	list.Print()
	return list.ToNameList(), nil
}

//将指定目录内的md文件转换为html文件，输出到指定目录，目录结构和md文件的目录结构一致
func ConvertMdFilesToHtml(mdFileDir, outputPath string) error {
	//创建对应的目录
	walkFn := func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matched, _ := regexp.MatchString(`\/\.\w+`, fullPath); matched == true {
			// DebugInfo(fullPath)
			return nil
		}
		if info.IsDir() == true { //在应用目录内查看是否已创建该目录
			dirPath := strings.Replace(fullPath, mdFileDir, outputPath, 1)
			DebugTrace(fmt.Sprintf("需要创建目录：%s", dirPath) + GetFileLocation())
			if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
				DebugMust(fmt.Sprintf("创建目录 %s 失败：%s", dirPath, err.Error()) + GetFileLocation())
				return err
			} else {
				DebugInfo(fmt.Sprintf("创建目录 %s 成功", dirPath) + GetFileLocation())
			}
		}
		return nil
	}
	if err := filepath.Walk(mdFileDir, walkFn); err != nil {
		return err
	}
	//转换文件
	walkFnFile := func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matched, _ := regexp.MatchString(`\/\.\w+`, fullPath); matched == true {
			// DebugInfo(fullPath)
			return nil
		}
		if info.IsDir() == false { //在应用目录内查看是否已创建该目录
			//如果是md文件，将文件转换到应用对应目录内
			if (filepath.Ext(info.Name())) == ".md" {
				destFilePath := strings.Replace(fullPath, mdFileDir, outputPath, 1)
				destFilePath = strings.TrimRight(destFilePath, ".md") + ".html"
				DebugTrace(fmt.Sprintf("将 %s 转换为 %s ", fullPath, destFilePath) + GetFileLocation())
				if errConvert := ConvertMd2Html(fullPath, destFilePath); errConvert != nil {
					return errConvert
				}
			}
			//如果文件为图片，则将文件拷贝到对应目录
			if ext := filepath.Ext(info.Name()); ext == ".png" || ext == ".jpg" {
				destFilePath := strings.Replace(fullPath, mdFileDir, outputPath, 1)
				if errCopy := CopyFile(destFilePath, fullPath); errCopy != nil {
					return errCopy
				}
				DebugTrace(fmt.Sprintf("复制 %s 到 %s 成功", fullPath, destFilePath) + GetFileLocation())
			}
		}
		return nil
	}
	if err := filepath.Walk(mdFileDir, walkFnFile); err != nil {
		return err
	}
	return nil
}

//转换一个md文件为html文件，放到指定目录中
func ConvertMd2Html(mdFileNamePath, outputPath string) error {
	md := md2min.New("none")
	var r io.ReadCloser

	f, err := os.Open(mdFileNamePath)
	if err != nil {
		return err
	}
	DebugTrace(fmt.Sprintf("找到文件 %s ", mdFileNamePath) + GetFileLocation())

	r = f
	defer func() { r.Close() }()
	text, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	DebugTrace(fmt.Sprintf("读取文件 %s 成功", mdFileNamePath) + GetFileLocation())
	// newname := strings.TrimRight(mdFileNamePath, ".md") + ".html"
	// newname = outputPath + newname
	outfile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	DebugTrace(fmt.Sprintf("创建文件 %s 成功", outputPath) + GetFileLocation())
	defer outfile.Close()

	err = md.Parse(text, outfile)
	if err != nil {
		return err
	}
	DebugInfo(fmt.Sprintf("文件 %s 转成 %s 成功", mdFileNamePath, outputPath) + GetFileLocation())
	return nil
}

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	this.Data["Website"] = "beego.me"
	this.Data["Email"] = "astaxie@gmail.com"
	this.TplNames = "index.tpl"
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
func CopyFile(dstName, srcName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
