package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/danube-messaging/loadtest_danube/pkg/config"
	"github.com/danube-messaging/loadtest_danube/pkg/runner"
	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "loadtest",
	Short: "Load testing tool for Danube",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a test scenario from a config file",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")
		duration, _ := cmd.Flags().GetString("duration")
		_ = duration // reserved for override in later phases
		if cfgPath == "" {
			log.Fatal("--config is required")
		}

		cfg, err := config.LoadFile(cfgPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		if errs := config.Validate(cfg); len(errs) > 0 {
			for _, e := range errs {
				log.Printf("config error: %v", e)
			}
			log.Fatal("configuration is invalid")
		}

		fmt.Printf("Loaded test '%s' targeting %s\n", cfg.TestName, cfg.Danube.ServiceURL)
		fmt.Printf("Topics: %d, Producers: %d groups, Consumers: %d groups\n",
			len(cfg.Topics), len(cfg.Producers), len(cfg.Consumers))
		if err := runner.Run(cfg); err != nil {
			log.Fatalf("run failed: %v", err)
		}
	},
}

func init() {
	runCmd.Flags().String("config", "", "Path to YAML config file")
	runCmd.Flags().String("duration", "", "Override test duration (e.g. 2m)")
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a config file",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, _ := cmd.Flags().GetString("config")
		if cfgPath == "" {
			log.Fatal("--config is required")
		}
		cfg, err := config.LoadFile(cfgPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		if errs := config.Validate(cfg); len(errs) > 0 {
			fmt.Println("Config validation failed:")
			for _, e := range errs {
				fmt.Printf(" - %v\n", e)
			}
			os.Exit(1)
		}
		fmt.Println("Config is valid ")
	},
}

func init() {
	validateCmd.Flags().String("config", "", "Path to YAML config file")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate example config templates",
	Run: func(cmd *cobra.Command, args []string) {
		tmpl, _ := cmd.Flags().GetString("template")
		out, _ := cmd.Flags().GetString("output")
		var content string
		var defaultName string
		switch tmpl {
		case "simple":
			content = config.SimpleTemplate
			defaultName = "simple_test.yaml"
		case "stress":
			content = config.StressTemplate
			defaultName = "stress_test.yaml"
		default:
			log.Fatalf("unknown template: %s", tmpl)
		}

		path := out
		if path == "" {
			path = filepath.Join("configs", defaultName)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatalf("failed to create output dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			log.Fatalf("failed to write file: %v", err)
		}
		fmt.Printf("Wrote %s template to %s\n", tmpl, path)
	},
}

func init() {
	initCmd.Flags().String("template", "simple", "Template to generate: simple|stress")
	initCmd.Flags().String("output", "", "Output file path (optional)")
}
