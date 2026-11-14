package llm

import (
	"encoding/json"
	"sync"
)

const replayCacheMax = 256

// replayCache keeps a response's raw output blocks, keyed by the ID of its
// first tool call, so the follow-up request can send them back unchanged
// with the tool results. Claude's thinking blocks and OpenAI's reasoning
// items must come back that way, and Edda's message history only carries
// text and tool calls. The oldest entries are dropped after replayCacheMax.
type replayCache struct {
	mu    sync.Mutex
	items map[string][]json.RawMessage
	order []string
}

func (c *replayCache) save(firstCallID string, blocks []json.RawMessage) {
	if firstCallID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		c.items = map[string][]json.RawMessage{}
	}
	if _, ok := c.items[firstCallID]; !ok {
		c.order = append(c.order, firstCallID)
	}
	c.items[firstCallID] = blocks
	for len(c.order) > replayCacheMax {
		delete(c.items, c.order[0])
		c.order = c.order[1:]
	}
}

func (c *replayCache) get(firstCallID string) []json.RawMessage {
	if firstCallID == "" {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.items[firstCallID]
}
