package execute

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/log"
	"github.com/plugin-ops/pedestal/pedestal/rule"
	"github.com/plugin-ops/pedestal/pedestal/util"

	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
)

type CallbackFunc func(r rule.Rule) error

// Executor is used to execute the rule, it should not be one-off, and it should be idempotent every time the rule is executed
type Executor interface {

	// Start the executor
	Start() error

	// Stop the executor
	Stop() error

	// CheckAction It is used to check whether the rule can be executed.
	//The []string of the return value informs the caller that it cannot be executed due to the lack of dependent actions.
	//The error is a running error. A simple lack of dependencies will not report an error.
	CheckAction(r rule.Rule) ([]string, error)

	// Execute A separate thread will immediately execute the rule without queuing for the thread pool to become free
	// Returns the ID of the executed rule and execution error
	Execute(r rule.Rule, params map[string]interface{}, callback CallbackFunc) (string, error)

	// Add will add the rule to the execution queue for execution
	// Returns the ID of the executed rule and execution error
	AddTask(r rule.Rule, params map[string]interface{}, callback CallbackFunc) (string, error)

	// AddScheduledRule A timed rule will be added, and the timed rule will never be automatically recycled
	// Returns the ID of the executed rule and execution error
	AddScheduledTask(cron string, r rule.Rule, params map[string]interface{}, callback CallbackFunc) (string, error)

	// RemoveRule will stop the corresponding Task execution plan based on the rule id and remove the rule from itself
	RemoveTask(taskID string) error

	// Clean will immediately clean up the garbage
	// The garbage includes rules that have been executed
	// But does not include rules that are not executed, rules that are being executed, and rules that are executed regularly
	Clean() error

	// GetRule will get the rule object saved by itself according to the rule id
	GetTask(id string) *Task
}

var (
	std      Executor
	stdMutex = &sync.RWMutex{}
)

func InitExecute() (err error) {
	stdMutex.Lock()
	defer stdMutex.Unlock()
	std, err = NewBuiltInExecutor()
	if err != nil {
		return err
	}
	return std.Start()
}

func GetExecutor() Executor {
	stdMutex.RLock()
	defer stdMutex.RUnlock()
	return std
}

type BuiltInExecutor struct {
	config ExecutorConfig

	isStart    bool
	threadPool *ants.Pool
	cronRunner *cron.Cron

	executeQueue *util.StringQueue // element: Task id
	endQueue     *util.StringQueue // element: Task id
	taskSet      map[string] /*rule id*/ *Task

	mutex *sync.RWMutex
	stage *log.Stage
}

func NewBuiltInExecutor(configs ...ExecutorConfig) (*BuiltInExecutor, error) {
	b := &BuiltInExecutor{
		config:       DefaultExecutorConfig,
		cronRunner:   cron.New(),
		executeQueue: util.NewStringQueue(),
		endQueue:     util.NewStringQueue(),
		taskSet:      map[string]*Task{},
		mutex:        &sync.RWMutex{},
		stage:        log.NewStage(),
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
			glog.Warning(b.stage.Context(), "executor stopped")
			break
		}

		id := b.executeQueue.Pull()
		t := b.taskSet[id]
		glog.Infof(b.stage.Context(), "ready to execute Task [%v]", id)
		switch t.TaskType {
		case TaskType_Once:
			b.addToThreadPool(t)
		case TaskType_Cycle:
			cronID, err := b.cronRunner.AddFunc(t.Cron, func() {
				b.addToThreadPool(t)
			})
			if err != nil {
				t.failed(err)
			}
			t.CronID = fmt.Sprintf("%v", cronID)
		}

	}
}

func (b *BuiltInExecutor) addToThreadPool(t *Task) {
	glog.Infof(b.stage.Context(), "add Task [%v] to Task running pool", t.TaskID)
	err := b.threadPool.Submit(func() {
		b.execute(t)
	})
	if err != nil {
		t.failed(err)
	}
}

func (b *BuiltInExecutor) Start() error {
	glog.Infof(b.stage.Context(), "start executor")
	b.mutex.Lock()
	b.isStart = true
	b.mutex.Unlock()
	go b.run()
	return nil
}

