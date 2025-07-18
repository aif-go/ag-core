package ag_conf

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

var WatcherM *WatcherManager
var global_watcherChan = make(chan Watcher, 512) // TODO 评估该缓存方式的可行性

// RegisterWatcher 注册watcher
func RegisterWatcher(w Watcher) {
	global_watcherChan <- w
}

type ConfigChangeListener func(k, v string)
type ChangePropertySources func(propertySources []IPropertySource)

type Watcher interface {
	// Start(context.Context, chan []IPropertySource)
	Start(context.Context, ChangePropertySources)
	Stop()
}

type WatcherManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	env         IConfigurableEnvironment
	refreshMap  map[string][]func(k, v string)
	refreshChan chan []IPropertySource
	watcherChan chan Watcher

	watchers []Watcher

	refreshMapLock sync.RWMutex
	watchersLock   sync.RWMutex
	closeOnce      sync.Once
}

func NewConfigWatcherManager(env IConfigurableEnvironment) *WatcherManager {
	ctx, cancel := context.WithCancel(context.Background())

	watcher := &WatcherManager{
		ctx:            ctx,
		cancel:         cancel,
		env:            env,
		refreshMap:     make(map[string][]func(k, v string)),
		refreshChan:    make(chan []IPropertySource, 50),
		watcherChan:    global_watcherChan,
		watchers:       make([]Watcher, 0),
		refreshMapLock: sync.RWMutex{},
		watchersLock:   sync.RWMutex{},
	}
	WatcherM = watcher
	return watcher
}

func (wm *WatcherManager) AddConfigChangeListener(key string, listener ConfigChangeListener) {
	wm.refreshMapLock.Lock()
	defer wm.refreshMapLock.Unlock()

	if _, ok := wm.refreshMap[key]; !ok {
		wm.refreshMap[key] = make([]func(k, v string), 0)
	}
	wm.refreshMap[key] = append(wm.refreshMap[key], listener)
}

func (wm *WatcherManager) RegisterWatcher(w Watcher) {
	wm.watcherChan <- w
}

func (wm *WatcherManager) run(ctx context.Context) {
	defer func() {
		slog.Info("watcherManager run stoped")
	}()

	for {
		select {
		case watcher := <-wm.watcherChan:
			// 增加watcher
			wm.startWatcher(watcher)
		case propertySources := <-wm.refreshChan:
			// 参数刷新
			wm.refreshPropertySources(propertySources)
		// watcherManager超时轮询，用于watcher自检
		case <-time.After(60 * time.Second):
			// TODO 自检
			slog.Info("watcherManager check")
		case <-wm.ctx.Done():
			slog.Info("watcherManager ctx done")
			return
		case <-ctx.Done(): // 外部ctx关闭，关闭watcherManager
			slog.Info("parent ctx done")
			wm.close()
			return
		}
	}
}

func (wm *WatcherManager) close() {
	wm.closeOnce.Do(func() {
		slog.Info("watcherManager close")

		wm.watchersLock.Lock()
		defer func() {
			wm.cancel()
			wm.watchersLock.Unlock()
		}()

		for _, watcher := range wm.watchers {
			slog.Info(fmt.Sprintf("stop watcher %T", watcher))
			watcher.Stop()
		}
	})
}

func (wm *WatcherManager) startWatcher(w Watcher) {
	wm.watchersLock.Lock()
	defer wm.watchersLock.Unlock()

	// 如果wm已关闭，不启动watcher
	if wm.ctx.Err() != nil {
		return
	}

	wm.watchers = append(wm.watchers, w)
	go func() {
		// 启动watcher
		w.Start(wm.ctx, wm.changePropertySources)
	}()
}

func (wm *WatcherManager) changePropertySources(propertySources []IPropertySource) {
	wm.refreshChan <- propertySources
}

type WatcherServer struct {
	wm *WatcherManager
}

func NewWatcherServer(wm *WatcherManager) *WatcherServer {
	return &WatcherServer{
		wm: wm,
	}
}

func (ws *WatcherServer) Start(ctx context.Context) error {
	ws.wm.run(ctx)
	return nil
}

func (ws *WatcherServer) Stop(ctx context.Context) error {
	ws.wm.close()
	return nil
}
