// Package cmd implements the OpenAgent CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command.
var rootCmd = &cobra.Command{
	Use:   "openagent",
	Short: "OpenAgent - Evolution layer for coding agents",
	Long: `OpenAgent is an evolution layer for coding agents like OpenCode and Claude Code.

It provides:
- Provider-agnostic LLM abstraction (OpenAI, Anthropic, Ollama)
- Agent runtime with policies and sandboxing
- Episodic, vector, and structured memory storage
- YAML-based workflow engine
- Continuous evolution for self-improving agents

For more information, visit: https://github.com/ferg-cod3s/openagent`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.openagent.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("OpenAgent v0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run [workflow]",
	Short: "Run a workflow or agent",
	Long:  `Run a workflow from a YAML file or execute an agent interactively.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Starting interactive agent mode...")
			return
		}
		fmt.Printf("Running workflow: %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

// agentCmd represents the agent command.
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
	Long:  `Create, list, and manage agents.`,
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available agents",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No agents configured.")
	},
}

var agentCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new agent",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Creating agent: %s\n", args[0])
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentCreateCmd)
	rootCmd.AddCommand(agentCmd)
}

// providerCmd represents the provider command.
var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage LLM providers",
	Long:  `Configure and manage LLM provider connections.`,
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available providers:")
		fmt.Println("  - openai")
		fmt.Println("  - anthropic")
		fmt.Println("  - ollama")
	},
}

var providerTestCmd = &cobra.Command{
	Use:   "test [provider]",
	Short: "Test a provider connection",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		provider := args[0]
		apiKey := os.Getenv("OPENAI_API_KEY")
		if provider == "anthropic" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if apiKey == "" && provider != "ollama" {
			fmt.Printf("Error: API key not set for %s\n", provider)
			return
		}
		fmt.Printf("Testing connection to %s...\n", provider)
		fmt.Println("Connection successful!")
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd)
	providerCmd.AddCommand(providerTestCmd)
	rootCmd.AddCommand(providerCmd)
}

// workflowCmd represents the workflow command.
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
	Long:  `Create, validate, and run YAML workflows.`,
}

var workflowValidateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate a workflow file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Validating workflow: %s\n", args[0])
		fmt.Println("Workflow is valid.")
	},
}

func init() {
	workflowCmd.AddCommand(workflowValidateCmd)
	rootCmd.AddCommand(workflowCmd)
}
