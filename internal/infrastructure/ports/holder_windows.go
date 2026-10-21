package ports

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
)

// Find returns the process listening on a TCP port, from netstat.
func Find(port int) (Holder, bool) {
	out, err := exec.Command("netstat", "-ano", "-p", "TCP").Output()
	if err != nil {
		return Holder{}, false
	}
	suffix := ":" + strconv.Itoa(port)
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 5 && strings.HasSuffix(fields[1], suffix) && fields[3] == "LISTENING" {
			pid, _ := strconv.Atoi(fields[4])
			return Holder{PID: pid, Process: process.Name(pid)}, pid > 0
		}
	}
	return Holder{}, false
}
