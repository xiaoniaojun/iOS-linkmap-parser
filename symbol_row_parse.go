package main

import (
	"strings"
	"strconv"
	"log"
)

var clazzes = make(map[string]*Clazz)
var lastClass string
var unresolvedCount uint64

type Clazz struct {
	Cls string
	Methods []Mth
	Size uint64
}

type Mth struct {
	MethodName string
	Size uint64
}

func ParseSymbolsRow(row string) {
	if len(row) <= 0 {
		return
	}
	components := strings.Split(row, string('\t'))
	record := strings.Split(components[2], "] ")
	rowNum, err := strconv.Atoi(strings.Trim(record[0], "[ "))
	if err != nil {
		log.Print("解析rowNum失败")
		return
	}
	size, err := strconv.ParseInt(components[1], 0, 64)
	if err != nil {
		log.Print("解析size失败")
	}
	lib := rowNum2LibMap[uint(rowNum)]
	lib.size += uint64(size)

	//record[1]
	if len(record[1]) < 3 {
		return
	}
	prefix1 := record[1][:1]
	prefix3 := record[1][:3]
	if prefix1 == "-" || prefix1 == "+" {
		// 常规方法
		cls, method := parseNormalMethod(record[1])
		if clazzes[cls] == nil {
			clazzes[cls] = &Clazz{}
		}
		clazzes[cls].Cls = cls
		clazzes[cls].Methods = append(clazzes[cls].Methods, Mth{prefix1 + method, uint64(size)})
		clazzes[cls].Size += uint64(size)
		lastClass = cls
	} else if prefix3 == "___" {
		// oc内置方法
		if strings.Contains(record[1], "_block_invoke") {
			cls, method := parseImplicitMethod(record[1])
			if cls == "" || method == "" {
				return
			}
			if clazzes[cls] == nil {
				clazzes[cls] = &Clazz{}
			}
			clazzes[cls].Methods = append(clazzes[cls].Methods, Mth{"[b]" + method, uint64(size)})
			clazzes[cls].Size += uint64(size)
		} else {
			clazzes[lastClass].Methods = append(clazzes[lastClass].Methods, Mth{record[1][3:], uint64(size)})
			clazzes[lastClass].Size += uint64(size)
		}
	} else if prefix1 == "_" {
		// 静态方法
		clsName := "static methods"
		clazzes[clsName] = &Clazz{}
		clazzes[clsName].Cls = clsName
		clazzes[clsName].Methods = append(clazzes[clsName].Methods, Mth{"[s]" + record[1][1:], uint64(size)})
		clazzes[clsName].Size += uint64(size)
	}

}

func parseNormalMethod(s string) (cls string, method string) {
	cm := s[2:len(s)-1]
	components := strings.Split(cm, " ")
	if len(components) != 2 {
		log.Fatal("格式解析错误")
	}
	return components[0], components[1]
}

func parseImplicitMethod(s string) (cls string, method string) {
	l := strings.Index(s, "[") + 1
	r := strings.Index(s, "]")
	if l >= r {
		unresolvedCount++
		return
	}
	cm := s[l:r]
	components := strings.Split(cm, " ")
	if len(components) != 2 {
		log.Fatal("格式解析错误")
	}
	return components[0], components[1]
}