package main

import (
	"log"
	"os"
	"bufio"
	"strings"
	"strconv"
)

type LinkMap struct {
	Path string
	Arch string
}

type Library struct {
	Name string
	Objects []string
	Indexes []uint
}

// 文件部位，用于控制读取方法
const (
	header = 1
	objects = 2
	segments = 3
)

var linkmap LinkMap
var parsePart = header
var libraryList = make(map[string]*Library)

func main() {
	if len(os.Args) != 2 {
		println("Usage: alinkmap path")
		os.Exit(1)
	}

	fb, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0755)
	defer fb.Close()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(fb)

	for scanner.Scan() {
		parse(scanner.Text())
	}

	// TODO：参数控制打印信息
	PrintLinkMap(linkmap)
	PrintLibraryList(libraryList)
	//--------
	// test
	//--------
	println("=== main.o Indexes:")
	for _, v := range libraryList["libWeChatSDK.a"].Objects {
		println(v)
	}
}



func PrintLibraryList(list map[string]*Library) {
	for _, lib := range list {
		println(lib.Name)
	}
}

func parse(row string) {
	switch parsePart {
	case header:
		ScanRow(row)
	case objects:
		if row == "# Symbols:" {
			parsePart = segments
			return
		}
		rowNum, valid := parseRowNumber(row)
		libName, objectName := ParseObjectFileRow(row)
		if objectName != "" {
			if libraryList[libName] == nil && valid {
				libraryList[libName] = &Library{libName, []string{objectName},    []uint {rowNum}}
			} else {
				libraryList[libName].Objects = append(libraryList[libName].Objects, objectName)
				libraryList[libName].Indexes = append(libraryList[libName].Indexes, rowNum)
			}
		} else {
			libraryList[libName] = &Library{libName, nil, []uint {rowNum}}
		}

	case segments:
		if strings.Contains(row, "# Address") {
			break
		}
		ParseSymbolsRow(row)
	}
}

func ParseSymbolsRow(row string) {

}

func ParseObjectFileRow(row string) (libName string, objectName string) {
	//noinspection GoRedundantSecondIndexInSlices
	libComponent := row[strings.LastIndex(row, "/") + 1:len(row)]

	leftBracketIdx := strings.Index(libComponent, "(")
	rightBracketIdx := strings.LastIndex(libComponent, ")")
	if leftBracketIdx != -1 && rightBracketIdx != -1 { 	//Bugly(libBugly.a-x86_64-master.o)
		libName = libComponent[0:leftBracketIdx]
		objectName = libComponent[leftBracketIdx + 1 : rightBracketIdx]
	} else { //BiliPadAppDelegate.o
		libName = libComponent
	}
	return libName, objectName
}

// 取出一行中[]中的数字
func parseRowNumber(row string) (rowNum uint, valid bool) {
	fstLeftBctIdx := strings.Index(row, "[")
	fstRightBctIdx := strings.Index(row, "]")
	if fstLeftBctIdx < 0 || fstRightBctIdx < 0 || fstLeftBctIdx >= fstRightBctIdx {
		return 0, false
	}
	NumStr := strings.Trim(row[fstLeftBctIdx:fstRightBctIdx], "[] ")
	rowNumber, err := strconv.Atoi(NumStr);
	if err != nil {
		return 0, false
	}
	return uint(rowNumber), true
}

func ScanRow(row string) {
	if strings.HasPrefix(row, "#") {
		components := strings.Split(row, ":")
		if len(components) != 2 {
			log.Fatal("Wrong format!")
		}
		prefix := strings.Trim(components[0], "# ")

		if prefix == "Path" {
			linkmap.Path = strings.Trim(components[1], " ")
		} else if prefix == "Arch" {
			linkmap.Arch = strings.Trim(components[1], " ")
		} else if prefix == "Object files"{
			parsePart = objects
		} else {
			log.Fatal("我们不该见面的")
		}
	}
}

func PrintLinkMap(linkmap LinkMap) {
	println("================== LinkMap ==================")
	println("Path: " + linkmap.Path)
	println("Arch: " + linkmap.Arch)
}