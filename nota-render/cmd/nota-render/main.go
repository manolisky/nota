package main

import (
	"fmt"
	"os"
	"strings"

	"nota-render/internal/nota"
)

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "nota-render: %v\n", err)
		usage()
		os.Exit(2)
	}

	if args.Version {
		fmt.Println(nota.Version)
		return
	}

	if args.Jobname == "" {
		usage()
		os.Exit(2)
	}

	cfg := nota.Config{
		Jobname:   args.Jobname,
		Jobs:      args.Jobs,
		Force:     args.Force,
		Verbose:   args.Verbose,
		Backend:   nota.NewVerovioCLIBackend("verovio", "mscore"),
		Converter: nota.NewRSVGConverter("rsvg-convert"),
	}

	if err := nota.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "nota-render: %v\n", err)
		os.Exit(1)
	}
}

type cliArgs struct {
	Jobname string
	Jobs    string
	Force   bool
	Verbose bool
	Version bool
}

func parseArgs(raw []string) (cliArgs, error) {
	args := cliArgs{Jobs: "auto"}
	for i := 0; i < len(raw); i++ {
		arg := raw[i]
		switch {
		case arg == "--force":
			args.Force = true
		case arg == "--verbose":
			args.Verbose = true
		case arg == "--version" || arg == "-v":
			args.Version = true
		case arg == "--jobs":
			i++
			if i >= len(raw) {
				return args, fmt.Errorf("--jobs requires a value")
			}
			args.Jobs = raw[i]
		case strings.HasPrefix(arg, "--jobs="):
			args.Jobs = strings.TrimPrefix(arg, "--jobs=")
		case strings.HasPrefix(arg, "-"):
			return args, fmt.Errorf("unknown option %s", arg)
		default:
			if args.Jobname != "" {
				return args, fmt.Errorf("unexpected extra argument %s", arg)
			}
			args.Jobname = arg
		}
	}
	return args, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: nota-render <jobname> [--jobs auto|N] [--force] [--verbose]")
}
