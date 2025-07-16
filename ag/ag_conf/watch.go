package ag_conf

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

var WatcherM *WatcherManager
var global_watcherChan = make(chan Watcher, 1024)

// RegisterWatcher 注册watcher
func RegisterWatcher(w Watcher) {
	global_watcherChan <- w
}

type Watcher interface {
	Start(context.Context, chan []IPropertySource)
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

func (wm *WatcherManager) AddConfigChangeListener(key string, listener func(k, v string)) {
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
	for {
		select {
		case watcher := <-wm.watcherChan:
			// 增加watcher
			wm.startWatcher(watcher)
		case propertySources := <-wm.refreshChan:
			// 参数刷新
			wm.refreshPropertySources(propertySources)
		// watcherManager超时轮询，用于watcher自检
		case <-time.After(3 * time.Second):
			// TODO 自检
			slog.Info("watcherManager check")
		case <-wm.ctx.Done():
			return
		case <-ctx.Done(): // 外部ctx关闭，关闭watcherManager
			wm.close()
			return
		}
	}
}

func (wm *WatcherManager) close() {
	wm.closeOnce.Do(func() {

		wm.watchersLock.Lock()
		defer func() {
			wm.cancel()
			wm.watchersLock.Unlock()
		}()

		for _, watcher := range wm.watchers {
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
		w.Start(wm.ctx, wm.refreshChan)
	}()
}

func (wm *WatcherManager) refreshPropertySources(propertySources []IPropertySource) {
	// TODO 参数刷新处理
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
