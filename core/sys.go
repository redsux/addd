package addd

import (
	"fmt"
    "os"
    "os/signal"
    "strconv"
    "syscall"
)

var (
	pid_file string
)

func StorePid(pfile string) error {
	pid_file = pfile
	if pid_file != "" {
		file, err := os.OpenFile(pid_file, os.O_RDWR|os.O_CREATE, 0666)
		if err == nil {
			file.Write([]byte(strconv.Itoa(syscall.Getpid())))
			defer file.Close()
		}
		return err
	}
	return fmt.Errorf("Pid file path missing")
}

func DeletePid() {
	if pid_file != "" {
		var err = os.Remove(pid_file)
		if err != nil {
			Log.Warning(err.Error())
		}
	}
}

func WaitSig() {
	sig := make(chan os.Signal)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	s := <- sig
	DeletePid()
	Log.WarningF("Signal (%d) received, stopping", s)
}