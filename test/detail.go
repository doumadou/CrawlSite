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


type FileInfo struct {
	FilePath string
	FileSize int64
}

type TorrentInfo struct {
	Title string
	Hashinfo string
	FileType string
	CreateTime time.Time 
	FileSize  int64
	FileCount int
	FileList []FileInfo
}

var (
	metaResult = regexp.MustCompile(`<div class="i-tle">资源名称：([\s\S.]*)</div>[\s\S.]*?文件数量：<b>(\d+)</b>[.\s\S]*?资源大小：<b>(.*?)</b>[.\s\S]*?收录时间：<b>(.*?)</b>[.\s\S]*?HASH值：<b>(.*?)</b>`)
	//metaResult = regexp.MustCompile(`<div class="i-tle">资源名称：([\s\S.]*)</div>[\s\S.]*?文件数量：<b>(\d+)</b>[\s\S.]*?资源大小：<b>([.\s]*?)</b>[\s\S.]*?收录时间：<b>(.*?)</b>[.\s\S]*?HASH值：<b>([a-z0-9A-Z]*)</b>`)
	resultBody = regexp.MustCompile(`<div class="i-go list">([.\S\s]*?)<div class="clear">`)
	resultList = regexp.MustCompile(`<span.*?>(.*?)</span>`)
	

	sizeMap = map[string]int64 {
		"b": 1, 
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
u, _ := url.Parse("http://127.0.0.1:8000/css/detail.html")
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


var torrentinfo TorrentInfo

//Viaje a las estrellas
//												
//701
//308 GB
//2015-08-03
//fc884d3f4bb996b53b518b561945ddba83845032
//2
//40



match := metaResult.FindStringSubmatch(string(result))
if match != nil {
	for _, v := range match {
		fmt.Println(v) 
	}

	torrentinfo.Title = strings.Trim(match[1], "	")
	torrentinfo.Title = strings.Trim(torrentinfo.Title, "\n")
	fmt.Println(torrentinfo.Title)
	torrentinfo.FileCount, _ = strconv.Atoi(match[2])
	torrentinfo.FileSize = StringConvertInt64Size(match[3])
	torrentinfo.CreateTime,_ =  time.Parse("2006-01-02", match[4])
	torrentinfo.Hashinfo = match[5]
}
  

	var content string
	match = resultBody.FindStringSubmatch(string(result))
	fmt.Println(len(match))
	if match != nil {
		content = match[1]
	}


	//fmt.Println(content)

	var results []FileInfo
	matches := resultList.FindAllStringSubmatch(content, 10000)
	fmt.Println(len(matches))
	//for _, v := range matches {
	//	var info FileInfo
	//	results = append(results, info)
	//	fmt.Println(v[1])
	//}  

	for i := 0; i < len(matches); i = i + 2 {
		var info FileInfo	
		info.FilePath = matches[i][1]
		info.FileSize = StringConvertInt64Size(matches[i + 1][1])
		results = append(results, info)
	}

	torrentinfo.FileList = results

	fmt.Println(torrentinfo)

	st := "2015-08-12"

	time3, _ := time.Parse("2006-01-02", st)

	fmt.Println(time3)
} 
