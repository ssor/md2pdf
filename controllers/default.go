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
	"strconv"
	"strings"
)

var (
	iniconf       config.ConfigContainer = nil
	G_dirDest                            = ""
	G_dirSrc                             = ""
	G_pdfFileName                        = "yiguo.pdf"
	htmlExt                              = ".html"
	mdExt                                = ".md"
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
				list, err := GetFileNameList(dir, htmlExt)
				if err != nil {
					DebugSysF("获取文件列表失败：%s", err.Error(), GetFileLocation())
				} else {
					DebugInfoF("文件列表 (%d)：", len(list), GetFileLocation())
					// for _, name := range list {
					// 	DebugTraceF(name, GetFileLocation())
					// }
					list.Print()
				}
			},
		}, {
			Name:        "toHtml",
			ShortName:   "html",
			Usage:       "将MD文件转换成html文档",
			Description: "生成的文档目录结构与原文件目录一致，生成的文档将会放置到对应的相同目录，图片文件直接复制到对应目录",
			Action: func(c *cli.Context) {
				DebugInfoF("MD目录为 %s", G_dirSrc, GetFileLocation())
				list, err := GetFileNameList(G_dirSrc, mdExt)
				if err != nil {
					DebugSysF("获取文件列表失败：%s", err.Error(), GetFileLocation())
				} else {
					DebugInfoF("文件列表 (%d)：", len(list), GetFileLocation())
					// for _, name := range list {
					// 	DebugTraceF(name, GetFileLocation())
					// }
					list.Print()
				}
				grouplist := list.SplitByH1()
				grouplist.Print()
				if err := PrepareHtmlFileDir(G_dirSrc, G_dirDest); err != nil {
					DebugSysF(err.Error())
					return
				}
				if err := ConvertMdFilesToHtml(grouplist, G_dirSrc, G_dirDest); err != nil {
					DebugSysF(err.Error(), GetFileLocation())
				}
			},
		}, {
			Name:        "outputPDF",
			ShortName:   "pdf",
			Usage:       "输出pdf文件",
			Description: "将之前的html列表转换成一个PDF文档输出",
			Action: func(c *cli.Context) {
				DebugInfoF("html目录为 %s", G_dirDest, GetFileLocation())
				list, err := GetFileNameList(G_dirDest, htmlExt)
				if err != nil {
					DebugSysF("获取文件列表失败：%s", err.Error(), GetFileLocation())
				} else {
					DebugInfoF("文件列表 (%d)：", len(list), GetFileLocation())
					// for _, name := range list {
					// 	DebugTraceF(name, GetFileLocation())
					// }
					list.Print()
					OutputPDF(G_pdfFileName, G_dirDest, list.ToNameList())
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
func OutputPDF(pdfName, htmlDirPath string, htmlNameList []string) {
	htmlPathList := []string{}
	for _, name := range htmlNameList {
		htmlPathList = append(htmlPathList, htmlDirPath+name)
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

func GetFileNameList(dir, ext string) (NumberNameList, error) {
	list := NumberNameList{}

	walkFn := func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if matched, _ := regexp.MatchString(`\/\.\w+`, fullPath); matched == true {
			return nil
		}
		if info.IsDir() == false {
			// DebugTrace(info.Name())
			if filepath.Ext(info.Name()) == ext {
				list = list.Add(info.Name())
			}
		}
		return nil
	}
	if err := filepath.Walk(dir, walkFn); err != nil {
		return nil, err
	}
	// list.Print()
	// return list.ToNameList(), nil
	return list, nil
}

//创建输出html文件的和md文件对应的目录，包括非转换文件（图片）的拷贝
func PrepareHtmlFileDir(mdFileDir, outputPath string) error {
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
		} else {
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
	if err := filepath.Walk(mdFileDir, walkFn); err != nil {
		return err
	}
	return nil
}

//将指定目录内的md文件转换为html文件，输出到指定目录，目录结构和md文件的目录结构一致
func ConvertMdFilesToHtml(nameGroupList NumberNameGroupList, mdSrcPath, outputPath string) error {
	for _, ng := range nameGroupList {
		list := []string{}
		for _, name := range ng.Names {
			list = append(list, mdSrcPath+name)
		}
		if err := ConvertMd2Html(list, outputPath+strconv.Itoa(ng.H1)+htmlExt); err != nil {
			DebugSysF("转换MD文件列表为html时出错：%s ", err.Error())
			return err
		}
	}
	// walkFnFile := func(fullPath string, info os.FileInfo, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if matched, _ := regexp.MatchString(`\/\.\w+`, fullPath); matched == true {
	// 		// DebugInfo(fullPath)
	// 		return nil
	// 	}
	// 	if info.IsDir() == false { //在应用目录内查看是否已创建该目录
	// 		//如果是md文件，将文件转换到应用对应目录内
	// 		if (filepath.Ext(info.Name())) == ".md" {
	// 			destFilePath := strings.Replace(fullPath, mdFileDir, outputPath, 1)
	// 			destFilePath = strings.TrimRight(destFilePath, ".md") + ".html"
	// 			DebugTrace(fmt.Sprintf("将 %s 转换为 %s ", fullPath, destFilePath) + GetFileLocation())
	// 			if errConvert := ConvertMd2Html([]string{}, destFilePath); errConvert != nil {
	// 				return errConvert
	// 			}
	// 		}
	// 	}
	// 	return nil
	// }
	// if err := filepath.Walk(mdFileDir, walkFnFile); err != nil {
	// 	return err
	// }
	return nil
}

//转换一个md文件列表为html文件，放到指定目录中，即多个md文件合为一个html文件
func ConvertMd2Html(mdFileFullPathlist []string, outputFileFullPath string) error {

	// DebugTrace(fmt.Sprintf("转换文件列表 %v 到文件 %s", mdFileFullPathlist, outputFileFullPath))
	// return nil
	mdContent := []byte{}

	// mdFileNamePath := ""

	mdFileReader := func(mdFileNamePath string) ([]byte, error) {
		var r io.ReadCloser
		f, err := os.Open(mdFileNamePath)
		if err != nil {
			return nil, err
		}
		DebugTrace(fmt.Sprintf("找到文件 %s ", mdFileNamePath) + GetFileLocation())

		r = f
		defer func() { r.Close() }()
		text, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		DebugTrace(fmt.Sprintf("读取文件 %s 成功", mdFileNamePath) + GetFileLocation())
		return text, nil
	}
	for _, mdFile := range mdFileFullPathlist {
		content, err := mdFileReader(mdFile)
		if err != nil {
			return err
		}
		if errTryParse := md2min.New("none").TryParse(content); errTryParse != nil {
			DebugSysF("解析文件 %s 时出错：%s", mdFile, errTryParse.Error())
			return errTryParse
		}
		strContent := string(content)
		if linksIndex := strings.Index(strContent, "links"); linksIndex > 0 {
			content = []byte(strContent[:linksIndex-2])
		}
		mdContent = append(mdContent, content...)
		mdContent = append(mdContent, []byte(strings.Repeat("\r\n", 10))...)
	}
	// newname := strings.TrimRight(mdFileNamePath, ".md") + ".html"
	// newname = outputFileFullPath + newname
	outfile, err := os.Create(outputFileFullPath)
	if err != nil {
		return err
	}
	DebugTrace(fmt.Sprintf("创建文件 %s 成功", outputFileFullPath) + GetFileLocation())
	defer outfile.Close()

	md := md2min.New("none")
	err = md.Parse(mdContent, outfile)
	if err != nil {
		DebugSysF("转换文件时出错：原因：%s", err.Error())
		DebugSys(fmt.Sprintf("文件列表：%v", mdFileFullPathlist) + GetFileLocation())
		return err
	}
	// DebugInfo(fmt.Sprintf("文件 %s 转成 %s 成功", mdFileNamePath, outputPath) + GetFileLocation())
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
