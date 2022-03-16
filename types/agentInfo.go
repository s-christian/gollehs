package types

import (
	"fmt"
	"os"
	"strconv"

	//"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
)

type AgentConfig struct {
	UUID string // agent UUID, identifier
}

type AgentCallback struct {
	UID             string // user UID
	GID             string // all group IDs the user belongs to
	Username        string // login name
	Name            string // display name
	Hostname        string // hostname of system
	Cwd             string // current directory
	Output          string // command output
	ExitCode        int    // exit code returned from last command
	LastInteraction string
}

func (a AgentCallback) PrintTable() {
	data := []string{
		a.UID,
		a.GID,
		a.Username,
		a.Name,
		a.Hostname,
		a.Cwd,
		fmt.Sprintf("%d bytes", len(a.Output)),
		strconv.Itoa(a.ExitCode),
		a.LastInteraction,
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"UID", "GID", "Username", "Name", "Hostname", "Cwd", "Output", "Exit Code", "Last Interaction",
	})
	table.SetBorder(false)

	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgBlueColor},
	)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgHiBlackColor, tablewriter.Italic},
		tablewriter.Colors{tablewriter.FgHiBlackColor, tablewriter.Italic},
		tablewriter.Colors{tablewriter.FgGreenColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Italic},
	)

	table.Append(data)
	table.Render()
}
