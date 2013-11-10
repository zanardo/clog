// Copyright (c) 2013, Antonio Zanardo <zanardo@gmail.com>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//  * Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//  * Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR AND CONTRIBUTORS ``AS IS'' AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE AUTHOR AND CONTRIBUTORS BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// This is clog client. See https://github.com/zanardo/clog for more details.
//

package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	Version = "0.3dev"
)

var queuePath string
var scriptsPath string

type RunStat struct {
	ScriptName string
	HostName   string
	UserName   string
	StartTime  int64
	EndTime    int64
	Duration   float64
	Status     string
}

func usage() {
	fmt.Printf(`%s version %s usage:

clog [-queue-path <path> -scripts-path <path>] run <script>
clog [-queue-path <path>] send-queue <server url>

`, os.Args[0], Version)
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.StringVar(&queuePath, "queue-path", DefaultQueuePath(), "queue path")
	flag.StringVar(&scriptsPath, "scripts-path", DefaultScriptsPath(), "scripts path")
	flag.Parse()
	CreatePaths()
	if flag.NArg() == 2 {
		switch flag.Arg(0) {
		case "run":
			runScript(flag.Arg(1))
		case "send-queue":
			runQueue(flag.Arg(1))
		default:
			usage()
		}
	} else {
		usage()
	}
}

func UserHome() (path string) {
	if me, err := user.Current(); err != nil {
		panic(err)
	} else {
		return me.HomeDir
	}
}

func UserName() (username string) {
	if me, err := user.Current(); err != nil {
		panic(err)
	} else {
		return me.Username
	}
}

func HostName() (hostname string) {
	if hostname, err := os.Hostname(); err != nil {
		panic(err)
	} else {
		return hostname
	}
}

func DefaultScriptsPath() (path string) {
	return filepath.Join(UserHome(), ".clog-scripts")
}

func DefaultQueuePath() (path string) {
	return filepath.Join(UserHome(), ".clog-queue")
}

func QueuePath() (path string) {
	return queuePath
}

func ScriptsPath() (path string) {
	return scriptsPath
}

func createPath(path string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Print("creating path " + path)
			if err := os.Mkdir(path, 0700); err != nil {
				panic("error creating directory " + path)
			}
		} else {
			panic("error getting directory information!")
		}
	}
}

func CreatePaths() {
	createPath(ScriptsPath())
	createPath(QueuePath())
}

func DieIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Saves queue metadata for delivery.
func (runStat *RunStat) writeQueueMetadata(queuePath string) {
	fp, err := os.OpenFile(queuePath+".meta.tmp", os.O_CREATE|os.O_WRONLY, 0600)
	DieIfErr(err)
	json, err := json.Marshal(runStat)
	DieIfErr(err)
	fp.Write(json)
	fp.Close()
	err = os.Rename(queuePath+".meta.tmp", queuePath+".meta")
	DieIfErr(err)
}

// Read queue metadata.
func (runStat *RunStat) readQueueMetadata(queuePath string) {
	fp, err := os.Open(queuePath)
	DieIfErr(err)
	defer fp.Close()
	b := make([]byte, 64000)
	n, err := fp.Read(b)
	DieIfErr(err)
	err = json.Unmarshal(b[:n], &runStat)
	DieIfErr(err)
}

func runScript(script string) {

	runStat := new(RunStat)
	runStat.ScriptName = script

	scriptpath := filepath.Join(ScriptsPath(), script)
	log.Print("running script ", scriptpath)

	id := GenId()
	log.Print("queue id: ", id)
	queueLogPath := filepath.Join(QueuePath(), id)
	log.Print("queue path: ", queueLogPath)

	runStat.HostName = HostName()
	runStat.UserName = UserName()

	runStat.StartTime = time.Now().Unix()
	startTime := time.Now().UnixNano()

	cmd := exec.Command(scriptpath)

	// Merging stdout and stderr.
	stdout, err := cmd.StdoutPipe()
	DieIfErr(err)
	cmd.Stderr = cmd.Stdout

	scanner := bufio.NewScanner(stdout)

	err = cmd.Start()
	DieIfErr(err)

	queuelog, err := os.OpenFile(queueLogPath+".out", os.O_CREATE|os.O_WRONLY, 0600)
	DieIfErr(err)

	log.Print("command output starts")
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		fmt.Print(line)
		queuelog.WriteString(line)
	}
	log.Print("command output ends")
	queuelog.Close()

	// Checking command exit status.
	runStat.Status = "ok"
	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			runStat.Status = "fail"
		} else {
			log.Fatal(err)
		}
	}

	endTime := time.Now().UnixNano()
	runStat.EndTime = time.Now().Unix()
	runStat.Duration = float64(endTime-startTime) / 1000000000.0

	log.Print("status: ", runStat.Status)
	log.Print("duration: ", fmt.Sprintf("%0.3fs", runStat.Duration))
	log.Print("user: ", runStat.UserName)
	log.Print("hostname: ", runStat.HostName)

	runStat.writeQueueMetadata(queueLogPath)

}

func runQueue(serverurl string) {
	log.Print("target server: ", serverurl)
	queuePath := QueuePath()
	log.Print("queue path: ", queuePath)
	files, err := ioutil.ReadDir(queuePath)
	DieIfErr(err)
	for _, file := range files {
		name := file.Name()
		if len(name) == 41 && strings.HasSuffix(name, ".meta") {
			queueId := name[:36]
			log.Print("dispatching ", queueId)
			runStat := new(RunStat)
			runStat.readQueueMetadata(path.Join(queuePath, queueId) + ".meta")
			v := url.Values{}
			v.Set("start_time", fmt.Sprintf("%d", runStat.StartTime))
			v.Set("end_time", fmt.Sprintf("%d", runStat.EndTime))
			v.Set("duration", fmt.Sprintf("%0.3f", runStat.Duration))
			v.Set("status", runStat.Status)
			v.Set("script", runStat.ScriptName)
			v.Set("computername", runStat.HostName)
			v.Set("username", runStat.UserName)
			v.Set("id", queueId)
			fp, err := os.Open(path.Join(queuePath, queueId) + ".out")
			DieIfErr(err)
			fi, err := fp.Stat()
			DieIfErr(err)
			if fi.Size() > (1024 * 1024) {
				log.Print(" output too large, truncating")
				_, err := fp.Seek(fi.Size()-(1024*1024), 0)
				DieIfErr(err)
			}
			bout := make([]byte, 1024*1024)
			n, err := fp.Read(bout)
			DieIfErr(err)
			str := base64.StdEncoding.EncodeToString(bout[:n])
			v.Set("output", str)
			r, err := http.PostForm(serverurl, v)
			DieIfErr(err)
			b := make([]byte, 4096)
			n, err = r.Body.Read(b)
			DieIfErr(err)
			if r.StatusCode == 200 && string(b[:n]) == "ok" {
				err = os.Remove(path.Join(queuePath, queueId) + ".out")
				DieIfErr(err)
				err = os.Remove(path.Join(queuePath, queueId) + ".meta")
				DieIfErr(err)
			} else {
				log.Println("error -- server response:")
				log.Print(string(b[:n]))
			}
		}
	}
}

func GenId() (id string) {
	// Poor man's UUID generator
	const validchars = "0123456789abcdef"
	b := make([]byte, 36)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	for i, j := range b {
		b[i] = validchars[j%byte(len(validchars))]
	}
	b[8] = '-'
	b[13] = '-'
	b[18] = '-'
	b[23] = '-'
	b[14] = '4'
	return string(b)
}
