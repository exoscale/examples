package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"stathat.com/c/jconfig"
)

var conf *jconfig.Config

func main() {
	conf = jconfig.LoadConfig("yeller.conf")

	var errors []string

	error := checkLoad()
	if error != "" {
		errors = append(errors, error)
	}

	error = checkDiskSpace()
	if error != "" {
		errors = append(errors, error)
	}

	if len(errors) > 0 {
		reportErrors(errors)
	}
}

func pushMessage(message string) {
	client := &http.Client{}
	reader := strings.NewReader("{\"message\": \"" + message + "\"}")
	req, err := http.NewRequest("POST", conf.GetString("webhook"), reader)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	body := string(respBody)

	if !strings.Contains(body, "status\": \"success") {
		log.Fatal(body)
	}
}

func reportErrors(errors []string) {
	for i := 0; i < len(errors); i++ {
		pushMessage(errors[i])
	}
}

func checkDiskSpace() string {
	wd, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	fs := syscall.Statfs_t{}
	err = syscall.Statfs(wd, &fs)

	if err != nil {
		log.Fatal(err)
	}

	var disksize = fs.Blocks * uint64(fs.Bsize)
	var free = fs.Bfree * uint64(fs.Bsize)
	var freepercent = float64(100) / float64(disksize) * float64(free)
	var percentage = conf.GetFloat("diskfree")

	if freepercent <= float64(percentage) {
		return fmt.Sprintf("Diskspace limit reached (free %f%%, limit %f%%)", freepercent, percentage)
	}

	return ""
}

func checkLoad() string {
	var out bytes.Buffer
	cmd := exec.Command("uptime")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	var load = out.String()
	r := regexp.MustCompile("load averages: (?P<now>[0-9]*.[0-9]*) (?P<5min>[0-9]*.[0-9]*) (?P<15min>[0-9]*.[0-9]*)")

	match := r.FindStringSubmatch(load)

	loadavg, err := strconv.ParseFloat(match[2], 64)
	loadmax := conf.GetFloat("load")

	if err != nil {
		log.Fatal(err)
	}

	if loadavg >= loadmax {
		return fmt.Sprintf("Server load too high (current: %f, max %f)", loadavg, loadmax)
	}

	return ""
}
