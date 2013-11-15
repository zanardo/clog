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
	"regexp"
	"time"
)

const (
	Version       = "0.3"
	MaxOutputSize = 1024 * 1024 // 1MB output tail
)

type RunStat struct {
	ScriptName string
	HostName   string
	UserName   string
	StartTime  int64
	EndTime    int64
	Duration   float64
	Status     string
}

func Usage() {
	fmt.Printf(`%s version %s Usage:

clog [-queue-path <path> -scripts-path <path>] run <script>
clog [-queue-path <path>] send-queue <server url>

`, os.Args[0], Version)
	os.Exit(1)
}

func main() {
	var queuePath string
	var scriptsPath string

	flag.Usage = Usage
	flag.StringVar(&queuePath, "queue-path", DefaultQueuePath(), "queue path")
	flag.StringVar(&scriptsPath, "scripts-path", DefaultScriptsPath(), "scripts path")
	flag.Parse()

	CreatePath(queuePath)
	CreatePath(scriptsPath)

	if flag.NArg() == 2 {
		switch flag.Arg(0) {
		case "run":
			RunScript(flag.Arg(1), queuePath, scriptsPath)
		case "send-queue":
			RunQueue(flag.Arg(1), queuePath)
		default:
			Usage()
		}
	} else {
		Usage()
	}
}

// Get current user home path
func UserHome() (path string) {
	if me, err := user.Current(); err != nil {
		panic(err)
	} else {
		return me.HomeDir
	}
}

// Get current user name
func UserName() (username string) {
	if me, err := user.Current(); err != nil {
		panic(err)
	} else {
		return me.Username
	}
}

// Get current hostname
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

func CreatePath(path string) {
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

// Saves queue metadata for delivery.
func (runStat *RunStat) WriteQueueMetadata(queuePath string) {
	fp, err := os.OpenFile(queuePath+".meta.tmp", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	json, err := json.Marshal(runStat)
	if err != nil {
		panic(err)
	}
	fp.Write(json)
	fp.Close()
	err = os.Rename(queuePath+".meta.tmp", queuePath+".meta")
	if err != nil {
		panic(err)
	}
}

// Read queue metadata.
func (runStat *RunStat) ReadQueueMetadata(queuePath string) error {
	fp, err := os.Open(queuePath)
	if err != nil {
		return err
	}
	defer fp.Close()
	b := make([]byte, 64000)
	n, err := fp.Read(b)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b[:n], &runStat)
	if err != nil {
		return err
	}
	return nil
}

func RunScript(script string, queuePath string, scriptsPath string) {

	runStat := new(RunStat)
	runStat.ScriptName = script

	scriptpath := filepath.Join(scriptsPath, script)
	log.Print("running script ", scriptpath)

	id := GenId()
	log.Print("queue id: ", id)
	queueLogPath := filepath.Join(queuePath, id)
	log.Print("queue path: ", queueLogPath)

	runStat.HostName = HostName()
	runStat.UserName = UserName()

	runStat.StartTime = time.Now().Unix()
	startTime := time.Now().UnixNano()

	cmd := exec.Command(scriptpath)

	// Merging stdout and stderr.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = cmd.Stdout

	scanner := bufio.NewScanner(stdout)

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	queuelog, err := os.OpenFile(queueLogPath+".out",
		os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

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

	runStat.WriteQueueMetadata(queueLogPath)
}

func RunQueue(serverurl string, queuePath string) {
	log.Print("target server: ", serverurl)
	log.Print("queue path: ", queuePath)

	files, err := ioutil.ReadDir(queuePath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		rxname := regexp.MustCompile(
			`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}\.meta$`)
		if !rxname.MatchString(name) {
			continue
		}
		queueId := name[:36]

		log.Print("dispatching ", queueId)
		runStat := new(RunStat)

		err := runStat.ReadQueueMetadata(path.Join(queuePath, queueId) + ".meta")
		if err != nil {
			log.Print(err)
			continue
		}

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
		if err != nil {
			log.Print(err)
			continue
		}
		defer fp.Close()
		fi, err := fp.Stat()
		if err != nil {
			log.Print(err)
			continue
		}

		if fi.Size() > MaxOutputSize {
			log.Print(" output too large, truncating")
			_, err := fp.Seek(fi.Size()-MaxOutputSize, 0)
			if err != nil {
				log.Print(err)
				continue
			}
		}

		bout := make([]byte, MaxOutputSize)
		n, err := fp.Read(bout)
		if n > 0 && err != nil {
			log.Print(err)
			return
		}

		fp.Close()

		str := base64.StdEncoding.EncodeToString(bout[:n])
		v.Set("output", str)

		r, err := http.PostForm(serverurl, v)
		if err != nil {
			log.Print(err)
			continue
		}
		b := make([]byte, 4096)
		n, err = r.Body.Read(b)
		if err != nil {
			log.Print(err)
			return
		}
		if r.StatusCode == 200 && string(b[:n]) == "ok" {
			err = os.Remove(path.Join(queuePath, queueId) + ".out")
			if err != nil {
				log.Print(err)
				return
			}
			err = os.Remove(path.Join(queuePath, queueId) + ".meta")
			if err != nil {
				log.Print(err)
				return
			}
		} else {
			log.Println("error -- server response:")
			log.Print(string(b[:n]))
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
