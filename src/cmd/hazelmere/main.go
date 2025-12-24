package main

import (
	"fmt"
	"os"

	"github.com/ctfloyd/hazelmere-api/src/internal/cli/backfill"
	"github.com/ctfloyd/hazelmere-api/src/internal/cli/dump"
	"github.com/ctfloyd/hazelmere-api/src/internal/cli/fix"
	"github.com/ctfloyd/hazelmere-api/src/internal/cli/serve"
)

const usage = `hazelmere - Hazelmere API CLI

Usage:
  hazelmere <command> [arguments]

Commands:
  serve                Start the API server
  dump                 Dump all database collections to JSON files
  backfill deltas      Backfill delta records from snapshots
  backfill snapshots   Backfill snapshots from Wise Old Man
  fix snapshot-xp      Fix snapshot experience change values

Options:
  -h, --help           Show this help message
  -c, --config PATH    Path to config file (default: config/dev.json)

Examples:
  hazelmere serve
  hazelmere serve -c config/prod.json
  hazelmere dump
  hazelmere dump ~/backups/hazelmere
  hazelmere backfill deltas
  hazelmere backfill snapshots
  hazelmere fix snapshot-xp
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(0)
	}

	cmd := os.Args[1]

	// Handle help flags
	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		fmt.Print(usage)
		os.Exit(0)
	}

	// Parse global flags and get remaining args
	args := os.Args[2:]
	configPath := "config/dev.json"

	// Extract config flag if present
	var filteredArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-c" || args[i] == "--config" {
			if i+1 < len(args) {
				configPath = args[i+1]
				i++ // Skip next arg
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s requires a path argument\n", args[i])
				os.Exit(1)
			}
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	// Route to subcommand
	var err error
	switch cmd {
	case "serve":
		err = serve.Run(configPath, filteredArgs)

	case "dump":
		err = dump.Run(configPath, filteredArgs)

	case "backfill":
		if len(filteredArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: backfill requires a subcommand (deltas, snapshots)")
			fmt.Fprintln(os.Stderr, "Usage: hazelmere backfill <deltas|snapshots>")
			os.Exit(1)
		}
		subcmd := filteredArgs[0]
		subargs := filteredArgs[1:]
		switch subcmd {
		case "deltas":
			err = backfill.RunDeltas(configPath, subargs)
		case "snapshots":
			err = backfill.RunSnapshots(configPath, subargs)
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown backfill subcommand: %s\n", subcmd)
			fmt.Fprintln(os.Stderr, "Available: deltas, snapshots")
			os.Exit(1)
		}

	case "fix":
		if len(filteredArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: fix requires a subcommand (snapshot-xp)")
			fmt.Fprintln(os.Stderr, "Usage: hazelmere fix <snapshot-xp>")
			os.Exit(1)
		}
		subcmd := filteredArgs[0]
		subargs := filteredArgs[1:]
		switch subcmd {
		case "snapshot-xp":
			err = fix.RunSnapshotXP(configPath, subargs)
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown fix subcommand: %s\n", subcmd)
			fmt.Fprintln(os.Stderr, "Available: snapshot-xp")
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", cmd)
		fmt.Fprintln(os.Stderr, "Run 'hazelmere --help' for usage")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
