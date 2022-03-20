package execute

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/rule"
	"github.com/plugin-ops/pedestal/pedestal/util"

	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
)

type CallbackFunc func(r rule.Rule) error

// Executor is used to execute the script, it should not be one-off, and it should be idempotent every time the script is executed
type Executor interface {

	// Start the executor
	Start() error

	// Stop the executor
	Stop() error

	// CheckAction It is used to check whether the script can be executed.
	//The []string of the return value informs the caller that it cannot be executed due to the lack of dependent actions.
	//The error is a running error. A simple lack of dependencies will not report an error.
	CheckAction(r rule.Rule) ([]string, error)

	// Execute A separate thread will immediately execute the script without queuing for the thread pool to become free
	// Returns the ID of the executed script and execution error
	Execute(r rule.Rule, callback CallbackFunc) (string, error)

	// Add will add the script to the execution queue for execution
	// Returns the ID of the executed script and execution error
	Add(r rule.Rule, callback CallbackFunc) (string, error)

	// AddScheduledScript A timed script will be added, and the timed script will never be automatically recycled
	// Returns the ID of the executed script and execution error
	AddScheduledScript(cron string, r rule.Rule, callback CallbackFunc) (string, error)

	// RemoveScript will stop the corresponding script execution plan based on the script id and remove the script from itself
	RemoveScript(scriptID string) error

	// Clean will immediately clean up the garbage
	// The garbage includes scripts that have been executed
	// But does not include scripts that are not executed, scripts that are being executed, and scripts that are executed regularly
	Clean() error

	// GetScript will get the script object saved by itself according to the script id
	GetScript(id string) rule.Rule
}

type BuiltInExecutor struct {
	config ExecutorConfig

	isStart    bool
	counter    uint64
	threadPool *ants.Pool
	cronRunner *cron.Cron

	executeQueue *util.StringQueue // element: task id
	endQueue     *util.StringQueue // element: task id
	scriptSet    map[string] /*script id*/ *task

	mutex *sync.RWMutex
}

func NewBuiltInExecutor(configs ...ExecutorConfig) (*BuiltInExecutor, error) {
	b := &BuiltInExecutor{
		config:       DefaultExecutorConfig,
		executeQueue: util.NewStringQueue(),
		endQueue:     util.NewStringQueue(),
		cronRunner:   cron.New(),
		mutex:        &sync.RWMutex{},
	}
	(&b.config).Update(configs...)

	var err error

	b.threadPool, err = ants.NewPool(int(b.config.MaxPoolSize))
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *BuiltInExecutor) run() {
	for {
		b.mutex.RLock()
		isStart := b.isStart
		b.mutex.RUnlock()
		if !isStart {
			break
		}

		id := b.executeQueue.Pull()
		t := b.scriptSet[id]
		switch t.taskType {
		case TaskType_Once:
			b.addToThreadPool(t)
		case TaskType_Cycle:
			cronID, err := b.cronRunner.AddFunc(t.cron, func() {
				b.addToThreadPool(t)
			})
			if err != nil {
				t.failed(err)
			}
			t.cronID = fmt.Sprintf("%v", cronID)
		}

	}
}

func (b *BuiltInExecutor) addToThreadPool(t *task) {
	err := b.threadPool.Submit(func() {
		b.execute(t)
	})
	if err != nil {
		t.failed(err)
	}
}

func (b *BuiltInExecutor) Start() error {
	b.mutex.Lock()
	b.isStart = true
	b.mutex.Unlock()
	go b.run()
	return nil
}

func (b *BuiltInExecutor) Stop() error {
	b.mutex.Lock()
	b.isStart = false
	b.mutex.Unlock()
	return nil
}

func (b *BuiltInExecutor) CheckAction(r rule.Rule) ([]string, error) {
	relyOn := []string{}
	for name := range r.Info().GetRelyOn() {
		relyOn = append(relyOn, name)
	}
	notExist := action.CheckActionExist(relyOn...)
	return notExist, nil
}

func (b *BuiltInExecutor) Execute(r rule.Rule, callback CallbackFunc) (string, error) {
	b.mutex.Lock()
	b.counter++
	id := fmt.Sprintf("%v", b.counter)
	t := &task{
		Rule:     r,
		status:   TaskStatus_Wait,
		taskType: TaskType_Once,
		callback: callback,
	}
	b.scriptSet[id] = t
	b.mutex.Unlock()
	b.execute(t)
	return id, t.Error
}

var ErrNotAction = fmt.Errorf("the following actions are not loaded in the pedestal, the script cannot be executed")

