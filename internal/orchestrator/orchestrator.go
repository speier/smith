package orchestrator

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/speier/smith/internal/coordinator"
)

type Config struct {
	ProjectPath string
	MaxParallel int
	DryRun      bool
}

type Orchestrator struct {
	config Config
	coord  *coordinator.Coordinator
	wg     sync.WaitGroup
	sem    chan struct{} // Semaphore for max parallel agents
}

func New(config Config) *Orchestrator {
	return &Orchestrator{
		config: config,
		coord:  coordinator.New(config.ProjectPath),
		sem:    make(chan struct{}, config.MaxParallel),
	}
}

func (o *Orchestrator) Run() error {
	fmt.Println("📋 Reading task queue...")

	tasks, err := o.coord.GetAvailableTasks()
	if err != nil {
		return fmt.Errorf("failed to read tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Println("✨ No available tasks")
		return nil
	}

	fmt.Printf("Found %d available tasks\n\n", len(tasks))

	// Process each task
	for _, task := range tasks {
		o.wg.Add(1)
		go o.processTask(task)
	}

	// Wait for all agents to complete
	o.wg.Wait()

	return nil
}

func (o *Orchestrator) processTask(task coordinator.Task) {
	defer o.wg.Done()

	// Acquire semaphore (limit concurrency)
	o.sem <- struct{}{}
	defer func() { <-o.sem }()

	fmt.Printf("🚀 Spawning %s agent for task #%s: %s\n",
		task.Role, task.ID, task.Title)

	if o.config.DryRun {
		time.Sleep(2 * time.Second) // Simulate work
		fmt.Printf("  ✅ [DRY RUN] Task #%s completed\n", task.ID)
		return
	}

	// Spawn agent subprocess
	cmd := exec.Command("smith", "agent",
		"--role", task.Role,
		"--task", task.ID,
		"--path", o.config.ProjectPath,
	)

	// Stream output
	cmd.Stdout = &prefixedWriter{prefix: fmt.Sprintf("[%s #%s] ", task.Role, task.ID)}
	cmd.Stderr = &prefixedWriter{prefix: fmt.Sprintf("[%s #%s ERR] ", task.Role, task.ID)}

	if err := cmd.Run(); err != nil {
		fmt.Printf("  ❌ Agent failed for task #%s: %v\n", task.ID, err)
		return
	}

	fmt.Printf("  ✅ Task #%s completed\n", task.ID)
}

// prefixedWriter prefixes each line with a string
type prefixedWriter struct {
	prefix string
}

func (w *prefixedWriter) Write(p []byte) (n int, err error) {
	fmt.Print(w.prefix + string(p))
	return len(p), nil
}
