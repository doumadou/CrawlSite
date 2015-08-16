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
	searchResult = regexp.MustCompile(`<div class="s-tle">([\s\S.]*?)<form`)
	resultList = regexp.MustCompile(`<div class="s-item">[.\s\S]*?【(.*?)】<a href="/info-([0-9a-zA-Z]+?).html" title="(.*?)">[.\s\S]*?收录时间：<b>(.*?)</b>[.\s\S]*?文件数量：<b>(.*?)</b>[.\s\S]*?资源大小：<b>(.*?)</b>`)
	maxPageReg = regexp.MustCompile(`<a class="end" href="/search-.*?-(\d+)-time.html">最后一页</a>`)

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


func GetHtml(crawlUrl string) string{
	u, _ := url.Parse(crawlUrl)
	q := u.Query()
	//q.Set("username", "user")
	//q.Set("password", "passwd")
	u.RawQuery = q.Encode()
	res, err := http.Get(u.String());
	if err != nil { 
	      log.Fatal(err)
		return ""
	}
	result, err := ioutil.ReadAll(res.Body) 
	res.Body.Close() 
	if err != nil { 
	      log.Fatal(err)
		return ""
	} 
	fmt.Printf("%s", result)

	return string(result)
}

func crawlListPage(keyword string, page int) ([]TorrentInfo, string, int64){

	result := GetHtml("http://127.0.0.1:8000/css/")
	//result := GetHtml("http://www.cilidao.com/search-" + keyword + "-1.html")

	var content string
	var maxPage int64
	maxPage = 0

	match := searchResult.FindStringSubmatch(result)
    if match != nil {
        content = match[1]
    }
  
	fmt.Printf("%s", content) 

	match = maxPageReg.FindStringSubmatch(result)
	if match != nil {
		maxPage, _ = strconv.ParseInt(match[1], 10, 0)
	}

	var results []TorrentInfo
	matches := resultList.FindAllStringSubmatch(result, 10000)
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


	return results, content, maxPage
} 

var (
	metaResult = regexp.MustCompile(`<div class="i-tle">资源名称：([\s\S.]*)</div>[\s\S.]*?文件数量：<b>(\d+)</b>[.\s\S]*?资源大小：<b>(.*?)</b>[.\s\S]*?收录时间：<b>(.*?)</b>[.\s\S]*?HASH值：<b>(.*?)</b>`)
	//metaResult = regexp.MustCompile(`<div class="i-tle">资源名称：([\s\S.]*)</div>[\s\S.]*?文件数量：<b>(\d+)</b>[\s\S.]*?资源大小：<b>([.\s]*?)</b>[\s\S.]*?收录时间：<b>(.*?)</b>[.\s\S]*?HASH值：<b>([a-z0-9A-Z]*)</b>`)
	resultBody = regexp.MustCompile(`<div class="i-go list">([.\S\s]*?)<div class="clear">`)
	resultFileList = regexp.MustCompile(`<span.*?>(.*?)</span>`)
	

	sizeMap = map[string]int64 {
		"b": 1, 
		"kb": 1024,
		"mb": 1024 * 1024,
		"gb": 1024 * 1024 * 1024,
	}
)


func crawlTorrentDetail(infohash string) TorrentInfo{

	var torrentinfo TorrentInfo
	

	//result := GetHtml("http://127.0.0.1:8000/css/detail.html")
	result := GetHtml("http://www.cilidao.com/info-" + infohash + ".html")
	
	match := metaResult.FindStringSubmatch(result)
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
	match = resultBody.FindStringSubmatch(result)
	fmt.Println(len(match))
	if match != nil {
		content = match[1]
	}


	//fmt.Println(content)

	var results []FileInfo
	matches := resultFileList.FindAllStringSubmatch(content, 10000)
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

	return torrentinfo
}
