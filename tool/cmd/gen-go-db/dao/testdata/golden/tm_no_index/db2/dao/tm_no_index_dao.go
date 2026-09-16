package dao

import (
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"context"
	"reflect"
	"errors"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"

	agdao "github.com/aif-go/ag-core/contribute/agdb/agdao"
	

	"gorm.io/gorm"
)

// TmNoIndexDao tm_no_index DAO
// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT
type TmNoIndexDao struct {
	*gormdb.Repository
	info    agdao.TableInfo
	baseDao agdao.BaseDao
}

// ITmNoIndexDao TmNoIndex DAO接口
type ITmNoIndexDao interface {
	InsertOne(ctx context.Context, entity *model.TmNoIndex) (int64, error)
	InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.TmNoIndex) (int64, error)
	UpdateByPrimaryKey(ctx context.Context, entity *model.TmNoIndex) (int64, error)
	UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.TmNoIndex) (int64, error)
	FindByPrimaryKey(ctx context.Context, primaryKey model.TmNoIndexPrimarkey) (*model.TmNoIndex, error)
	FindByStruct(ctx context.Context, entity *model.TmNoIndex) ([]*model.TmNoIndex, error)
	FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error)
	FindByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder, page *gormdb.Page) ([]*model.TmNoIndex, *gormdb.PageResult, error)
	FindFirstOneByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder) (*model.TmNoIndex, error)
}

// NewTmNoIndexDao get dao instance
func NewTmNoIndexDao(repository *gormdb.Repository, baseDao agdao.BaseDao) ITmNoIndexDao {
	
	return &TmNoIndexDao{
		Repository: repository,
		baseDao:    baseDao,
		info: agdao.TableInfo{
			TableName: "tm_no_index",
		},
	}
}

// insertOne 插入一条数据库数据
func (dao *TmNoIndexDao) InsertOne(ctx context.Context, entity *model.TmNoIndex) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Create(entity)
	return result.RowsAffected, result.Error
}

// InsertOneIgnorenNullCols 插入数据时，自动剔除零值的列
func (dao *TmNoIndexDao) InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.TmNoIndex) (int64, error) {
	// 1. 剔除结构体中除主键和索引以及特殊列之外的零值列
	colnames,_,err:=entity.ListZeroValueCols(true, true, false, true)
	if err!= nil{
		return 0, err	
	}
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Omit(colnames...).Create(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKey 根据主键或者唯一键更新，该操作只适合从数据库查询原实体修改值之后使用
func (dao *TmNoIndexDao) UpdateByPrimaryKey(ctx context.Context, entity *model.TmNoIndex) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// 4. 更新条件（主键）
	where := make(map[string]any)
	// 检查主键是否为空，如果为空继续检查唯一键
	if ((entity.TenantId == 0) || (entity.StudentNo == "")) {
		return 0, errors.New("when update,primary key or unique key is required")
	} else {
		where["tenant_id"] = entity.TenantId
		where["student_no"] = entity.StudentNo
	}

	if len(where) == 0 {
		return 0, errors.New("when update,primary key or unique key is required")
	}
	// 5. 使用支持更新的列
	result := db.Model(&model.TmNoIndex{}).Where(where).Save(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKeyIngoreZeroValCols 根据主键或者唯一键更新，自动剔除参数中的零值列
func (dao *TmNoIndexDao) UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.TmNoIndex) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}	
	// 4. 更新条件（主键）
	where := make(map[string]any)
	// 检查主键是否为空，如果为空继续检查唯一键
	if ((entity.TenantId == 0) || (entity.StudentNo == "")) {
		return 0, errors.New("when update,primary key or unique key is required")
	} else {
		where["tenant_id"] = entity.TenantId
		where["student_no"] = entity.StudentNo
	}

	if len(where) == 0 {
		return 0, errors.New("when update,primary key or unique key is required")
	}
	// 使用支持更新的列
	result := db.Model(&model.TmNoIndex{}).Where(where).Updates(entity)
	return result.RowsAffected, result.Error
}

// FindByPrimaryKey 根据主键查询
func (dao *TmNoIndexDao) FindByPrimaryKey(ctx context.Context, primaryKey model.TmNoIndexPrimarkey) (*model.TmNoIndex, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}
	
	var entity model.TmNoIndex
	result := db.Where("tenant_id = ? AND student_no = ?", primaryKey.TenantId, primaryKey.StudentNo).First(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entity, result.Error
}

