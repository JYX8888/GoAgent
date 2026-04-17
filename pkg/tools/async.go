package tools

import (
	"context"
	"fmt"
	"sync"
)

type Task struct {
	ToolName  string
	InputData string
}

type TaskResult struct {
	TaskID    int
	ToolName  string
	InputData string
	Result    string
	Status    string
}

type AsyncToolExecutor struct {
	Registry   *ToolRegistry
	MaxWorkers int
	semaphore  chan struct{}
	wg         sync.WaitGroup
}

func NewAsyncToolExecutor(registry *ToolRegistry, maxWorkers int) *AsyncToolExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	return &AsyncToolExecutor{
		Registry:   registry,
		MaxWorkers: maxWorkers,
		semaphore:  make(chan struct{}, maxWorkers),
	}
}

func (e *AsyncToolExecutor) ExecuteToolAsync(ctx context.Context, toolName, inputData string) string {
	select {
	case <-ctx.Done():
		return fmt.Sprintf("❌ Tool '%s' async execution cancelled", toolName)
	default:
	}

	e.wg.Add(1)
	defer e.wg.Done()

	e.semaphore <- struct{}{}
	defer func() { <-e.semaphore }()

	result := e.Registry.Execute(toolName, map[string]interface{}{"input": inputData})
	return result
}

func (e *AsyncToolExecutor) ExecuteToolsParallel(ctx context.Context, tasks []Task) []TaskResult {
	fmt.Printf("🚀 Starting parallel execution of %d tasks\n", len(tasks))

	results := make([]TaskResult, len(tasks))
	var resultMu sync.Mutex
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t Task) {
			defer wg.Done()

			result := e.ExecuteToolAsync(ctx, t.ToolName, t.InputData)

			resultMu.Lock()
			results[idx] = TaskResult{
				TaskID:    idx,
				ToolName:  t.ToolName,
				InputData: t.InputData,
				Result:    result,
				Status:    "success",
			}
			resultMu.Unlock()

			fmt.Printf("✅ Task %d completed: %s\n", idx+1, t.ToolName)
		}(i, task)
	}

	wg.Wait()

	successCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		}
	}
	fmt.Printf("🎉 Parallel execution completed, success: %d/%d\n", successCount, len(results))

	return results
}

func (e *AsyncToolExecutor) ExecuteToolsBatch(ctx context.Context, toolName string, inputList []string) []TaskResult {
	tasks := make([]Task, len(inputList))
	for i, input := range inputList {
		tasks[i] = Task{ToolName: toolName, InputData: input}
	}
	return e.ExecuteToolsParallel(ctx, tasks)
}

func (e *AsyncToolExecutor) Close() {
	e.wg.Wait()
	fmt.Println("🔒 Async tool executor closed")
}

func RunParallelTools(registry *ToolRegistry, tasks []Task, maxWorkers int) []TaskResult {
	executor := NewAsyncToolExecutor(registry, maxWorkers)
	defer executor.Close()

	ctx := context.Background()
	return executor.ExecuteToolsParallel(ctx, tasks)
}

func RunBatchTool(registry *ToolRegistry, toolName string, inputList []string, maxWorkers int) []TaskResult {
	executor := NewAsyncToolExecutor(registry, maxWorkers)
	defer executor.Close()

	ctx := context.Background()
	return executor.ExecuteToolsBatch(ctx, toolName, inputList)
}
