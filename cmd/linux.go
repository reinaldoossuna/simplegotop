package main

import (
	"fmt"
	"os"
	"time"
	"bufio"
	"regexp"
	"strings"
	"io/ioutil"
	"strconv"
	"syscall"
	"encoding/binary"
)

const (
	_AT_CLKTCK = 17

	uintSize uint = 32 << (^uint(0) >> 63)
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
	time_seconds, _ := strconv.ParseFloat(list[0], 32)
	hours := int32(time_seconds)/3600
	remain := int32(time_seconds) % 3600
	minutes := remain / 60
	seconds := remain % 60

	return fmt.Sprintf("%v Hours, %v Min, %v Sec", hours, minutes, seconds)
}

func batteryPerc() int {
	f, err := os.Open("/sys/class/power_supply/BAT0/capacity")
	check(err)

	reader := bufio.NewReader(f)
	l, _, _:= reader.ReadLine()

	p, _ := strconv.Atoi(string(l))
	return p
}

func lastUpgrade() string {
	f, err := os.Open("/var/log/pacman.log")
	check(err)

	reader := bufio.NewReader(f)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)

	//full date parse
	match := regexp.MustCompile(`20\d{2}(-|\/)((0[1-9])|(1[0-2]))(-|\/)((0[1-9])|([1-2][0-9])|(3[0-1]))(T|\s)(([0-1][0-9])|(2[0-3])):([0-5][0-9]):([0-5][0-9])`)

	last := time.Date(0, time.January, 1, 0, 0, 0,0, time.UTC)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.Contains(l, "pacman -S -y") {
			reg := match.FindString(l)

			// its should be a better way to get the localtime
			reg += "-04:00"

			date, _ := time.Parse(time.RFC3339, reg)
			if last.Before(date) {
				last = date
			}
		}
	}

	diff := time.Now().Sub(last)

	return fmt.Sprintf("%v Hours, %v Min", int(diff.Hours()), int(diff.Minutes()) % 60)
}

func freeSpace(path string) uint64 {
	var stat syscall.Statfs_t
	syscall.Statfs(path, &stat)
	return stat.Bavail * uint64(stat.Bsize) / (1024 * 1024 * 1024)
}

func rootFreeSpace() uint64 {
	return freeSpace("/")
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

//find CLK_TCK
//code from cpu_linux.go in shirou/gopsutil
func clkTck() int64 {
	buf, err := ioutil.ReadFile("/proc/self/auxv")
	check(err)

	pb := int(uintSize / 8)
	for i := 0; i < len(buf) - pb*2; i+= pb * 2{
		tag := uint(binary.LittleEndian.Uint64(buf[i:]))
		val := uint(binary.LittleEndian.Uint64(buf[i+pb:]))

		if tag == _AT_CLKTCK {
			return int64(val)
		}
	}
	return int64(0)
}

//TODO: uptime for each pid
func upTimePID(pid string) int64 {
	return int64(0)
}

func main() {
	fmt.Printf("Kernel: %v\n", kernel())
	fmt.Printf("OS: %v\n", findOs())
	fmt.Printf("Up time: %v\n\n", upTime())

	fmt.Printf("Mem: %v %%\n", MemInfo())
	fmt.Printf("Battery: %v %%\n", batteryPerc())
	fmt.Printf("last upgrade: %v\n\n", lastUpgrade())

	/*
	pidsList := pids()
	lastPID := pidsList[len(pidsList) - 1]
	fmt.Printf("Last PID: %v\n", lastPID)
	fmt.Printf("Command: %v\n", commandPID(lastPID))
	fmt.Printf("Ram: %v\n", ramPID(lastPID))
	fmt.Printf("User: %v\n", uidPID(lastPID))
	fmt.Printf("clk tck: %v\n", clkTck())
	fmt.Printf("up time task: %v\n", upTimePID(lastPID))
	*/

	fmt.Printf("Root Free Space: %v Gb\n", rootFreeSpace())
	fmt.Printf("Hdd Free Space: %v Gb\n ", freeSpace("/mnt/hdd"))
}
