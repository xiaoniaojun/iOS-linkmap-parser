package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type LinkMap struct {
	Path string
	Arch string
}

type Library struct {
	Name    string
	Objects []string
	size    uint64
}

// 文件部位，用于控制读取方法
const (
	header             = 1
	objects            = 2
	symbols            = 3
	deadstripedsymbols = 4
	segment            = 5
)

var linkmap LinkMap
var parsePart = header
var libraryList = make(map[string]*Library)
var rowNum2LibMap = make(map[uint]*Library)

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
	PrintLinkMapHeader(linkmap)
	//PrintLibraryList(libraryList)

	PrintLibSize()
}

func PrintLibSize() {
	var totalSize uint64 = 0
	var libSizePair = make(map[uint64]string)
	var sizes []int
	for libName, lib := range libraryList {
		libSizePair[lib.size] = libName
		sizes = append(sizes, int(lib.size))
		totalSize += lib.size
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	fmt.Printf("TotalSize: %dkb\n", totalSize/1024)
	fmt.Printf("%-50sSize\n", "Library Name")
	for _, k := range sizes {
		fmt.Printf("%-50s%dkb\n", libSizePair[uint64(k)], k/1024)
		//println(lib_size_pair[uint64(k)], strconv.Itoa(k/1024)+"kb")
	}

}
func PrintLibraryList(list map[string]*Library) {
	for _, lib := range list {
		println("name:\t", lib.Name)
	}
}

func parse(row string) {
	switch parsePart {
	case header:
		ParseHeader(row)
	case objects:
		if row == "# Sections:" {
			parsePart = segment
			return
		}
		rowNum, valid := parseRowNumber(row)
		libName, objectName := ParseObjectFileRow(row)
		if objectName != "" {
			if libraryList[libName] == nil && valid {
				libraryList[libName] = &Library{libName, []string{objectName}, 0}
			} else {
				libraryList[libName].Objects = append(libraryList[libName].Objects, objectName)
			}
		} else {
			libraryList[libName] = &Library{libName, nil, 0}
		}
		rowNum2LibMap[rowNum] = libraryList[libName]

	case segment:
		if row == "# Symbols:" {
			parsePart = symbols
			return
		}
	case symbols:
		if strings.HasPrefix(row, "# Dead") {
			parsePart = deadstripedsymbols
			break
		}
		if strings.Contains(row, "# Address") {
			break
		}
		ParseSymbolsRow(row)
	case deadstripedsymbols:
		// 弃用或删除的方法、应该不影响编译包大小
		break
	}

}

func parseSymbolsSizeAndRowNum(row string) (retSize uint64, rowNum uint) {
	num, _ := parseRowNumber(row)
	l := strings.Index(row, "	")
	r := strings.Index(row, "[")
	sizeStr := strings.Trim(row[l:r], "	")
	size, err := strconv.ParseInt(sizeStr, 0, 64)
	if err != nil {
		log.Fatal("解析size的时候出错")
	}
	return uint64(size), num
}

func ParseObjectFileRow(row string) (libName string, objectName string) {
	//noinspection GoRedundantSecondIndexInSlices
	libComponent := row[strings.LastIndex(row, "/")+1 : len(row)]

	leftBracketIdx := strings.Index(libComponent, "(")
	rightBracketIdx := strings.LastIndex(libComponent, ")")
	if leftBracketIdx != -1 && rightBracketIdx != -1 { //Bugly(libBugly.a-x86_64-master.o)
		libName = libComponent[0:leftBracketIdx]
		objectName = libComponent[leftBracketIdx+1 : rightBracketIdx]
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
	rowNumber, err := strconv.Atoi(NumStr)
	if err != nil {
		return 0, false
	}
	return uint(rowNumber), true
}

func ParseHeader(row string) {
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
		} else if prefix == "Object files" {
			parsePart = objects
		} else {
			log.Fatal("我们不该见面的")
		}
	}
}

func PrintLinkMapHeader(linkmap LinkMap) {
	fmt.Println("================== LinkMap ==================")
	fmt.Println("Path: " + linkmap.Path)
	fmt.Println("Arch: " + linkmap.Arch)
	fmt.Println("")
}
