package main

import (
	"fmt"
	"os"
	"bufio"
	"regexp"
	"strings"
	"io/ioutil"
	"strconv"
)

type CPUStates int
const (
	kName_ CPUStates = iota
	kUser_
	kNice_
	kSystem_
	kIdle_
	kIOwait_
	kIRQ_
	kSoftIRQ_
	kSteal_
	kGuest_
	kGuestNice_
 )

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func kernel() string{
	f, err := os.Open("/proc/version")
	check(err)

	reader := bufio.NewReader(f)
	str, err := reader.ReadString('\n')
	check(err)
	match := regexp.MustCompile(`\d.\d+.\d.+?([^\s]+)`)
	return match.FindString(str)
}
func findOs() string{
	f, err := os.Open("/etc/os-release")
	check(err)

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)

	var os string
	for scanner.Scan() {
		keyValue := strings.Split(scanner.Text(), "=")
		if keyValue[0] == "PRETTY_NAME" {
			reg, err := regexp.Compile("[^a-zA-Z0-9]+")
			check(err)
			os = reg.ReplaceAllString(keyValue[1], "")
		}
	}
	return os
}

func MemInfo() int {
	f, err := os.Open("/proc/meminfo")
	check(err)

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)
	var memtotal, memfree int
	for scanner.Scan() {

		l := strings.Split(scanner.Text(), ":")
		if l[0] == "MemTotal" {

			s := strings.TrimSuffix(l[1], " kB")
			s = strings.TrimSpace(s)
			memtotal, _ = strconv.Atoi(s)
		} else if l[0] == "MemFree" {
			s := strings.TrimSuffix(l[1], " kB")
			s = strings.TrimSpace(s)
			memfree, _ = strconv.Atoi(s)
			break
		}
	}
	return 100 * (memtotal - memfree)/ memtotal
}

func upTime() string {
	path := "/proc/uptime"

	f, err := os.Open(path)
	check(err)

	reader := bufio.NewReader(f)
	line , _, _:= reader.ReadLine()

	list := strings.Split(string(line), " ")
	return list[1]
}

func batteryPerc() int {
	f, err := os.Open("/sys/class/power_supply/BAT0/capacity")
	check(err)

	reader := bufio.NewReader(f)
	l, _, _:= reader.ReadLine()

	p, _ := strconv.Atoi(string(l))
	return p
}


func dirFromPath(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	var dirs []string

	for _, file := range files {
		if file.Mode().IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs
}

func isAllDigit(s string) bool {
	re := regexp.MustCompile(`\D`)

	return !re.MatchString(s)
}

func pids() []string{

	var pids []string

	procDirs := dirFromPath("/proc")
	for _, dir := range procDirs {
		if isAllDigit(dir) {
			pids = append(pids, dir)
		}
	}
	return pids
}

func commandPID(pid string) string {
	procDir := "/proc/"
	pidDir := procDir + pid
	cmdFile := pidDir + "/cmdline"

	f, err := os.Open(cmdFile)
	check(err)

	reader := bufio.NewReader(f)
	l, _, _:= reader.ReadLine()

	return string(l)
}

func ramPID(pid string) int{
	procDir := "/proc/"
	pidDir := procDir + pid
	ramFile := pidDir + "/status"

	f, err := os.Open(ramFile)
	check(err)

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)
	var ram int
	for scanner.Scan() {

		l := strings.Split(scanner.Text(), ":")
		if l[0] == "VmSize" {

			s := strings.TrimSuffix(l[1], " kB")
			s = strings.TrimSpace(s)
			ram, _ = strconv.Atoi(s)
		}
	}
	return ram / 1000
}

func uidPID(pid string) string {

	procDir := "/proc/"
	pidDir := procDir + pid
	ramFile := pidDir + "/status"

	f, err := os.Open(ramFile)
	check(err)

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)
	var uid string
	for scanner.Scan() {

		l := strings.Split(scanner.Text(), ":")
		if l[0] == "Uid" {
			s := strings.Split(l[1], "\t")
			uid = s[1]
		}
	}
	return uid
}

//TODO: find CLK_TCK
func clkTck() int64 {
	return int64(0)
}

//TODO: uptime for each pid
func upTimePID(pid string) int64 {
	return int64(0)
}

func main() {
	fmt.Printf("%v\n", kernel())
	fmt.Printf("%v\n", findOs())
	fmt.Printf("%v\n", upTime())
	fmt.Printf("%v\n", MemInfo())
	fmt.Printf("%v\n", batteryPerc())
	pidsList := pids()
	lastPID := pidsList[len(pidsList) - 1]
	fmt.Printf("%v\n", lastPID)
	fmt.Printf("%v\n", commandPID(lastPID))
	fmt.Printf("%v\n", ramPID(lastPID))
	fmt.Printf("%v\n", uidPID(lastPID))

}
