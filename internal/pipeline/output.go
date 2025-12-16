package pipeline

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xiaolfeng/builder-cli/internal/config"
	"github.com/xiaolfeng/builder-cli/internal/executor"
	"github.com/xiaolfeng/builder-cli/internal/types"
)

type outputBatcher struct {
	taskID        string
	send          func(tea.Msg)
	flushInterval time.Duration
	maxBatch      int

	ch       chan types.OutputLine
	done     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

func newOutputBatcher(taskID string, send func(tea.Msg)) *outputBatcher {
	b := &outputBatcher{
		taskID:        taskID,
		send:          send,
		flushInterval: 120 * time.Millisecond,
		maxBatch:      200,
		ch:            make(chan types.OutputLine, 2048),
		done:          make(chan struct{}),
	}

	b.wg.Add(1)
	go b.loop()

	return b
}

func (b *outputBatcher) handle(line string, isError bool) {
	l := types.OutputLine{Line: line, IsError: isError}
	select {
	case <-b.done:
		b.send(types.NewOutputMsg(b.taskID, l.Line, l.IsError))
	case b.ch <- l:
	}
}

func (b *outputBatcher) stop() {
	b.stopOnce.Do(func() { close(b.done) })
	b.wg.Wait()
}

func (b *outputBatcher) loop() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	batch := make([]types.OutputLine, 0, b.maxBatch)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		lines := make([]types.OutputLine, len(batch))
		copy(lines, batch)
		b.send(types.NewOutputBatchMsg(b.taskID, lines))
		batch = batch[:0]
	}

	drain := func() {
		for {
			select {
			case l := <-b.ch:
				batch = append(batch, l)
				if len(batch) >= b.maxBatch {
					flush()
				}
			default:
				flush()
				return
			}
		}
	}

	for {
		select {
		case l := <-b.ch:
			batch = append(batch, l)
			if len(batch) >= b.maxBatch {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-b.done:
			drain()
			return
		}
	}
}

func (p *Pipeline) newTaskOutputHandler(task *Task) (executor.OutputHandler, func()) {
	sendLine := func(line string, isError bool) {
		p.sendMsg(types.NewOutputMsg(task.ID, line, isError))
	}

	// 仅对 Docker 构建 / SSH 远程支持可选“强制降级刷新”（减少消息频率，避免刷屏）
	if task.Type != config.TaskTypeDockerBuild && task.Type != config.TaskTypeSSH {
		return sendLine, func() {}
	}

	// force_refresh=true：启用降级（批量）以避免刷屏；否则保持逐行实时输出
	if !task.Config.ForceRefresh {
		return sendLine, func() {}
	}

	batcher := newOutputBatcher(task.ID, p.sendMsg)
	return batcher.handle, batcher.stop
}
