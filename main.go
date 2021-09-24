package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/godbus/dbus/v5"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <executable> [<executable args>]\n", os.Args[0])
		os.Exit(0)
	}

	dbusConnection, err := dbus.ConnectSessionBus()
	if err != nil {
		printToStdErr("failed to connect to D-Bus", err)
		os.Exit(1)
	}

	inhibitCookie, err := inhibitSuspend(dbusConnection, os.Args[1])
	if err != nil {
		printToStdErr("failed to inhibit power management", err)
		os.Exit(1)
	}

	setupSignalHandlers(dbusConnection, inhibitCookie)

	execCommand := createExecuteCommand(os.Args[1:])
	err = execCommand.Start()
	if err != nil {
		printToStdErr("failed to start command", err)
		uninhibitSuspendAndExit(dbusConnection, inhibitCookie, 1)
	}

	err = execCommand.Wait()
	if err != nil {
		printToStdErr("command failed", err)
		uninhibitSuspendAndExit(dbusConnection, inhibitCookie, execCommand.ProcessState.ExitCode())
	}

	uninhibitSuspendAndExit(dbusConnection, inhibitCookie, 0)
}

func inhibitSuspend(conn *dbus.Conn, appName string) (uint, error) {
	var inhibitCookie uint
	err := inhibitPowerManagementObject(conn).
		Call("org.freedesktop.PowerManagement.Inhibit.Inhibit", 0, appName, fmt.Sprintf("inhibiting power management while %s is running", appName)).
		Store(&inhibitCookie)

	return inhibitCookie, err
}

func uninhibitSuspend(conn *dbus.Conn, inhibitCookie uint) error {
	return inhibitPowerManagementObject(conn).
		Call("org.freedesktop.PowerManagement.Inhibit.UnInhibit", 0, inhibitCookie).Err
}

func uninhibitSuspendAndExit(conn *dbus.Conn, inhibitCookie uint, exitCode int) {
	if err := uninhibitSuspend(conn, inhibitCookie); err != nil {
		printToStdErr("failed to uninhibit power management", err)
	}
	conn.Close()
	os.Exit(exitCode)
}

func createExecuteCommand(args []string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func printToStdErr(message string, err error) {
	fmt.Fprintf(os.Stderr, message+": %v\n", err)
}

func inhibitPowerManagementObject(conn *dbus.Conn) dbus.BusObject {
	return conn.Object("org.freedesktop.PowerManagement", "/org/freedesktop/PowerManagement/Inhibit")
}

func setupSignalHandlers(conn *dbus.Conn, inhibitCookie uint) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		uninhibitSuspend(conn, inhibitCookie)
		conn.Close()
	}()
}
