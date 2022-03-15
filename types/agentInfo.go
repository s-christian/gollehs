package types

type AgentCallback struct {
	UID      string   // effective UID
	GIDs     []string // all group IDs the user belongs to
	Username string   // login name
	Name     string   // display name
	Hostname string   // hostname of system
	Cwd      string   // current directory
	Output   string   // command output
	ExitCode int      // exit code returned from last command
}