func (b *BuiltInExecutor) execute(t *task) {
	t.start()
	sc := t.Rule

	{ // check rely on
		notExist, err := b.CheckAction(sc)
		if err != nil {
			t.failed(err)
			return
		}
		if len(notExist) != 0 {
			t.failed(ErrNotAction)
			return
		}

	}
	{ // add rely on
		for d, r := range t.Rule.Info().GetRelyOn() {
			a := action.GetAction(d)
			if a == nil {
				t.failed(ErrNotAction)
				return
			}
			err := sc.Set(r, a)
			if err != nil {
				t.failed(err)
				return
			}
		}
	}
	{ // do script
		err := sc.Compile()
		if err != nil {
			t.failed(err)
			return
		}
		err = sc.Do(b.config.Ctx)
		if err != nil {
			var e error
			if t.callback != nil {
				e = t.callback(t.Rule)
			}
			t.failed(fmt.Errorf("running rule failed, error: %v; callback error: %v", err, e))
			return
		}
	}

	t.over()
}

func (b *BuiltInExecutor) Add(r rule.Rule, callback CallbackFunc) (string, error) {
	b.mutex.Lock()
	b.counter++
	id := fmt.Sprintf("%v", b.counter)
	b.scriptSet[id] = &task{
		Rule:     r,
		taskType: TaskType_Once,
		status:   TaskStatus_Wait,
		callback: callback,
	}
	b.executeQueue.Push(id)
	b.mutex.Unlock()
	return id, nil
}

func (b *BuiltInExecutor) AddScheduledScript(cron string, r rule.Rule, callback CallbackFunc) (string, error) {
	b.mutex.Lock()
	b.counter++
	id := fmt.Sprintf("%v", b.counter)
	b.scriptSet[id] = &task{
		Rule:     r,
		cron:     cron,
		taskType: TaskType_Cycle,
		status:   TaskStatus_Wait,
		callback: callback,
	}
	b.executeQueue.Push(id)
	b.mutex.Unlock()
	return id, nil
}

func (b *BuiltInExecutor) RemoveScript(scriptID string) error {
	b.mutex.Lock()
	delete(b.scriptSet, scriptID)
	b.mutex.Unlock()
	return nil
}

func (b *BuiltInExecutor) Clean() error {
	b.mutex.Lock()
	size := b.endQueue.Size()
	for i := 0; i < size; i++ {
		delete(b.scriptSet, b.endQueue.Pull())
	}
	b.mutex.Unlock()

	return nil
}

func (b *BuiltInExecutor) GetScript(id string) rule.Rule {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.scriptSet[id].Rule
}

type task struct {
	rule.Rule

	cron        string
	cronID      string
	status      TaskStatus
	taskType    TaskType
	callback    CallbackFunc
	runningTime time.Time
	endTime     time.Time

	Error error
}

func (t *task) failed(err error) {
	t.Error = err
	t.status = TaskStatus_Fail
	t.endTime = time.Now()
}

func (t *task) over() {
	t.status = TaskStatus_OK
	t.endTime = time.Now()
}

func (t *task) start() {
	t.status = TaskStatus_Run
	t.runningTime = time.Now()
}

type TaskStatus string

const (
	TaskStatus_OK   = "ok"
	TaskStatus_Wait = "wait"
	TaskStatus_Fail = "fail"
	TaskStatus_Run  = "run"
)

type TaskType string

const (
	TaskType_Once  = "once"
	TaskType_Cycle = "cycle"
)

// ExecutorConfig is Executor's config
type ExecutorConfig struct {

	// Ctx is the context that may be passed to the executor
	Ctx context.Context

	// MaxPoolSize unlimited by default
	MaxPoolSize int64

	// AutoCleanScriptQuantity in second, unlimited by default
	AutoCleanInterval int64

	// AutoCleanScriptQuantity unlimited by default
	AutoCleanScriptQuantity int64
}

func (e *ExecutorConfig) Update(config ...ExecutorConfig) {
	for _, executorConfig := range config {
		if executorConfig.MaxPoolSize > 0 {
			e.MaxPoolSize = executorConfig.MaxPoolSize
		}
		if executorConfig.AutoCleanInterval > 0 {
			e.AutoCleanInterval = executorConfig.AutoCleanInterval
		}
		if executorConfig.AutoCleanScriptQuantity > 0 {
			e.AutoCleanScriptQuantity = executorConfig.AutoCleanScriptQuantity
		}
	}
}

var DefaultExecutorConfig = ExecutorConfig{
	Ctx:                     context.TODO(),
	MaxPoolSize:             int64(runtime.NumCPU()),
	AutoCleanInterval:       0,
	AutoCleanScriptQuantity: 0,
}
