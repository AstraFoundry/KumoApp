//go:build !windows

package ports

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Find returns the process listening on a TCP port, when the system says.
func Find(port int) (Holder, bool) {
	if h, ok := lsof(port); ok {
		return h, true
	}
	if runtime.GOOS == "linux" {
		return procNet(port)
	}
	return Holder{}, false
}

func lsof(port int) (Holder, bool) {
	out, err := exec.Command("lsof", "-nP", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-Fpc").Output()
	if err != nil {
		return Holder{}, false
	}
	var h Holder
	for line := range strings.SplitSeq(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "p") && h.PID == 0:
			h.PID, _ = strconv.Atoi(line[1:])
		case strings.HasPrefix(line, "c") && h.Process == "":
			h.Process = line[1:]
		}
	}
	return h, h.PID > 0
}

// procNet maps the port's listening socket to its inode, then the inode to
// the process holding it, without lsof.
func procNet(port int) (Holder, bool) {
	inodes := map[string]bool{}
	for _, file := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 10 || fields[3] != "0A" { // 0A is LISTEN
				continue
			}
			_, hexPort, ok := strings.Cut(fields[1], ":")
			if p, err := strconv.ParseInt(hexPort, 16, 32); ok && err == nil && int(p) == port {
				inodes[fields[9]] = true
			}
		}
		f.Close()
	}
	if len(inodes) == 0 {
		return Holder{}, false
	}
	procs, _ := os.ReadDir("/proc")
	for _, proc := range procs {
		pid, err := strconv.Atoi(proc.Name())
		if err != nil {
			continue
		}
		fds, _ := os.ReadDir(fmt.Sprintf("/proc/%d/fd", pid))
		for _, fd := range fds {
			link, err := os.Readlink(fmt.Sprintf("/proc/%d/fd/%s", pid, fd.Name()))
			if err == nil && strings.HasPrefix(link, "socket:[") && inodes[strings.Trim(link[7:], "[]")] {
				name := ""
				if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
					name = filepath.Base(exe)
				}
				return Holder{PID: pid, Process: name}, true
			}
		}
	}
	return Holder{}, false
}
