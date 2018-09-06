package addd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hashicorp/go-sockaddr"
)

var (
	pidFile string
)

// StorePid write our PID in a file
func StorePid(pfile string) error {
	pidFile = pfile
	if pidFile != "" {
		file, err := os.OpenFile(pidFile, os.O_RDWR|os.O_CREATE, 0666)
		if err == nil {
			file.Write([]byte(strconv.Itoa(syscall.Getpid())))
			defer file.Close()
		}
		return err
	}
	return fmt.Errorf("Pid file path missing")
}

// DeletePid delete the file storing our PID
func DeletePid() {
	if pidFile != "" {
		var err = os.Remove(pidFile)
		if err != nil {
			Log.Warning(err.Error())
		}
	}
}

// WaitSig wait SIGINT, SIGTERM, SIGKILL and SIGQUIT signals
func WaitSig() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	s := <-sig
	DeletePid()
	Log.WarningF("Signal (%d) received, stopping", s)
}

// ExternalIP deprecated
func ExternalIP() (string, error) {
	return sockaddr.GetInterfaceIP("eth0")
}
