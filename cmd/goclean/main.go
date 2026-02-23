package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alterpix/linux-cleaner/internal/cleaner"
	"github.com/alterpix/linux-cleaner/internal/tui"
	"github.com/alterpix/linux-cleaner/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Simulate cleaning without deleting files")
	noConfirm := flag.Bool("yes", false, "Skip confirmation prompt")
	useTUI := flag.Bool("tui", true, "Use Text User Interface (default true)")
	noTUI := flag.Bool("no-tui", false, "Disable TUI and use CLI mode (overrides -tui)")

	flag.Parse()

	// Logic to determine if we use TUI
	// If no-tui is true, disable checks.
	isManualCLI := *dryRun || *noConfirm || !*useTUI || *noTUI

	cleaners := []cleaner.Cleaner{
		&cleaner.AptCleaner{},
		&cleaner.LogCleaner{},
		&cleaner.TrashCleaner{},
		&cleaner.UserCacheCleaner{},
		&cleaner.DockerCleaner{},
		&cleaner.BrowserCleaner{},
		&cleaner.GoCacheCleaner{},
		&cleaner.DynamicCacheCleaner{},
		&cleaner.NpmCacheCleaner{},
		&cleaner.FlatpakCleaner{},
		&cleaner.TmpCleaner{},
		&cleaner.CargoCacheCleaner{},
		&cleaner.AppCacheCleaner{},
		&cleaner.LargeFileCleaner{SkipConfirmation: *noConfirm},
	}

	if !isManualCLI {
		// Start TUI
		p := tea.NewProgram(tui.InitialModel(cleaners))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		return
	}

	// Legacy CLI Mode (if --no-tui or scripting flags used)
	runCLI(cleaners, *dryRun, *noConfirm)
}

// runCLI contains the old main function logic
func runCLI(cleaners []cleaner.Cleaner, dryRun, noConfirm bool) {
	ui.Bold("Linux System Cleaner (CLI Mode)\n")
	ui.Info("-------------------------------\n")

	var totalSize int64
	var cleanable []cleaner.Cleaner

	// Scan Phase
	ui.Info("Scanning system...\n")
	for _, c := range cleaners {
		fmt.Printf("Scanning %s... ", c.Name())
		size, err := c.Scan()
		if err != nil {
			ui.Error("[ERROR] %v\n", err)
			continue
		}

		if size > 0 {
			ui.Success("Found %s\n", ui.PrintSize(size))
			totalSize += size
			cleanable = append(cleanable, c)
		} else {
			fmt.Println("Clean")
		}
	}

	if totalSize == 0 {
		ui.Success("\nSystem is already clean!\n")
		return
	}

	ui.Bold("\nTotal reclaimable space: %s\n", ui.PrintSize(totalSize))

	if dryRun {
		ui.Warning("\n[DRY RUN] No changes were made.\n")
		return
	}

	// Confirmation Phase
	if !noConfirm {
		ui.Warning("\nWARNING: This will permanently delete the listed files.")
		fmt.Print("Are you sure you want to proceed? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			ui.Info("Cleanup cancelled.\n")
			return
		}
	}

	// Clean Phase
	ui.Info("\nCleaning...\n")
	for _, c := range cleanable {
		fmt.Printf("Cleaning %s... ", c.Name())
		if err := c.Clean(); err != nil {
			ui.Error("FAILED: %v\n", err)
		} else {
			ui.Success("Done\n")
		}
	}
	ui.Success("\nCleanup complete!\n")
}
