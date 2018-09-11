package addd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

// IsValidIp return an error if the ip address is invalid
func IsValidIp(ipAddr string, v6 bool) error {
	err := fmt.Errorf("Invalid ip address %s", ipAddr)
	ip := net.ParseIP( ipAddr )
	if ip == nil {
		return err
	}
	v4 := ip.To4() != nil
	if v6 == v4 { // v6 xor v4
		return err
	}
	return nil
}