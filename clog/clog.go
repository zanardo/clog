package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"
	"strings"
)

var queuePath string
var scriptsPath string

type RunStat struct {
	ScriptName string
	HostName   string
	UserName   string
	StartTime  int64
	EndTime    int64
	Duration   int64
	Status     bool
}

func main() {
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
		}
	}
}

func UserHome() (path string) {
	me, err := user.Current()
	DieIfErr(err)
	return me.HomeDir
}

func UserName() (username string) {
	me, err := user.Current()
	DieIfErr(err)
	return me.Username
}

func HostName() (hostname string) {
	hostname, err := os.Hostname()
	DieIfErr(err)
	return hostname
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

func runScript(script string) {

	runStat := new(RunStat)
	runStat.ScriptName = script

	scriptpath := filepath.Join(ScriptsPath(), script)
	log.Print("running script ", scriptpath)

	id, err := GenId()
	DieIfErr(err)
	log.Print("queue id: ", id)
	queueLogPath := filepath.Join(QueuePath(), id)
	log.Print("queue path: ", queueLogPath)

	runStat.HostName = HostName()
	runStat.UserName = UserName()
	runStat.StartTime = time.Now().Unix()

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
	runStat.Status = true
	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			runStat.Status = false
		} else {
			log.Fatal(err)
		}
	}

	runStat.EndTime = time.Now().Unix()
	runStat.Duration = (runStat.EndTime - runStat.StartTime)

	log.Print("sucess: ", runStat.Status)
	log.Print("duration: ", runStat.Duration, "s")
	log.Print("user: ", runStat.UserName)
	log.Print("hostname: ", runStat.HostName)

	runStat.writeQueueMetadata(queueLogPath)

}

func runQueue(url string) {
	log.Print("target server: ", url)
	queuePath := QueuePath()
	log.Print("queue path: ", queuePath)
	files, err := ioutil.ReadDir(queuePath)
	DieIfErr(err)
	for _, file := range files {
		name := file.Name()
		if len(name) == 21 && strings.HasSuffix(name, ".meta") {
			queueId := name[:16]
			log.Print("dispatching ", queueId)
		}
	}

}

func GenId() (id string, err error) {
	const validchars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i, j := range b {
		b[i] = validchars[j%byte(len(validchars))]
	}
	return string(b), nil
}