// FindByStruct 根据实体查询
func (dao *TmNoIndexDao) FindByStruct(ctx context.Context, entity *model.TmNoIndex) ([]*model.TmNoIndex, error) {
	var list []*model.TmNoIndex
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}

	// 检查是否使用了主键或索引，避免全表扫描
	keyUsed := false
	// 检查主键
	if entity.TenantId != 0 {
		keyUsed = true
	}
	if !keyUsed {
		return nil, errors.New("query not use any index")
	}

	// 全部非零列（含主键、索引列、特殊列）如果有值，也作为查询条件
	colnames, colvals, err := entity.ListZeroValueCols(false, false, true, false)
	if err != nil {
		return nil, err
	}
	if len(colnames) > 0 {
		for i, colname := range colnames {
			db = db.Where(colname+" = ?", colvals[i])
		}
	}

	// 执行查询
	result := db.Find(&list)
	return list, result.Error
}

// FindByCustomerRule 根据自定义规则查询
func (dao *TmNoIndexDao) FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error) {

	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	if namingInfo == nil {
		return nil, errors.New("namingInfo is nil")
	}

	if namingInfo.SqlName == "" {
		return nil, errors.New("namingInfo.SqlName is empty")
	}

	// 判断请求参数类型和实际类型是否一致
	reqType := reflect.TypeOf(namingInfo.ReqType)
	reqValue := reflect.ValueOf(args)
	if reqType != reqValue.Type() {
		return nil, errors.New("req type not match")
	}
	switch namingInfo.SqlName {
	default:
		return nil, errors.New("not found naming sql")
	}
}

// FindByCondition 根据条件构建器查询
func (dao *TmNoIndexDao) FindByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder, page *gormdb.Page) ([]*model.TmNoIndex, *gormdb.PageResult, error) {
	var list []*model.TmNoIndex
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 主动使用where条件
	where, args, err := condition.Build()
	if err != nil {
		return nil, nil, err
	}
	// 主动拼接where条件
	db = db.Where(where, args...)

	var totalCount int64
	// 统计总数
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}

	var pageResult *gormdb.PageResult
	// 如果需要分页
	if page != nil {
		start, _, totalPage, enablePage, err := gormdb.CalcPageStartRecord(page.PageNum, page.PageSize, totalCount, dao.DbType)
		if err != nil {
			return nil, nil, err
		}
		pageResult = &gormdb.PageResult{
			CurrentPage: page.PageNum,
			PageSize:    page.PageSize,
			TotalCount:  totalCount,
			TotalPage:   totalPage,
		}
		// 总记录数为0或者当前页码超过总页数时，不执行查询，直接返回空结果和分页信息
		if !enablePage {
			return nil, pageResult, nil
		}
		db = db.Limit(int(page.PageSize)).Offset(int(start))
	}

	// 主动拼排序条件
	if orderBuilder != nil {
		db = db.Order(orderBuilder.Build())
	}

	result := db.Find(&list)
	if result.Error != nil {
		return nil, nil, result.Error
	}

	return list, pageResult, nil
}

// FindFirstOneByCondition 根据条件构建器查询第一条记录
func (dao *TmNoIndexDao) FindFirstOneByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder) (*model.TmNoIndex, error) {
	var entity model.TmNoIndex
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}

	// 主动使用where条件
	where, args, err := condition.Build()
	if err != nil {
		return nil, err
	}
	// 主动拼接where条件
	db = db.Where(where, args...)

	// 主动拼排序条件
	if orderBuilder != nil {
		db = db.Order(orderBuilder.Build())
	}

	result := db.Limit(1).Find(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entity, result.Error
}



// getInfo 获取表信息
func (dao *TmNoIndexDao) getInfo() agdao.TableInfo {
	return dao.info
}

// getApplyInfo 获取应用表信息
func (dao *TmNoIndexDao) getApplyInfo(ctx context.Context) agdao.TableInfo {
	info := dao.getInfo()
	dao.baseDao.ApplyTbInfoOpts(ctx, &info)
	return info
}

// newDB 创建一个新的DB实例
func (dao *TmNoIndexDao) newDB(ctx context.Context) (*gorm.DB, error) {
	db := dao.DB(ctx)
	info := dao.getApplyInfo(ctx)
	tbname := info.TableName
	if tbname == "" {
		return nil, errors.New("表名不能为空")
	}

	db = db.Table(tbname)
	return db, nil
}




