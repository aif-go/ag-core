// Package usage 演示 ag_cache 的完整实践用法（贴近真实业务的可读示例）。
// 与 test/ 根目录的测试工具不同，这里是"业务怎么用"的完整案例：
//   - UserService：用户缓存（Cache-Aside，绑定 loader 读穿透）
//   - ParamService：参数缓存（Clear 批量失效）
//   - 监控探活：TryGet 预期 Hit
//
// 业务代码零框架概念：构造时注入 *ag_cache.Manager，用 GetCacheWithLoader 绑定一次，
// 方法内直接复用缓存句柄。
package usage

import (
	"context"

	ag_cache "github.com/aif-go/ag-core/ag/ag_cache"
)

// User 用户实体。
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserRepo 模拟用户数据源（DB）。
type UserRepo struct {
	DB map[string]*User
}

// GetUser 是 UserService 的 loader（方法引用直接可传，无需适配代码）。
func (r *UserRepo) GetUser(ctx context.Context, id string) (*User, error) {
	u, ok := r.DB[id]
	if !ok {
		return nil, ag_cache.ErrCacheMiss
	}
	return u, nil
}

// UserService 用户缓存业务（Cache-Aside）。
type UserService struct {
	repo  *UserRepo
	users *ag_cache.LoaderCache[*User]
}

// NewUserService 构造业务服务：注入 *Manager，构造时用 GetCacheWithLoader 绑定缓存（读穿透）。
func NewUserService(m *ag_cache.Manager, repo *UserRepo) *UserService {
	return &UserService{
		repo:  repo,
		users: ag_cache.GetCacheWithLoader(m, "users", repo.GetUser),
	}
}

// GetUser 读缓存：miss → loader（repo.GetUser）→ 写缓存 → 返回（读穿透）。
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
	return s.users.Get(ctx, id)
}

// RefreshUser 用户变更 → 失效缓存（下次 Get 自动重载新值）。
func (s *UserService) RefreshUser(ctx context.Context, id string) error {
	return s.users.Del(ctx, id)
}

// ProbeUser 监控探活：TryGet 预期 Hit，不触发 loader、不产生 miss 异常。
func (s *UserService) ProbeUser(ctx context.Context, id string) (string, bool) {
	v, ok, _ := s.users.TryGet(ctx, id)
	if !ok {
		return "", false
	}
	return v.Name, true
}

// Param 参数实体。
type Param struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ParamCenter 参数中心数据源。
type ParamCenter struct {
	Values map[string]string
}

// Get 是 ParamService 的 loader。
func (p *ParamCenter) Get(ctx context.Context, key string) (*Param, error) {
	v, ok := p.Values[key]
	if !ok {
		return nil, ag_cache.ErrCacheMiss
	}
	return &Param{Key: key, Value: v}, nil
}

// ParamService 参数缓存业务（批量失效）。
type ParamService struct {
	center *ParamCenter
	params *ag_cache.LoaderCache[*Param]
}

// NewParamService 构造参数服务：注入 *Manager，构造时绑定缓存。
func NewParamService(m *ag_cache.Manager, center *ParamCenter) *ParamService {
	return &ParamService{
		center: center,
		params: ag_cache.GetCacheWithLoader(m, "params", center.Get),
	}
}

// GetParam 读参数（读穿透）。
func (s *ParamService) GetParam(ctx context.Context, key string) (*Param, error) {
	return s.params.Get(ctx, key)
}

// BroadcastUpdate 参数系统更新广播 → 清空 params 缓存（独立实例，不影响 users）。
func (s *ParamService) BroadcastUpdate(ctx context.Context) error {
	return s.params.Clear(ctx)
}
