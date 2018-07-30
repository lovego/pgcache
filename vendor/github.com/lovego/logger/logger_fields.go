package logger

import (
	"net"
	"os"
	"strings"
)

// Set a default field by key and value.
// Don't use "level", "at", "msg", "stack", "duration" they will be overwritten.
func (l *Logger) Set(key string, value interface{}) *Logger {
	// if l.fields == nil {
	// 	l.fields = make(map[string]interface{})
	// }
	l.fields[key] = value
	return l
}

// Set a default hostname field
func (l *Logger) SetMachineName() *Logger {
	hostname, _ := os.Hostname()
	l.fields["machineName"] = hostname
	return l
}

// Set a default ip field
func (l *Logger) SetMachineIP() *Logger {
	addrs, _ := net.InterfaceAddrs()
	slice := []string{}
	for _, addr := range addrs {
		ip := strings.Split(addr.String(), `/`)[0]
		IP := net.ParseIP(ip)
		if mask := IP.DefaultMask(); mask != nil && !IP.IsLoopback() {
			slice = append(slice, ip)
		}
	}
	l.fields["machineIP"] = slice
	return l
}

// Set a default pid field
func (l *Logger) SetPid() *Logger {
	l.fields["pid"] = os.Getpid()
	return l
}
