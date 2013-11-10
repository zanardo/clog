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
	Duration   float64
	Status     string
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
	runStat.Duration = float64(endTime - startTime)/1000000000.0

	log.Print("status: ", runStat.Status)
	log.Print("duration: ", fmt.Sprintf("%0.3fs", runStat.Duration))
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
		if len(name) == 41 && strings.HasSuffix(name, ".meta") {
			queueId := name[:36]
			log.Print("dispatching ", queueId)
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
