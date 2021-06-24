package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"

	"github.com/veraison/common"
	"github.com/veraison/policy"
)

func runCommand(config *common.Config, command string, args []string, logger *zap.Logger) error {
	var err error

	if config.Debug {
		if logger, err = zap.NewDevelopment(); err != nil {
			return err
		}
		defer logger.Sync() //nolint

	}

	pm := policy.NewManager()
	if err := pm.InitializeStore(
		config.PluginLocations, config.PolicyStoreName, config.PolicyStoreParams,
	); err != nil {
		return err
	}
	defer pm.Close()

	switch command {
	case "list":
		return runListCommand(config, args, pm, logger)
	case "verify":
		return runVerifyCommand(config, args, pm, logger)
	case "set":
		return runSetCommand(config, args, pm, logger)
	case "get":
		return runGetCommand(config, args, pm, logger)
	case "delete":
		return runDeleteCommand(config, args, pm, logger)
	default:
		return fmt.Errorf("unexpected command: \"%s\"", command)
	}
}

func runListCommand(config *common.Config, args []string, pm *policy.Manager, logger *zap.Logger) error {
	var tenantID int

	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	listFlags.IntVar(&tenantID, "t", 0, "List stored policies only for the specified tenant ID.")

	if err := listFlags.Parse(args); err != nil {
		return err
	}

	argsRest := listFlags.Args()
	if len(argsRest) != 0 {
		return fmt.Errorf("unexpected arguments: %v (see -h for usage)", argsRest)
	}

	policies, err := pm.ListPolicies(tenantID)
	if err != nil {
		return err
	}

	for _, pentry := range policies {
		fmt.Println(pentry.TenantID, pentry.TokenFormatName)
	}

	return nil
}

func runGetCommand(config *common.Config, args []string, pm *policy.Manager, logger *zap.Logger) error {
	var tenantID int
	var outPath string

	setFlags := flag.NewFlagSet("add", flag.ExitOnError)
	setFlags.IntVar(&tenantID, "t", 1, "TenantID to which the policies should be added.")
	setFlags.StringVar(&outPath, "o", "policy.zip", "Output zip path.")

	if err := setFlags.Parse(args); err != nil {
		return err
	}

	argsRest := setFlags.Args()
	if len(argsRest) == 0 {
		return fmt.Errorf("token format(s) not specified (see -h for usage)")
	}

	var formats []common.TokenFormat
	for _, fmtName := range argsRest {
		format, err := common.TokenFormatFromString(fmtName)
		if err != nil {
			return err
		}

		formats = append(formats, format)
	}

	var policies []*common.Policy
	for _, format := range formats {

		policy, err := pm.GetPolicy(tenantID, format)
		if err != nil {
			return fmt.Errorf(
				"could not get policy for tenant %d, format %q: %v",
				tenantID,
				format,
				err,
			)
		}

		policies = append(policies, policy)
	}

	return common.WritePoliciesToPath(policies, outPath)
}

func runSetCommand(config *common.Config, args []string, pm *policy.Manager, logger *zap.Logger) error {
	var tenantID int
	var force bool

	setFlags := flag.NewFlagSet("seet", flag.ExitOnError)
	setFlags.IntVar(&tenantID, "t", 1, "TenantID to which the policies should be added.")
	setFlags.BoolVar(&force, "f", false, "If specified, an existing policy will be overwritten, if it exists.")

	if err := setFlags.Parse(args); err != nil {
		return err
	}

	argsRest := setFlags.Args()
	if len(argsRest) == 0 {
		return fmt.Errorf("policy path not specified (see -h for usage)")
	} else if len(argsRest) > 1 {
		return fmt.Errorf("unexpected arguments: %v (see -h for usage)", argsRest[1:])
	}

	policiesZipPath := argsRest[0]

	policies, err := common.ReadPoliciesFromPath(policiesZipPath)
	if err != nil {
		return fmt.Errorf("problem reading policies: %v", err)
	}

	if len(policies) == 0 {
		return fmt.Errorf("no policies found inside %v", policiesZipPath)
	}

	for _, policy := range policies {

		_, err := pm.GetPolicy(tenantID, policy.TokenFormat)
		if err == nil { // policy exists
			if force {
				err = pm.DeletePolicy(tenantID, policy.TokenFormat)
				if err != nil {
					return fmt.Errorf(
						"could not remove existing policy for tenantID %d and format %q: %v",
						tenantID,
						policy.TokenFormat,
						err,
					)
				}
			} else {
				return fmt.Errorf(
					"policy for tenant %d and format %q already exists; use -f to overwrite",
					tenantID,
					policy.TokenFormat,
				)
			}
		}

		err = pm.PutPolicy(tenantID, policy)
		if err != nil {
			return fmt.Errorf("problem adding policy %q: %v", policy.TokenFormat, err)
		}
	}

	return nil
}

