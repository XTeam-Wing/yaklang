package yaklib

import (
	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/privileged"
	"os"
	"runtime"
)

func lookupHost(i string) []string {
	return netx.LookupAll(i)
}

func lookupIP(i string) []string {
	return netx.LookupAll(i)
}

var SystemExports = map[string]interface{}{
	"IsTCPPortOpen": func(p int) bool {
		return !utils.IsTCPPortAvailable(p)
	},
	"IsUDPPortOpen": func(p int) bool {
		return !utils.IsUDPPortAvailable(p)
	},
	"LookupHost":                lookupHost,
	"LookupIP":                  lookupIP,
	"IsTCPPortAvailable":        utils.IsTCPPortAvailable,
	"IsUDPPortAvailable":        utils.IsUDPPortAvailable,
	"GetRandomAvailableTCPPort": utils.GetRandomAvailableTCPPort,
	"GetRandomAvailableUDPPort": utils.GetRandomAvailableUDPPort,
	"IsRemoteTCPPortOpen": func(host string, p int) bool {
		return utils.IsTCPPortOpen(host, p)
	},

	// 机器唯一的码
	"GetMachineID": utils.GetMachineCode,

	// 继承自 os
	"Remove":       os.Remove,
	"RemoveAll":    os.RemoveAll,
	"Rename":       os.Rename,
	"TempDir":      os.TempDir,
	"Getwd":        os.Getwd,
	"Getpid":       os.Getpid,
	"Getppid":      os.Getppid,
	"Getuid":       os.Getuid,
	"Geteuid":      os.Geteuid,
	"Getgid":       os.Getgid,
	"Getegid":      os.Getegid,
	"Environ":      os.Environ,
	"Hostname":     os.Hostname,
	"Unsetenv":     os.Unsetenv,
	"LookupEnv":    os.LookupEnv,
	"Clearenv":     os.Clearenv,
	"Setenv":       os.Setenv,
	"Getenv":       os.Getenv,
	"Exit":         os.Exit,
	"Args":         os.Args,
	"Stdout":       os.Stdout,
	"Stdin":        os.Stdin,
	"Stderr":       os.Stderr,
	"Executable":   os.Executable,
	"ExpandEnv":    os.ExpandEnv,
	"Pipe":         os.Pipe,
	"Chdir":        os.Chdir,
	"Chmod":        os.Chmod,
	"Chown":        os.Chown,
	"OS":           runtime.GOOS,
	"ARCH":         runtime.GOARCH,
	"IsPrivileged": privileged.GetIsPrivileged(),
	"GetDefaultDNSServers": func() []string {
		return netx.NewDefaultReliableDNSConfig().SpecificDNSServers
	},
	"WaitConnect": utils.WaitConnect,
}
