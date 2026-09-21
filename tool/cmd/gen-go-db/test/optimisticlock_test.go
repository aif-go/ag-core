//go:build db

package test

// 乐观锁 DB 集成测试（方案 §五 用例；UpdateIgnore 与 UpdateByPrimaryKey 同等双方法矩阵）。
// 仅 tm_teacher 存在乐观锁列（jpa_version）；其余表保持现状无锁列。
// 断言一律回查 DB 实际值（version/列实际存储内容），不依赖返回实体回填。
//
// 行为校准（插件 v1.1.3 实测语义，方案 §五 case1 的"初始 0"假设需修正为：）
//   - Insert：锁列 Valid=false → 落库初始版本 1（VersionCreateClause 强制写 1）；
//     显式装载值则写入装载值（0 也是合法装载值）。
//   - Update：Valid=true 时插件追加 WHERE jpa_version = 装载值，SET 版本 = 版本+1；
//     Valid=false 时不追加版本 WHERE（静默失效）——由 LockCheck 显式拦截。

import (
	"context"
	"fmt"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"gorm.io/plugin/optimisticlock"
)

// ver 快捷构造"已装载版本"（区分 Valid=false 的未装载零值态）
func ver(v int64) optimisticlock.Version {
	return optimisticlock.Version{Int64: v, Valid: true}
}

// lockSeed 乐观锁用例自备数据（满足唯一约束列、长度约束）
func lockSeed(id int64) *model.TmTeacher {
	return &model.TmTeacher{
		Id:      id,
		Name:    fmt.Sprintf("Lock%d", id),
		Address: fmt.Sprintf("lock-addr-%d", id),
		Phone:   fmt.Sprintf("1393%06d", id),
		ClassId: "LK",
		CardNo:  fmt.Sprintf("CARD%d", id),
	}
}

// seedTeacher 清空 tm_teacher 并插入自备数据，返回清理函数
func seedTeacher(t *testing.T, ctx context.Context, rows ...*model.TmTeacher) func() {
	t.Helper()
	clearTable(t, "tm_teacher")
	teacherDao := GetRepository()
	for _, r := range rows {
		if _, err := teacherDao.InsertOne(ctx, r); err != nil {
			t.Fatalf("插入 tm_teacher 自备数据失败: %v", err)
		}
	}
	return func() { clearTable(t, "tm_teacher") }
}

// dbTeacherRow 回查 DB 实际值：(version, address, found)
func dbTeacherRow(t *testing.T, id int64) (version int64, address string, found bool) {
	t.Helper()
	var rows []struct {
		Version int64
		Address *string
	}
	if err := getTestDB().Table("tm_teacher").
		Select("jpa_version AS version, address").
		Where("id = ?", id).Scan(&rows).Error; err != nil {
		t.Fatalf("回查 tm_teacher 失败: %v", err)
	}
	if len(rows) == 0 {
		return 0, "", false
	}
	addr := ""
	if rows[0].Address != nil {
		addr = *rows[0].Address
	}
	return rows[0].Version, addr, true
}

func mustTeacherRow(t *testing.T, id int64) (int64, string) {
	t.Helper()
	version, address, found := dbTeacherRow(t, id)
	if !found {
		t.Fatalf("id=%d 行不存在（预期存在）", id)
	}
	return version, address
}

func assertVersionEq(t *testing.T, id int64, want int64) {
	t.Helper()
	got, _ := mustTeacherRow(t, id)
	if got != want {
		t.Fatalf("id=%d DB version=%d; want %d", id, got, want)
	}
}

// findTeacher 主键查询载入原实体（先查后改约定）
func findTeacher(t *testing.T, ctx context.Context, teacherDao dao.ITmTeacherDao, id int64) *model.TmTeacher {
	t.Helper()
	entity, err := teacherDao.FindByPrimaryKey(ctx, model.TmTeacherPrimaryKey(id))
	if err != nil {
		t.Fatalf("FindByPrimaryKey(%d) 失败: %v", id, err)
	}
	if entity == nil {
		t.Fatalf("FindByPrimaryKey(%d) 未找到", id)
	}
	return entity
}

// updateMethod 更新方法矩阵（工方案：UpdateIgnore 与 UpdateByPrimaryKey 同等测试）
type updateMethod struct {
	name   string
	update func(ctx context.Context, teacherDao dao.ITmTeacherDao, entity *model.TmTeacher) (int64, error)
}

