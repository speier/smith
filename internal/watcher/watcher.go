package watcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/speier/smith/internal/coordinator"
	"github.com/speier/smith/internal/orchestrator"
)

type Config struct {
	ProjectPath   string
	CheckInterval int
	MaxParallel   int
	AutoApprove   bool
}

type Watcher struct {
	config        Config
	coord         *coordinator.FileCoordinator
	lastTodoHash  string
	orchestrating bool
}

func New(config Config) *Watcher {
	return &Watcher{
		config: config,
		coord:  coordinator.New(config.ProjectPath),
	}
}

func (w *Watcher) Start() error {
	fmt.Println("🔍 Starting file watcher...")

	// Initial hash
	hash, err := w.hashFile("TODO.md")
	if err != nil {
		return fmt.Errorf("failed to read TODO.md: %w", err)
	}
	w.lastTodoHash = hash

	ticker := time.NewTicker(time.Duration(w.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	fmt.Println("✅ Watching for changes... (Ctrl+C to stop)")
	fmt.Println()

	for {
		select {
		case <-ticker.C:
			w.check()
		}
	}
}

func (w *Watcher) check() {
	// Check if TODO.md changed
	currentHash, err := w.hashFile("TODO.md")
	if err != nil {
		fmt.Printf("⚠️  Error reading TODO.md: %v\n", err)
		return
	}

	if currentHash == w.lastTodoHash {
		return // No changes
	}

	fmt.Printf("\n📝 TODO.md changed! (detected at %s)\n", time.Now().Format("15:04:05"))
	w.lastTodoHash = currentHash

	// Prevent concurrent orchestrations
	if w.orchestrating {
		fmt.Println("⏸️  Orchestration already in progress, skipping...")
		return
	}

	w.orchestrating = true
	defer func() { w.orchestrating = false }()

	// Run workflow
	w.runWorkflow()
}

func (w *Watcher) runWorkflow() {
	fmt.Println()
	fmt.Println("🚀 Starting workflow pipeline...")
	fmt.Println()

	// Step 1: Check for feature requests (unstructured tasks)
	if w.hasFeatureRequests() {
		fmt.Println("📋 Step 1: Planning Agent - Breaking down features...")
		if err := w.runPlanningAgent(); err != nil {
			fmt.Printf("❌ Planning failed: %v\n", err)
			return
		}
		fmt.Println("✅ Planning complete")
		fmt.Println()
	}

	// Step 2: Check for available tasks
	tasks, err := w.coord.GetAvailableTasks()
	if err != nil {
		fmt.Printf("❌ Failed to read tasks: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("✨ No available tasks")
		return
	}

	fmt.Printf("📊 Step 2: Found %d tasks ready for implementation\n\n", len(tasks))

	// Step 3: Orchestrate implementation + testing + review
	orc := orchestrator.New(orchestrator.Config{
		ProjectPath: w.config.ProjectPath,
		MaxParallel: w.config.MaxParallel,
		DryRun:      false,
	})

	if err := orc.Run(); err != nil {
		fmt.Printf("❌ Orchestration failed: %v\n", err)
		return
	}

	fmt.Println("\n🎉 Workflow pipeline completed!")
	w.printSummary()
}

func (w *Watcher) hasFeatureRequests() bool {
	// Check TODO.md for "## Feature Requests" section or similar
	// This is where user adds high-level features before planning agent breaks them down
	content, err := w.readFile("TODO.md")
	if err != nil {
		return false
	}

	// Simple check: look for "Feature Request" or unmarked items
	// TODO: More sophisticated detection
	return len(content) > 0
}

func (w *Watcher) runPlanningAgent() error {
	// Spawn planning agent to process feature requests
	// The planning agent will:
	// 1. Read feature requests from TODO.md
	// 2. Break them down into atomic tasks
	// 3. Update TODO.md with structured tasks

	// For now, placeholder
	fmt.Println("  🤖 Planning agent analyzing features...")
	time.Sleep(1 * time.Second) // Simulate work
	fmt.Println("  ✍️  Created 3 new tasks")

	return nil
}

func (w *Watcher) printSummary() {
	stats, _ := w.coord.GetTaskStats()
	fmt.Printf("\n📈 Current Status:\n")
	fmt.Printf("   ✅ Done: %d\n", stats.Done)
	fmt.Printf("   🔄 In Progress: %d\n", stats.InProgress)
	fmt.Printf("   ⏳ Available: %d\n", stats.Available)
}

func (w *Watcher) hashFile(name string) (string, error) {
	path := filepath.Join(w.config.ProjectPath, name)
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (w *Watcher) readFile(name string) (string, error) {
	path := filepath.Join(w.config.ProjectPath, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
