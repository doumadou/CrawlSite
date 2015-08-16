package main

import (
"fmt"
"net/url"
"net/http"
"io/ioutil"
"log"
"regexp"
"time"
"strings"
"strconv"
)

type TorrentInfo struct {
	Title string
	Hashinfo string
	FileType string
	CreateTime time.Time 
	FileSize  int64
	FileCount int
}

var (
	searchResult = regexp.MustCompile(`<div class="s-tle">([\s\S.]*?)<form`)
	resultList = regexp.MustCompile(`<div class="s-item">[.\s\S]*?【(.*?)】<a href="/info-([0-9a-zA-Z]+?).html" title="(.*?)">[.\s\S]*?收录时间：<b>(.*?)</b>[.\s\S]*?文件数量：<b>(.*?)</b>[.\s\S]*?资源大小：<b>(.*?)</b>`)

	sizeMap = map[string]int64 {
		"kb": 1024,
		"mb": 1024 * 1024,
		"gb": 1024 * 1024 * 1024,
	}
)

func stringParseInt(str string) string {
	var i int
	for i = 0; i< len(str); i++ {
		d := str[i]
		if '0' <= d && d <= '9' {
			continue
		}
		break
	}
	if i >= 1 {
		return str[:i]
	}
	return str
}


		//str = "3天前"
		//str = "4分钟前"
		//str = "1小时前"	
func StringConvertDateTime (str string) time.Time{
	baseTime := time.Now()

	if str == "昨天" {
        baseTime = baseTime.Add(24 * time.Hour)
	} else {
		value, _ := strconv.ParseInt(stringParseInt(str), 10, 32)
		fmt.Println(time.Duration(value))
		if strings.HasSuffix(str, "天前") {
			baseTime = baseTime.Add(time.Duration(value) * 24 * time.Hour * -1) 
		} else if strings.HasSuffix(str, "分钟前") {
			baseTime = baseTime.Add(time.Duration(value) * 60 * time.Second * -1) 
		} else if strings.HasSuffix(str, "小时前") {
			baseTime = baseTime.Add(time.Duration(value) * time.Hour * -1) 
		}
	}
    
	return baseTime
}

func StringConvertInt64Size(str string) int64 {
	var i int
	for i = 0; i< len(str); i++ {
		d := str[i]
		if '0' <= d && d <= '9' || d == '.' {
			continue
		}
		break
	}

	value, _ := strconv.ParseFloat(str[:i], 64)

	level := strings.ToLower(str[i:])
	level = strings.Trim(level, " ")

	return int64(value * float64(sizeMap[level]))
}

func main() {
u, _ := url.Parse("http://127.0.0.1:8000/css/")
q := u.Query()
q.Set("username", "user")
q.Set("password", "passwd")
u.RawQuery = q.Encode()
res, err := http.Get(u.String());
if err != nil { 
      log.Fatal(err)
	return 
}
result, err := ioutil.ReadAll(res.Body) 
res.Body.Close() 
if err != nil { 
      log.Fatal(err)
	return 
} 
fmt.Printf("%s", result)


var content string
match := searchResult.FindStringSubmatch(string(result))
    if match != nil {
        content = match[1]
    }
  
fmt.Printf("%s", content) 

	var results []TorrentInfo
	matches := resultList.FindAllStringSubmatch(string(result), 10000)
	for _, v := range matches {
		var info TorrentInfo
		info.Title = v[3]
		info.FileType = v[1]
		info.Hashinfo = v[2]
		info.CreateTime = StringConvertDateTime(v[4])
		info.FileCount, _ = strconv.Atoi(v[5])
		info.FileSize = StringConvertInt64Size(v[6])
		fmt.Println(info)
		//for _, vv := range v[1:] {
		//	fmt.Println(vv)
		//}
		results = append(results, info)
	}  


	var str string
	str = "3天前"

	a := StringConvertDateTime(str)

	fmt.Println(a)

	str = "4分钟前"
	a = StringConvertDateTime(str)
	fmt.Println(a)

	str = "1小时前"	
	a = StringConvertDateTime(str)
	fmt.Println(a)

	str = "1.6 GB"
	str = "568.5 MB"

	aa := StringConvertInt64Size(str)
	fmt.Println(aa)

} 
