package kitchen

import "github.com/google/uuid"

// --- KitchenID ---

type KitchenID string

func (id KitchenID) String() string { return string(id) }

// --- TaskID ---

type TaskID string

func newTaskID() TaskID { return TaskID(uuid.New().String()) }

// --- TaskStatus ---

type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskCooking
	TaskCompleted
)

// --- MaxCapacity ---

const MaxConcurrentTasks = 10
