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

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("No network interfaces found...")
}