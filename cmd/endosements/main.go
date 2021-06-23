package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"go.uber.org/zap"

	"github.com/veraison/common"
	"github.com/veraison/endorsement"
)

func runQueryCommand(
	config *common.Config,
	args []string,
	em *endorsement.Manager,
	logger *zap.Logger,
) error {
	var queryParams = make(common.QueryArgs)
	queryFlags := flag.NewFlagSet("run_query", flag.ExitOnError)
	queryFlags.Func("a", "Additional argument in the form KEY=VALUE.", queryParams.AddFromText)

	if err := queryFlags.Parse(args); err != nil {
		return err
	}

	argsRest := queryFlags.Args()
	if len(argsRest) != 1 {
		return fmt.Errorf("unexpected arguments (expected one query name): %v", argsRest)
	}

	queryName := argsRest[0]
	queryResult, err := em.RunQuery(queryName, queryParams)
	if err != nil {
		return err
	}

	text, err := json.Marshal(queryResult)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", text)

	return nil
}

func runAddCommand(
	config *common.Config,
	args []string,
	em *endorsement.Manager,
	logger *zap.Logger,
) error {
	var queryParams = make(common.QueryArgs)
	var update bool

	queryFlags := flag.NewFlagSet("run_query", flag.ExitOnError)
	queryFlags.Func("a", "Additional argument in the form KEY=VALUE.", queryParams.AddFromText)
	queryFlags.BoolVar(&update, "u", false, "Update existing endorsements with new values.")

	if err := queryFlags.Parse(args); err != nil {
		return err
	}

	argsRest := queryFlags.Args()
	if len(argsRest) != 1 {
		return fmt.Errorf("unexpected arguments (expected one query name): %v", argsRest)
	}

	endorsementName := argsRest[0]
	err := em.AddEndorsement(endorsementName, queryParams, update)

	if err == nil {
		var verb string
		if update {
			verb = "updated"
		} else {
			verb = "added"
		}

		fmt.Printf("Successfully %s %s endorsement.", verb, endorsementName)
	}

	return err
}

func runListQueriesCommand(
	config *common.Config,
	args []string,
	em *endorsement.Manager,
	logger *zap.Logger,
) error {
	name := em.GetName()
	queries := em.GetSupportedQueries()

	sort.Strings(queries)

	fmt.Printf("Endorsement Store \"%s\" supports the following queries:\n", name)
	for _, q := range queries {
		fmt.Printf("\t%s\n", q)
	}

	return nil
}

func runCommand(config *common.Config, command string, args []string, logger *zap.Logger) error {
	var err error

	if config.Debug {
		if logger, err = zap.NewDevelopment(); err != nil {
			return err
		}
		defer logger.Sync() //nolint
	}

	em := endorsement.NewManager()
	if err := em.InitializeStore(
		config.PluginLocations, config.EndorsementStoreName, config.EndorsementStoreParams,
	); err != nil {
		return err
	}

	switch command {
	case "list-queries":
		return runListQueriesCommand(config, args, em, logger)
	case "query":
		return runQueryCommand(config, args, em, logger)
	case "add":
		return runAddCommand(config, args, em, logger)
	default:
		return fmt.Errorf("unexpected command: \"%s\"", command)
	}
}

var usageString = `
Usage: %s [-c path/to/config/dir/] [-d] COMMAND [<command arg>...]

Where COMMAND is one of the following:

query
    Execute the query with the specified name and print the result. Parameters
    for the query can be specified with -a flag.

    Accepted arguments:

       [-a NAME=VALUE ...] QUERY_NAME

list
    List all endorsements currently stored in the store. This command takes no
    additional parameters.

Top-level Options:
`

func main() {
	var configPath string
	var debug bool
	var logger *zap.Logger
	var err error

	if logger, err = zap.NewProduction(); err != nil {
		fmt.Printf("ERROR initializing logger: %v", err)
		os.Exit(1)
	}
	defer logger.Sync() //nolint

	flag.StringVar(&configPath, "c", "", "Path to the directory containing the config file.")
	flag.BoolVar(&debug, "d", false, "Enable debug output.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageString, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		logger.Fatal("Command not specified (use -h for usage).")
	}
	command, commandArgs := args[0], args[1:]

	config := common.NewConfig()
	config.AddPath(configPath)
	if err = config.Reload(); err != nil {
		logger.Fatal("Could not load config", zap.Error(err))
	}

	if err = runCommand(config, command, commandArgs, logger); err != nil {
		logger.Fatal("Could not execute command", zap.String("command", command), zap.Error(err))
	}
}