func (b *BuiltInExecutor) Stop() error {
	glog.Infof(b.stage.Context(), "stop executor")
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

var ErrNotAction = fmt.Errorf("the following actions are not loaded in the pedestal, the rule cannot be executed")

func (b *BuiltInExecutor) execute(t *Task) {
	t.start()
	sc := t.Rule

	{ // check rely on
		glog.Infof(b.stage.Context(), "check [%v] rely on", t.TaskID)
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
		glog.Infof(b.stage.Context(), "add rely on to [%v]", t.TaskID)
		for d, r := range t.Rule.Info().GetRelyOn() {
			a, exist := action.GetAction(b.stage, d, t.Rule.Info().Version())
			if !exist {
				t.failed(ErrNotAction)
				return
			}
			err := sc.AddRelyOn(r, a)
			if err != nil {
				t.failed(err)
				return
			}
		}
	}
	{ // add Params
		glog.Infof(b.stage.Context(), "add params to [%v]", t.TaskID)
		for k, v := range t.Params {
			err := sc.Set(k, action.NewValue(v))
			if err != nil {
				t.failed(err)
				return
			}
		}
	}
	{ // do rule
		glog.Infof(b.stage.Context(), "compile [%v] of rule[%v]", t.TaskID, t.Rule.Info().Key())
		err := sc.Compile()
		if err != nil {
			t.failed(err)
			return
		}

		glog.Infof(b.stage.Context(), "doing task[%v]", t.TaskID)
		err = sc.Do(b.config.Ctx)
		if err != nil {
			var e error
			if t.Callback != nil {
				e = t.Callback(t.Rule)
			}
			t.failed(fmt.Errorf("running rule failed, error: %v; Callback error: %v", err, e))
			return
		}
	}

	t.over()
}

func (b *BuiltInExecutor) Execute(r rule.Rule, params map[string]interface{}, callback CallbackFunc) (string, error) {
	b.mutex.Lock()
	id := guid.S()
	t := &Task{
		TaskID:   id,
		Rule:     r,
		Status:   TaskStatus_Wait,
		TaskType: TaskType_Once,
		Callback: callback,
		Params:   params,
		stage:    log.NewStage().Go(id),
	}
	b.taskSet[t.TaskID] = t
	b.mutex.Unlock()
	b.execute(t)
	return t.TaskID, t.Error
}

func (b *BuiltInExecutor) AddTask(r rule.Rule, params map[string]interface{}, callback CallbackFunc) (string, error) {
	b.mutex.Lock()
	id := guid.S()
	t := &Task{
		TaskID:   id,
		Rule:     r,
		TaskType: TaskType_Once,
		Status:   TaskStatus_Wait,
		Callback: callback,
		Params:   params,
		stage:    log.NewStage().Go(id),
	}
	b.taskSet[t.TaskID] = t
	b.executeQueue.Push(t.TaskID)
	b.mutex.Unlock()
	glog.Infof(b.stage.Context(), "add new task[%v] with rule[%v]", t.TaskID, r.Info().Key())
	return t.TaskID, nil
}

func (b *BuiltInExecutor) AddScheduledTask(cron string, r rule.Rule, params map[string]interface{}, callback CallbackFunc) (string, error) {
	b.mutex.Lock()
	id := guid.S()
	t := &Task{
		TaskID:   id,
		Rule:     r,
		Cron:     cron,
		TaskType: TaskType_Cycle,
		Status:   TaskStatus_Wait,
		Callback: callback,
		Params:   params,
		stage:    log.NewStage().Go(id),
	}
	b.taskSet[t.TaskID] = t
	b.executeQueue.Push(t.TaskID)
	b.mutex.Unlock()
	glog.Infof(b.stage.Context(), "add new scheduled task[%v] with rule[%v],cron: '%v'", t.TaskID, r.Info().Key(), cron)
	return t.TaskID, nil
}

func (b *BuiltInExecutor) RemoveTask(taskID string) error {
	glog.Infof(b.stage.Context(), "remove task[%v]", taskID)
	b.mutex.Lock()
	delete(b.taskSet, taskID)
	b.mutex.Unlock()
	return nil
}

func (b *BuiltInExecutor) Clean() error {
	glog.Infof(b.stage.Context(), "clean ended task cache")
	b.mutex.Lock()
	size := b.endQueue.Size()
	for i := 0; i < size; i++ {
		id := b.endQueue.Pull()
		glog.Infof(b.stage.Context(), "clean ended task[%v]", id)
		delete(b.taskSet, id)
	}
	b.mutex.Unlock()
	runtime.GC()
	return nil
}

func (b *BuiltInExecutor) GetTask(id string) *Task {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.taskSet[id]
}

type Task struct {
	rule.Rule

	TaskID      string
	Cron        string
	CronID      string
	Status      TaskStatus
	TaskType    TaskType
	Callback    CallbackFunc
	Params      map[string]interface{}
	RunningTime time.Time
	EndTime     time.Time
	Error       error

	stage *log.Stage
}

func (t *Task) failed(err error) {
	t.Error = err
	t.Status = TaskStatus_Fail
	t.EndTime = time.Now()
	glog.Infof(t.stage.Context(), "[%v] failed at %v", t.TaskID, t.RunningTime)
}

func (t *Task) over() {
	t.Status = TaskStatus_OK
	t.EndTime = time.Now()
	glog.Infof(t.stage.Context(), "[%v] finished at %v", t.TaskID, t.RunningTime)
}

func (t *Task) start() {
	t.Status = TaskStatus_Run
	t.RunningTime = time.Now()
	glog.Infof(t.stage.Context(), "[%v] started at %v", t.TaskID, t.RunningTime)
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

	// AutoCleanRuleQuantity in second, unlimited by default
	AutoCleanInterval int64

	// AutoCleanRuleQuantity unlimited by default
	AutoCleanRuleQuantity int64
}

func (e *ExecutorConfig) Update(config ...ExecutorConfig) {
	for _, executorConfig := range config {
		if executorConfig.MaxPoolSize > 0 {
			e.MaxPoolSize = executorConfig.MaxPoolSize
		}
		if executorConfig.AutoCleanInterval > 0 {
			e.AutoCleanInterval = executorConfig.AutoCleanInterval
		}
		if executorConfig.AutoCleanRuleQuantity > 0 {
			e.AutoCleanRuleQuantity = executorConfig.AutoCleanRuleQuantity
		}
	}
}

var DefaultExecutorConfig = ExecutorConfig{
	Ctx:                   context.TODO(),
	MaxPoolSize:           int64(runtime.NumCPU()),
	AutoCleanInterval:     0,
	AutoCleanRuleQuantity: 0,
}