func runDeleteCommand(config *common.Config, args []string, pm *policy.Manager, logger *zap.Logger) error {
	var tenantID int
	var tokenFormat common.TokenFormat

	deleteFlags := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteFlags.IntVar(&tenantID, "t", 1, "TenantID for which the policies should be deleted.")
	if err := deleteFlags.Parse(args); err != nil {
		return err
	}

	argsRest := deleteFlags.Args()
	if len(argsRest) == 0 {
		return fmt.Errorf("policy token format not specified (see -h for usage)")
	} else if len(argsRest) > 1 {
		return fmt.Errorf("unexpected arguments: %v (see -h for usage)", argsRest[1:])
	}

	if err := tokenFormat.FromString(argsRest[0]); err != nil {
		return err
	}

	if err := pm.DeletePolicy(tenantID, tokenFormat); err != nil {
		return fmt.Errorf(
			"could not delete policy for tenant %d, format %q: %v",
			tenantID,
			tokenFormat,
			err,
		)
	}

	return nil
}

func runVerifyCommand(config *common.Config, args []string, pm *policy.Manager, logger *zap.Logger) error {

	var endorsementsPath string
	var endorsements map[string]interface{}
	var evidenceContext common.EvidenceContext

	verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
	verifyFlags.StringVar(&endorsementsPath, "e", "", "Path to a JSON file containing endorsements to be used.")

	if err := verifyFlags.Parse(args); err != nil {
		return err
	}

	argsRest := verifyFlags.Args()
	if len(argsRest) == 0 {
		return fmt.Errorf("evidence context file not specified (see -h for usage)")
	} else if len(argsRest) > 1 {
		return fmt.Errorf("unexpected arguments: %v (see -h for usage)", argsRest[1:])
	}

	evidenceContextPath := argsRest[0]

	if endorsementsPath == "" {
		endorsements = make(map[string]interface{})
	} else {
		data, err := ioutil.ReadFile(endorsementsPath)
		if err != nil {
			return fmt.Errorf("problem reading endorsements: %v", err)
		}

		err = json.Unmarshal(data, &endorsements)
		if err != nil {
			return fmt.Errorf("problem parsing endorsements: %v", err)
		}

	}

	data, err := ioutil.ReadFile(evidenceContextPath)
	if err != nil {
		return fmt.Errorf("problem reading evidence context: %v", err)
	}

	err = json.Unmarshal(data, &evidenceContext)
	if err != nil {
		return fmt.Errorf("problem parsing evidence context: %v", err)
	}

	pe, _, rpcClient, err := common.LoadAndInitializePolicyEngine(
		config.PluginLocations,
		config.PolicyEngineName,
		config.PolicyEngineParams,
	)

	if err != nil {
		return fmt.Errorf("problem loading policy engine: %v", err)
	}
	defer rpcClient.Close()

	policy, err := pm.GetPolicy(evidenceContext.TenantID, evidenceContext.Format)
	if err != nil {
		return fmt.Errorf(
			"problem obtaining policy for tenant %d, format %q: %v",
			evidenceContext.TenantID,
			evidenceContext.Format,
			err,
		)
	}

	err = pe.LoadPolicy(policy.Rules)
	if err != nil {
		return fmt.Errorf("problem loading policy into engine: %v", err)
	}

	var result common.AttestationResult

	err = pe.GetAttetationResult(evidenceContext.Evidence, endorsements, false, &result)
	if err != nil {
		return fmt.Errorf("problem getting attestation result: %v", err)
	}

	resultBytes, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return fmt.Errorf("problem encoding attestation result: %v", err)
	}

	fmt.Println(string(resultBytes))

	return nil
}

var usageString = `
Usage: %s [-c path/to/config/dir/] [-d] COMMAND [<command arg>...]

Where COMMAND is one of the following:

list
    List all policies currently stored in the store. By default, tenant ID and
    token format for all stored policies are listed. Optionally, -t flag may be
    used to specify the tenant ID to which the output should be limited.

get
    Retrieves policies identified by the specified token formats (and,
    optionally, tenant via -t) and outputs them as a zip archive.

        [-t TENANT] [-o OUTFILE] FORMAT [FORMAT ...]

    If tenant ID is not specified, it defaults to 1. If the name of the output
    file is not specified, it defaults to "policy.zip" in the current
    directory. See policy zip format description in the README for the
    structure of the zip archive.

set
    Set policies from a zip file to the store. The following additional
    arguments are supported:

        [-t TENANT] [-f] POLICY_ZIP

    The tenant ID is specified with -t (and defaults to 1, if not specified),
    and POLICY_ZIP is the path to the zip file containing the policies (see
    policy zip format in the README).

    If a policy for the token format of one of the policies in the zip is
    already present in the store, an error will be returned, unless -f is
    specified, in which case, the existing policy will be overwritten.

delete
    Delete a policy identified by the specified token format (and, optionally,
    tenant via -t).

verify
    Run a policy  against the specified endorsements and evidence context. This
    can be used to test a policy to make sure it behaves as expected against a
    known input. The following additional arguments are supported:

	[-e ENDORSEMENTS] EVIDENCE_CONTEXT

    where ENDORSEMENTS is the path to JSON-encoded endorsement values, and
    EVIDENCE_CONTEXT is the path to the file containing JSON-serialized
    EvidenceContext.

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
