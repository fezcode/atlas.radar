package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	repoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#2B2D42")).
			Padding(0, 1)

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	cleanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	dirtyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	aheadStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF"))

	behindStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAF00"))
)

type GitStatus struct {
	Branch   string
	IsDirty  bool
	Ahead    int
	Behind   int
	Modified int
	Added    int
	Deleted  int
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [directory]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  atlas.radar                          # Scan current directory")
		fmt.Println("  atlas.radar D:\\Projects              # Scan specific directory")
		fmt.Println("  atlas.radar --show unclean --watch   # Monitor only dirty repos")
		fmt.Println("  atlas.radar --fetch                  # Fetch all repositories")
	}

	showFlag := flag.String("show", "all", "Filter repositories (all, clean, unclean)")
	watchFlag := flag.Bool("watch", false, "Continuously monitor status")
	fetchFlag := flag.Bool("fetch", false, "Fetch updates for all repositories")
	pullFlag := flag.Bool("pull", false, "Pull updates for all repositories")
	pushFlag := flag.Bool("push", false, "Push updates for all repositories")
	flag.Parse()

	targetDir := "."
	if flag.NArg() > 0 {
		targetDir = flag.Arg(0)
	}

	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Handle one-time operations: fetch, pull, push
	if *fetchFlag || *pullFlag || *pushFlag {
		handleBulkOperations(absDir, *fetchFlag, *pullFlag, *pushFlag)
		return
	}

	for {
		if *watchFlag {
			clearScreen()
		}

		fmt.Println(titleStyle.Render("Radar") + " scanning " + absDir)
		if *showFlag != "all" {
			fmt.Printf("Filtering: %s\n", *showFlag)
		}
		fmt.Println()

		entries, err := os.ReadDir(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
			os.Exit(1)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			repoPath := filepath.Join(absDir, entry.Name())
			gitPath := filepath.Join(repoPath, ".git")

			if _, err := os.Stat(gitPath); os.IsNotExist(err) {
				continue
			}

			status, err := getGitStatus(repoPath)
			if err != nil {
				continue
			}

			// Filter logic
			shouldShow := true
			switch *showFlag {
			case "clean":
				if status.IsDirty || status.Ahead > 0 || status.Behind > 0 {
					shouldShow = false
				}
			case "unclean":
				if !status.IsDirty && status.Ahead == 0 && status.Behind == 0 {
					shouldShow = false
				}
			}

			if shouldShow {
				printStatus(entry.Name(), status)
			}
		}

		if !*watchFlag {
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func handleBulkOperations(absDir string, fetch, pull, push bool) {
	entries, err := os.ReadDir(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		return
	}

	successCount := 0
	failCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoPath := filepath.Join(absDir, entry.Name())
		gitPath := filepath.Join(repoPath, ".git")

		if _, err := os.Stat(gitPath); os.IsNotExist(err) {
			continue
		}

		var cmd *exec.Cmd
		var opName string

		if fetch {
			opName = "fetch"
			cmd = exec.Command("git", "fetch")
		} else if pull {
			opName = "pull"
			cmd = exec.Command("git", "pull")
		} else if push {
			opName = "push"
			cmd = exec.Command("git", "push")
		}

		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			fmt.Printf("%s [%s]: %s\n", dirtyStyle.Render("FAIL"), opName, repoStyle.Render(entry.Name()))
			failCount++
		} else {
			fmt.Printf("%s [%s]: %s\n", cleanStyle.Render("OK"), opName, repoStyle.Render(entry.Name()))
			successCount++
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d successful, %d failed\n", successCount, failCount)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func getGitStatus(path string) (*GitStatus, error) {
	cmd := exec.Command("git", "status", "--branch", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty status output")
	}

	status := &GitStatus{}

	// Parse branch info
	branchLine := lines[0]
	if strings.HasPrefix(branchLine, "## ") {
		branchPart := strings.TrimPrefix(branchLine, "## ")
		parts := strings.Split(branchPart, "...")
		status.Branch = parts[0]

		if len(parts) > 1 {
			if idx := strings.Index(parts[1], "["); idx != -1 {
				info := strings.Trim(parts[1][idx:], "[]")
				for _, part := range strings.Split(info, ", ") {
					if strings.HasPrefix(part, "ahead ") {
						fmt.Sscanf(part, "ahead %d", &status.Ahead)
					} else if strings.HasPrefix(part, "behind ") {
						fmt.Sscanf(part, "behind %d", &status.Behind)
					}
				}
			}
		}
	}

	// Parse file changes
	for _, line := range lines[1:] {
		if len(line) < 3 {
			continue
		}
		status.IsDirty = true
		x := line[0]
		y := line[1]

		if x == 'M' || y == 'M' {
			status.Modified++
		} else if x == 'A' || y == '?' {
			status.Added++
		} else if x == 'D' || y == 'D' {
			status.Deleted++
		}
	}

	return status, nil
}

func printStatus(name string, status *GitStatus) {
	nameDisplay := repoStyle.Render(name)
	branchDisplay := branchStyle.Render("(" + status.Branch + ")")

	var statusParts []string

	if status.Ahead > 0 {
		statusParts = append(statusParts, aheadStyle.Render(fmt.Sprintf("↑%d", status.Ahead)))
	}
	if status.Behind > 0 {
		statusParts = append(statusParts, behindStyle.Render(fmt.Sprintf("↓%d", status.Behind)))
	}

	if !status.IsDirty && status.Ahead == 0 && status.Behind == 0 {
		statusParts = append(statusParts, cleanStyle.Render("clean"))
	} else if status.IsDirty {
		var details []string
		if status.Added > 0 {
			details = append(details, fmt.Sprintf("+%d", status.Added))
		}
		if status.Modified > 0 {
			details = append(details, fmt.Sprintf("~%d", status.Modified))
		}
		if status.Deleted > 0 {
			details = append(details, fmt.Sprintf("-%d", status.Deleted))
		}
		statusParts = append(statusParts, dirtyStyle.Render(strings.Join(details, " ")))
	}

	fmt.Printf("%s %s %s\n", nameDisplay, branchDisplay, strings.Join(statusParts, " "))
}