func updateMatrix() []updateMethod {
	return []updateMethod{
		{name: "UpdateByPrimaryKey", update: func(ctx context.Context, teacherDao dao.ITmTeacherDao, entity *model.TmTeacher) (int64, error) {
			return teacherDao.UpdateByPrimaryKey(ctx, entity)
		}},
		{name: "UpdateByPrimaryKeyIgnoreZeroValCols", update: func(ctx context.Context, teacherDao dao.ITmTeacherDao, entity *model.TmTeacher) (int64, error) {
			return teacherDao.UpdateByPrimaryKeyIgnoreZeroValCols(ctx, entity)
		}},
	}
}

// 用例 1：插入含锁列实体 → 未装载锁列被插件写初始版本 1
func TestLock_InsertInitialVersion(t *testing.T) {
	ctx := context.Background()
	cleanup := seedTeacher(t, ctx, lockSeed(93001))
	defer cleanup()
	assertVersionEq(t, 93001, 1)
}

// 用例 1b：插入装载版本 0 → 库中即为 0（初始值语义：0 为合法历史版本）
func TestLock_InsertExplicitZeroVersion(t *testing.T) {
	ctx := context.Background()
	cleanup := seedTeacher(t, ctx)
	defer cleanup()

	entity := lockSeed(93002)
	entity.JpaVersion = ver(0)
	if _, err := GetRepository().InsertOne(ctx, entity); err != nil {
		t.Fatalf("插入失败: %v", err)
	}
	defer clearTable(t, "tm_teacher")
	assertVersionEq(t, 93002, 0)
}

// 用例 2：先查后改更新成功（两方法同等；含 version 自增与业务列更新）
func TestLock_先查后改更新成功(t *testing.T) {
	for _, method := range updateMatrix() {
		t.Run(method.name, func(t *testing.T) {
			ctx := context.Background()
			teacherDao := GetRepository()
			cleanup := seedTeacher(t, ctx, lockSeed(93101))
			defer cleanup()

			entity := findTeacher(t, ctx, teacherDao, 93101)
			entity.Address = fmt.Sprintf("updated-by-%s", method.name)
			// 装载版本保持查询原值：插件以装载值构造 WHERE，SET 版本 = 装载值+1

			affected, err := method.update(ctx, teacherDao, entity)
			if err != nil {
				t.Fatalf("更新失败: %v", err)
			}
			if affected != 1 {
				t.Fatalf("affected=%d; want 1", affected)
			}
			assertVersionEq(t, 93101, 2)
			_, address := mustTeacherRow(t, 93101)
			wantAddress := fmt.Sprintf("updated-by-%s", method.name)
			if address != wantAddress {
				t.Fatalf("address=%q; want %q", address, wantAddress)
			}
		})
	}
}

// 用例 3：并发冲突——两实体装载同一版本先后提交，第二次 0 行、不覆盖、版本仅 +1 一次
func TestLock_冲突第二次0行(t *testing.T) {
	for _, method := range updateMatrix() {
		t.Run(method.name, func(t *testing.T) {
			ctx := context.Background()
			teacherDao := GetRepository()
			cleanup := seedTeacher(t, ctx, lockSeed(93102))
			defer cleanup()

			first := findTeacher(t, ctx, teacherDao, 93102)
			reader1 := *first // 读者1（模拟先查）
			reader2 := *first // 后到者：同一版本
			reader1.Address = "late-write-1"
			reader2.Address = "late-write-2"

			// 第一次提交：成功
			n1, err := method.update(ctx, teacherDao, &reader1)
			if err != nil {
				t.Fatalf("第一次更新失败: %v", err)
			}
			if n1 != 1 {
				t.Fatalf("第一次 affected=%d; want 1", n1)
			}

			// 第二次提交：0 行 + 无错误 + 数据未覆盖 + 版本不回写
			n2, err := method.update(ctx, teacherDao, &reader2)
			if err != nil {
				t.Fatalf("冲突应为 nil err, 实际: %v", err)
			}
			if n2 != 0 {
				t.Fatalf("第二次 affected=%d; want 0（冲突）", n2)
			}
			version, address := mustTeacherRow(t, 93102)
			if version != 2 {
				t.Fatalf("DB version=%d; want 2（仅自增一次）", version)
			}
			if address != "late-write-1" {
				t.Fatalf("数据被第二次提交覆盖: address=%q", address)
			}
		})
	}
}

