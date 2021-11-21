// Package singleflight provides a singleflight mechanisms similar to
// x/sync/singleflight but for messages.
package singleflight

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
)

// =============================================================================
// Manager
// =====================================================================================

type Manager struct {
	messageGroups map[discord.MessageID]*messageGroup
	mu            sync.Mutex
}

func NewManager() *Manager {
	return &Manager{messageGroups: make(map[discord.MessageID]*messageGroup)}
}

// DoAsync executes f blockingly.
// In contrast to DoSync, DoAsync may execute the function alongside other
// async functions.
// If a sync functions was added before DoAsync was called, DoAsync will wait
// for the function to finish before executing f.
func (m *Manager) DoAsync(messageID discord.MessageID, f func() error) error {
	return m.do(messageID, false, f)
}

// DoSync executes f blockingly.
// Once DoSync is called, it will wait for all already executing async
// functions to finish.
// Then it will execute f.
// Only one sync function will be executed at a time.
//
// Once f finishes executing, all functions added after f will be executed.
func (m *Manager) DoSync(messageID discord.MessageID, f func() error) error {
	return m.do(messageID, true, f)
}

// do executes f blockingly.
// Functions will be executed in the order they were given to do.
// If alone is false, the function may be executed alongside other functions
// who also have alone set to false.
// If alone is set to true, the function will be executed once all preceding
// non-alone functions have completed.
// During the execution of an alone function, no other function will be
// executed.
func (m *Manager) do(messageID discord.MessageID, sync bool, f func() error) error {
	m.mu.Lock()
	// m.mu is unlocked by mg.do

	group := m.messageGroups[messageID]
	if group == nil {
		group = newMessageGroup(m, messageID)
		m.messageGroups[messageID] = group
	}

	var err error
	group.do(sync, func() { err = f() })

	return err
}

// =============================================================================
// messageGroup
// =====================================================================================

type messageGroup struct {
	m *Manager

	id discord.MessageID

	queue  chan *task
	userWG sync.WaitGroup

	remTasks uint64
}

func newMessageGroup(m *Manager, id discord.MessageID) *messageGroup {
	mg := &messageGroup{
		m:     m,
		id:    id,
		queue: make(chan *task),
	}

	go mg.loop()

	return mg
}

type task struct {
	alone bool
	f     func()
}

// do executes f blockingly.
// Functions will be executed in the order they were given to do.
// Multiple functions per user and message can be executed.
// If userID is 0 the function will only be executed once all user functions
// have finished executing.
//
// When do is called, m.m.mu must be locked.
// Before returning, m.m.mu will be unlocked.
func (m *messageGroup) do(alone bool, f func()) {
	m.remTasks++

	m.m.mu.Unlock()

	done := make(chan struct{})

	m.queue <- &task{
		alone: alone,
		f: func() {
			f()
			close(done)
		},
	}

	<-done
}

func (m *messageGroup) loop() {
	for t := range m.queue {
		if t.alone {
			m.userWG.Wait()

			t.f()
			m.done()
		} else {
			m.userWG.Add(1)
			go func() {
				t.f()

				m.userWG.Done()
				m.done()
			}()
		}
	}
}

func (m *messageGroup) done() {
	m.m.mu.Lock()
	defer m.m.mu.Unlock()

	m.remTasks--
	if m.remTasks == 0 {
		delete(m.m.messageGroups, m.id)
		close(m.queue)
	}
}
