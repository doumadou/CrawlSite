package main

import (
	//	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var tmpl, _ = template.New("main").Parse(TMPL_MAIN)

func showCmdListPage(w http.ResponseWriter, req *http.Request) {
	tmpl.Execute(w, _config.Cmds)
}

/*
func showCmdResultInitPage(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	html := strings.Replace(_html, "{id}", id, -1)
	io.WriteString(w, html)
}*/

func writeString(w io.Writer, str string) {
	w.Write([]byte(str))
}

func exec_cmd(id int, w io.Writer) {
	cmdCfg := &_config.Cmds[id]
	if cmdCfg.Running {
		writeString(w, "The script is running, please waitting .......")
		return
	}
	cmdCfg.Running = true
	strCmd := cmdCfg.Script
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		content, err := ioutil.ReadFile(cmdCfg.Script)
		if err != nil {
			writeString(w, err.Error())
			return
		}
		strCmd = cmdCfg.Script + ".tmp" + strconv.Itoa(id) + ".bat"
		WriteStringFile(strCmd, "@echo off \r\n chcp 65001 \r\n"+string(content))
		defer os.Remove(strCmd)
	}
	cmd = exec.Command(strCmd)
	cmd.Env = os.Environ()
	cmd.Stdout = w

	path := GetPath(cmdCfg.Script)
	cmd.Dir = path

	err := cmd.Start()
	if err != nil {
		writeString(w, err.Error())
		return
	}

	cmd.Wait()
	cmdCfg.Running = false
	cmdCfg.LastRunTime = time.Now()
	writeString(w, "\n---------------------\nRUN OVER.......................")
	writeString(w, "\nDownload Url:\n"+cmdCfg.Url)

}

func execAndRefreshCmdResult(w http.ResponseWriter, req *http.Request) {
	id, _ := strconv.Atoi(req.FormValue("id"))
	if id >= len(_config.Cmds) {
		writeString(w, "Invalid Command.")
		return
	}
	exec_cmd(id, w)
}


func searchHandler(w http.ResponseWriter, req *http.Request) {
	pathInfo := strings.Trim(req.URL.Path, "/")
    parts := strings.Split(pathInfo, "/")
	keyword := parts[len(parts) - 1]
	//writeString(w, keyword)
	fmt.Println(keyword)

	//path := r.URL.Path
    //request_type := path[strings.LastIndex(path, "."):]
    //switch request_type {
    //	case ".css":
    //            w.Header().Set("content-type", "text/css")
    //    case ".js":
    //            w.Header().Set("content-type", "text/javascript")
    //    default:
    //}
    w.Header().Set("content-type", "text/html")

    fin, err := os.Open("./templates/" + _config.Template + "/list.html")
    defer fin.Close()
    if err != nil {
            //log.Fatal("static resource:", err)
			fmt.Println("template not exists", err)
    }
    fd, _ := ioutil.ReadAll(fin)
    w.Write(fd)

}

func crawlDataResult(w http.ResponseWriter, r *http.Request) {
    // read form value
    //value := r.FormValue("value")
    //if r.Method == "POST" {
    //    // receive posted data
    //    body, err := ioutil.ReadAll(r.Body)
	//}

}

type Cmd struct {
	Text        string
	Script      string
	Url         string
	Running     bool
	LastRunTime time.Time
}

type Config struct {
	WWWRoot string
	Port    int
	Template string
	Cmds    []Cmd
}

var _html string
var _config Config
var port int

func main() {
	flag.Parse()
	ParseJsonFile(&_config, "config.json")
	port = _config.Port
	_html = strings.Replace(HTML_EXEC, "{port}", strconv.Itoa(port), -1)
	http.HandleFunc("/run", showCmdListPage)
	//http.HandleFunc("/run/cmd", showCmdResultInitPage)
	//http.Handle("/run/cmd", websocket.Handler(execAndRefreshCmdResult))

	http.HandleFunc("/api/crawl/", crawlDataResult)

	http.HandleFunc("/run/cmd", execAndRefreshCmdResult)

	//http.Handle("/", http.FileServer(http.Dir(_config.WWWRoot))) //use fileserver directly

	//http.Handle("/", indexFileHandler) 
	http.HandleFunc("/search/", searchHandler) 
	http.Handle("/font", http.FileServer(http.Dir("templates/" + _config.Template))) 
	http.Handle("/css", http.FileServer(http.Dir("templates/" + _config.Template))) 
	http.Handle("/img", http.FileServer(http.Dir("templates/" + _config.Template))) 
	http.Handle("/js", http.FileServer(http.Dir("templates/" + _config.Template))) 
	http.Handle("/", http.FileServer(http.Dir("./templates/default/"))) 

	fmt.Printf("http://localhost:%d/run\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

const HTML_EXEC = `
<html>
<head>
<script type="text/javascript">
var path;
var ws;
function init() {
   console.log("init");
   if (ws != null) {
     ws.close();
     ws = null;
   }
   var div = document.getElementById("msg");
   var host = window.location.host;
   div.innerText =  "\n" + div.innerText;
   ws = new WebSocket("ws://" + host + "/run/exec?id={id}");
   ws.binaryType ="string";
   ws.onopen = function () {
    //div.innerText = "opened\n" + div.innerText;
	//ws.send("ok");
   };
   ws.onmessage = function (e) {
      div.innerText = div.innerText + e.data + "\n";
   };
   ws.onclose = function (e) {
     // div.innerText = div.innerText + "closed";
   };
   //div.innerText = "init\n" + div.innerText;
};
</script>
<body onLoad="init();"/>
<div id="msg"></div>
</html>
`

const TMPL_MAIN = `
<html>
<head>
</head>
<body>
<table border="0" cellspacing="8">
	<thead><tr><th>Name</th><th></th><th>Last run time</th></tr></thead>
	{{with .}}
	{{range $k, $v := .}}
	<tr>
		<td><a href="/run/cmd?id={{$k}}" target="_blank" onclick="return confirm('Do you really run this script?');">{{$v.Text}}</td>
		<td><a href="{{$v.Url}}">Download</td>
		{{with $v.LastRunTime}}
		<td>{{.}}</td>
		{{end}}
	</tr>
	{{end}}
	{{end}}
</table>
</body>
</html>
`