// 用例 4：不存在的主键更新 → 0 行 + 不落库（Save 退化 INSERT 缺陷已消除）
func TestLock_不存在主键不落库(t *testing.T) {
	for _, method := range updateMatrix() {
		t.Run(method.name, func(t *testing.T) {
			ctx := context.Background()
			teacherDao := GetRepository()
			cleanup := seedTeacher(t, ctx, lockSeed(93103))
			defer cleanup()

			entity := lockSeed(97777)
			entity.JpaVersion = ver(1) // 主键不存在且版本无效
			affected, err := method.update(ctx, teacherDao, entity)
			if err != nil && err.Error() == "" {
				t.Fatalf("预期 0 行 + nil err，实际: %v", err)
			}
			if affected != 0 {
				t.Fatalf("affected=%d; want 0", affected)
			}
			if _, _, found := dbTeacherRow(t, 97777); found {
				t.Fatal("Save 退化 INSERT 缺陷回归：不存在的 id 占据落库")
			}
		})
	}
}

// 用例 5：未装载版本（Valid=false）提交 → LockCheck 显式报错，数据不动（两方法同等）
func TestLock_未装载版本报错(t *testing.T) {
	for _, method := range updateMatrix() {
		t.Run(method.name, func(t *testing.T) {
			ctx := context.Background()
			teacherDao := GetRepository()
			cleanup := seedTeacher(t, ctx, lockSeed(93104))
			defer cleanup()

			beforeVersion, beforeAddress := mustTeacherRow(t, 93104)

			entity := lockSeed(93104)
			entity.Address = "should-not-write"
			_, err := method.update(ctx, teacherDao, entity)
			assertErrorContains(t, err, "optimistic lock version is required")

			version, address := mustTeacherRow(t, 93104)
			if version != beforeVersion || address != beforeAddress {
				t.Fatalf("报错路径不应写库: version=%d address=%q", version, address)
			}
		})
	}
}

// 用例 6：InsertOneIgnoreZeroValCols —— 锁列有效不进剔除名单（插入成功且锁列生效），普通零值列被剔
func TestLock_InsertIgnore_锁列不被剔除(t *testing.T) {
	ctx := context.Background()
	cleanup := seedTeacher(t, ctx)
	defer cleanup()

	entity := &model.TmTeacher{
		Id:      93200,
		Name:    "lock-ignore-insert",
		Phone:   "1393193200",
		ClassId: "LK",
		CardNo:  "CARD93200",
		// Address、Salary、JpaVersion 全部零值
	}
	if _, err := GetRepository().InsertOneIgnoreZeroValCols(ctx, entity); err != nil {
		t.Fatalf("InsertOneIgnoreZeroValCols 失败: %v", err)
	}
	defer clearTable(t, "tm_teacher")

	version, _, found := dbTeacherRow(t, 93200)
	if !found {
		t.Fatal("插入未落库")
	}
	if version != 1 {
		t.Fatalf("锁列不应被剔除（Valid=false 由插件生成初始值），实际 version=%d", version)
	}
}

// 用例 7：插入路径锁列 Valid=false 成功（无校验误报），初始版本 1
func TestLock_Insert_锁列零值成功(t *testing.T) {
	ctx := context.Background()
	cleanup := seedTeacher(t, ctx)
	defer cleanup()

	if _, err := GetRepository().InsertOne(ctx, lockSeed(93201)); err != nil {
		t.Fatalf("Insert 失败: %v", err)
	}
	defer clearTable(t, "tm_teacher")
	assertVersionEq(t, 93201, 1)
}

// 用例 8：UpdateIgnore 锁列 Valid=true + 若干零值列 → 零值列不进 SET、锁列自增
func TestLock_UpdateIgnore_零值剔除与锁列自增(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()
	cleanup := seedTeacher(t, ctx, lockSeed(93105))
	defer cleanup()

	first := findTeacher(t, ctx, teacherDao, 93105)
	beforeAddress := first.Address
	first.Address = "" // 置零值列（应被 Updates 排除出 SET，保持原值）
	first.Name = "partial-updated"
	// 装载版本保持查询原值：插件以装载值构造 WHERE，SET 版本 = 装载值+1

	affected, err := teacherDao.UpdateByPrimaryKeyIgnoreZeroValCols(ctx, first)
	if err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if affected != 1 {
		t.Fatalf("affected=%d; want 1", affected)
	}
	// 锁列照常自增（Version=2）
	assertVersionEq(t, 93105, 2)
	if _, address := mustTeacherRow(t, 93105); address != beforeAddress {
		t.Fatalf("零值列被意外覆盖: address=%q; want %q", address, beforeAddress)
	}
}

// 用例 9（UpdateIgnore）：锁列 Valid=false + 普通零值列 → LockCheck 先行报错（不被剔除逻辑掩盖）
func TestLock_UpdateIgnore_锁列未装载先行报错(t *testing.T) {
	ctx := context.Background()
	teacherDao := GetRepository()
	cleanup := seedTeacher(t, ctx, lockSeed(93106))
	defer cleanup()

	beforeVersion, beforeAddress := mustTeacherRow(t, 93106)

	entity := lockSeed(93106)
	entity.Address = "" // 零值列
	entity.Name = "should-not-write"
	_, err := teacherDao.UpdateByPrimaryKeyIgnoreZeroValCols(ctx, entity)
	assertErrorContains(t, err, "optimistic lock version is required")

	version, address := mustTeacherRow(t, 93106)
	if version != beforeVersion || address != beforeAddress {
		t.Fatalf("报错路径不应写库: version=%d address=%q", version, address)
	}
}

// 用例 10：无条件查询，直接构造完整实体提交更新（先查后改解耦）
func TestLock_直接构造实体提交(t *testing.T) {
	for _, method := range updateMatrix() {
		t.Run(method.name, func(t *testing.T) {
			ctx := context.Background()
			teacherDao := GetRepository()
			id := int64(93210)

			t.Run("装入当前版本直接提交成功", func(t *testing.T) {
				cleanup := seedTeacher(t, ctx, lockSeed(id))
				defer cleanup()
				seedVersion, _ := mustTeacherRow(t, id)

				entity := lockSeed(id)
				entity.Address = fmt.Sprintf("direct-%s", method.name)
				entity.CreateTime = otherTime // 全字段覆盖契约：直接构造实体必须装载非零时间列
				entity.JpaVersion = ver(seedVersion) // 直接构造实体携带库中当前版本

				affected, err := method.update(ctx, teacherDao, entity)
				if err != nil {
					t.Fatalf("失败: %v", err)
				}
				if affected != 1 {
					t.Fatalf("affected=%d; want 1", affected)
				}
				assertVersionEq(t, id, seedVersion+1)
			})

			t.Run("过期版本0行不覆盖", func(t *testing.T) {
				cleanup := seedTeacher(t, ctx, lockSeed(id+1))
				defer cleanup()
				seedVersion, _ := mustTeacherRow(t, id+1)

				entity := lockSeed(id + 1)
				entity.Address = "should-not-overwrite"
				entity.JpaVersion = ver(seedVersion + 100) // 过期版本

				affected, err := method.update(ctx, teacherDao, entity)
				if err != nil {
					t.Fatalf("过期版本应 nil err, 实际: %v", err)
				}
				if affected != 0 {
					t.Fatalf("affected=%d; want 0", affected)
				}
				assertVersionEq(t, id+1, seedVersion)
				if _, address := mustTeacherRow(t, id+1); address == "should-not-overwrite" {
					t.Fatal("过期版本提交仍覆盖了数据")
				}
			})
		})
	}
}

// 用例 4b：主键为零 → 两个更新方法均明确报错
func TestLock_主键为零报错(t *testing.T) {
	for _, method := range updateMatrix() {
		t.Run(method.name, func(t *testing.T) {
			ctx := context.Background()
			teacherDao := GetRepository()
			cleanup := seedTeacher(t, ctx, lockSeed(93300))
			defer cleanup()

			entity := lockSeed(93300)
			entity.Id = 0
			entity.JpaVersion = ver(1)
			_, err := method.update(ctx, teacherDao, entity)
			assertErrorContains(t, err, "primary key is required")
		})
	}
}

// 用例 12（运行时部分）：无锁列表的表更新行为保持不变（多主键形态回归）
func TestLock_无锁列表回归(t *testing.T) {
	ctx := context.Background()

	studentDao := GetStudentRepository()
	cleanup := seedTmStudent(t, ctx, studentDao)
	defer cleanup()

	entity, err := studentDao.FindByPrimaryKey(ctx, model.TmStudentPrimaryKey{TenantId: 1, StudentNo: "NO001"})
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	entity.Name = "regress-lock-free"
	affected, err := studentDao.UpdateByPrimaryKey(ctx, entity)
	if err != nil {
		t.Fatalf("无锁列表全字段更新失败: %v", err)
	}
	if affected != 1 {
		t.Fatalf("affected=%d; want 1", affected)
	}
}
