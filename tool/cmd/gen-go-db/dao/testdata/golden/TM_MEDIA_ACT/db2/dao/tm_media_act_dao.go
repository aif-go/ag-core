package dao

import (
	"context"
	"errors"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"reflect"

	agdao "github.com/aif-go/ag-core/contribute/agdb/agdao"
	"strings"

	"gorm.io/gorm"
)

// TmMediaActDao TM_MEDIA_ACT DAO
// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT
type TmMediaActDao struct {
	*gormdb.Repository
	info    agdao.TableInfo
	baseDao agdao.BaseDao
}

// ITmMediaActDao TmMediaAct DAO接口
type ITmMediaActDao interface {
	InsertOne(ctx context.Context, entity *model.TmMediaAct) (int64, error)
	InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.TmMediaAct) (int64, error)
	UpdateByPrimaryKey(ctx context.Context, entity *model.TmMediaAct) (int64, error)
	UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.TmMediaAct) (int64, error) // Deprecated: 历史拼写，使用 UpdateByPrimaryKeyIgnoreZeroValCols（永久并存，无移除计划）
	UpdateByPrimaryKeyIgnoreZeroValCols(ctx context.Context, entity *model.TmMediaAct) (int64, error)
	FindByPrimaryKey(ctx context.Context, id model.TmMediaActPrimaryKey) (*model.TmMediaAct, error)
	FindByStruct(ctx context.Context, entity *model.TmMediaAct) ([]*model.TmMediaAct, error)
	FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error)
	FindByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder, page *gormdb.Page) ([]*model.TmMediaAct, *gormdb.PageResult, error)
	FindFirstOneByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder) (*model.TmMediaAct, error)
}

// NewTmMediaActDao get dao instance
func NewTmMediaActDao(repository *gormdb.Repository, baseDao agdao.BaseDao) ITmMediaActDao {
	InitTmMediaActNamingSql()
	return &TmMediaActDao{
		Repository: repository,
		baseDao:    baseDao,
		info: agdao.TableInfo{
			TableName: "TM_MEDIA_ACT",
		},
	}
}

// insertOne 插入一条数据库数据
func (dao *TmMediaActDao) InsertOne(ctx context.Context, entity *model.TmMediaAct) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Create(entity)
	return result.RowsAffected, result.Error
}

// InsertOneIgnoreZeroValCols 插入数据时，自动剔除零值的列
func (dao *TmMediaActDao) InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.TmMediaAct) (int64, error) {
	// 1. 剔除结构体中除主键和索引以及特殊列之外的零值列
	colnames, _, err := entity.ListZeroValueCols(true, true, false, true)
	if err != nil {
		return 0, err
	}
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Omit(colnames...).Create(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKey 根据主键或者唯一键更新，全字段覆盖更新，该操作只适合从数据库查询原实体修改值之后使用
// 直接构造实体提交同样支持，但必须装载乐观锁版本字段（如有）
func (dao *TmMediaActDao) UpdateByPrimaryKey(ctx context.Context, entity *model.TmMediaAct) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// 检查主键是否为空
	if entity.Seq == 0 {
		return 0, errors.New("when update,primary key is required")
	}
	// 5. 全字段更新，gorm 以实体主键为 WHERE 条件
	result := db.Model(entity).Select("*").Updates(entity)

	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKeyIgnoreZeroValCols 根据主键或者唯一键更新，自动忽略零值列（乐观锁版本列除外）
func (dao *TmMediaActDao) UpdateByPrimaryKeyIgnoreZeroValCols(ctx context.Context, entity *model.TmMediaAct) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// 检查主键是否为空
	if entity.Seq == 0 {
		return 0, errors.New("when update,primary key is required")
	}
	// 使用支持更新的列
	result := db.Model(entity).Updates(entity)

	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKeyIngoreZeroValCols 根据主键或者唯一键更新，自动剔除参数中的零值列
// Deprecated: 历史拼写，使用 UpdateByPrimaryKeyIgnoreZeroValCols（两者永久并存，无移除计划）
func (dao *TmMediaActDao) UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.TmMediaAct) (int64, error) {
	return dao.UpdateByPrimaryKeyIgnoreZeroValCols(ctx, entity)
}

// FindByPrimaryKey 根据主键查询
func (dao *TmMediaActDao) FindByPrimaryKey(ctx context.Context, id model.TmMediaActPrimaryKey) (*model.TmMediaAct, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}

	// 检查主键是否为空
	if id == 0 {
		return nil, errors.New("when query,primary key is required")
	}

	var entity model.TmMediaAct
	result := db.Where("SEQ = ?", id).First(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entity, result.Error
}

// FindByStruct 根据实体查询
func (dao *TmMediaActDao) FindByStruct(ctx context.Context, entity *model.TmMediaAct) ([]*model.TmMediaAct, error) {
	var list []*model.TmMediaAct
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}

	// 检查是否使用了主键或索引，避免全表扫描
	keyUsed := false
	// 检查主键
	if entity.Seq != 0 {
		keyUsed = true
	}
	// 检查索引 INDEX1_TM_MEDOA_ACT
	if !entity.BizDate.IsZero() {
		keyUsed = true
	}
	// 检查索引 INDEX2_TM_MEDIA_ACT
	if entity.Address != "" {
		keyUsed = true
	}
	// 检查索引 TM_MEDIA_ACT_UNIUQE_1
	if entity.Cardno != "" {
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
func (dao *TmMediaActDao) FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error) {

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
	case "NoPageQuery":
		return dao.doNoPageQuery(ctx, namingInfo, args)
	case "Xxxxx":
		return dao.doXxxxx(ctx, namingInfo, args)
	default:
		return nil, errors.New("not found naming sql")
	}
}

// FindByCondition 根据条件构建器查询
func (dao *TmMediaActDao) FindByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder, page *gormdb.Page) ([]*model.TmMediaAct, *gormdb.PageResult, error) {
	var list []*model.TmMediaAct
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
func (dao *TmMediaActDao) FindFirstOneByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder) (*model.TmMediaAct, error) {
	var entity model.TmMediaAct
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

// doNoPageQuery 执行NoPageQuery查询（非分页）
func (dao *TmMediaActDao) doNoPageQuery(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) ([]*model.TmMediaActNoPageQueryRes, error) {

	queryArgs, ok := args.(*model.TmMediaActNoPageQueryArg)
	if !ok {
		return nil, errors.New("doNoPageQuery args type not match")
	}

	sqlName := dao.DbType + "_" + "TmMediaAct" + "_" + namingInfo.SqlName
	execSql := TmMediaActNamingSqlMap[sqlName]
	if execSql == "" {
		return nil, errors.New("not found naming sql")
	}

	newwhere, err := queryArgs.FieldMask.BuildWhereFromConfig("NoPageQuery", model.TmMediaActConditionMap)
	if err != nil {
		return nil, err
	}
	// 校验新的where条件是否使用了索引列，避免全表扫描
	check := conditonwhere.ValidateLeadingCol(newwhere, model.TmMediaActIndexLeadingCols)
	if !check {
		return nil, errors.New("query not use any index")
	}

	newTableName := dao.getApplyInfo(ctx).TableName
	if newTableName != "" {
		entity := &model.TmMediaAct{}
		execSql = strings.ReplaceAll(execSql, "FROM "+entity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
	}

	argsMap := queryArgs.ConvertToMap()
	var list []*model.TmMediaActNoPageQueryRes
	result := dao.DB(ctx).Raw(execSql, argsMap).Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}
	return list, nil
}

// doXxxxx 执行Xxxxx查询（分页）
func (dao *TmMediaActDao) doXxxxx(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (*model.TmMediaActXxxxxPageRes, error) {

	queryArgs, ok := args.(*model.TmMediaActXxxxxArg)
	if !ok {
		return nil, errors.New("doXxxxx args type not match")
	}

	sqlName := dao.DbType + "_" + "TmMediaAct" + "_" + namingInfo.SqlName
	execSql := TmMediaActNamingSqlMap[sqlName]
	if execSql == "" {
		return nil, errors.New("not found naming sql")
	}
	newwhere, err := queryArgs.FieldMask.BuildWhereFromConfig("Xxxxx", model.TmMediaActConditionMap)
	if err != nil {
		return nil, err
	}
	// 校验新的where条件是否使用了索引列，避免全表扫描
	check := conditonwhere.ValidateLeadingCol(newwhere, model.TmMediaActIndexLeadingCols)
	if !check {
		return nil, errors.New("query not use any index")
	}

	execCountSql := TmMediaActNamingSqlMap[sqlName+"_Count"]
	if execCountSql == "" {
		return nil, errors.New("not found naming sql count")
	}

	newTableName := dao.getApplyInfo(ctx).TableName
	if newTableName != "" {
		entity := &model.TmMediaAct{}
		execSql = strings.ReplaceAll(execSql, "FROM "+entity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
		execCountSql = strings.ReplaceAll(execCountSql, "FROM "+entity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
	}

	argsMap := queryArgs.ConvertToMap()
	var totalCount int64
	result := dao.DB(ctx).Raw(execCountSql, argsMap).Scan(&totalCount)
	if result.Error != nil {
		return nil, result.Error
	}
	startRecord, endRecord, totalPage, enablePage, err := gormdb.CalcPageStartRecord(queryArgs.PageNum, queryArgs.PageSize, totalCount, dao.DbType)
	if err != nil {
		return nil, err
	}
	if !enablePage {
		return &model.TmMediaActXxxxxPageRes{
			PageResult: gormdb.PageResult{
				CurrentPage: queryArgs.PageNum,
				PageSize:    queryArgs.PageSize,
				TotalCount:  totalCount,
				TotalPage:   totalPage,
			},
		}, nil
	}
	argsMap["Start"] = startRecord
	argsMap["End"] = endRecord
	var list []*model.TmMediaActXxxxxRes
	resultlist := dao.DB(ctx).Raw(execSql, argsMap).Find(&list)
	if resultlist.Error != nil {
		return nil, resultlist.Error
	}

	return &model.TmMediaActXxxxxPageRes{
		PageResult: gormdb.PageResult{
			CurrentPage: queryArgs.PageNum,
			PageSize:    queryArgs.PageSize,
			TotalCount:  totalCount,
			TotalPage:   totalPage,
		},
		ResultList: list,
	}, nil
}

// getInfo 获取表信息
func (dao *TmMediaActDao) getInfo() agdao.TableInfo {
	return dao.info
}

// getApplyInfo 获取应用表信息
func (dao *TmMediaActDao) getApplyInfo(ctx context.Context) agdao.TableInfo {
	info := dao.getInfo()
	dao.baseDao.ApplyTbInfoOpts(ctx, &info)
	return info
}

// newDB 创建一个新的DB实例
func (dao *TmMediaActDao) newDB(ctx context.Context) (*gorm.DB, error) {
	db := dao.DB(ctx)
	info := dao.getApplyInfo(ctx)
	tbname := info.TableName
	if tbname == "" {
		return nil, errors.New("表名不能为空")
	}

	db = db.Table(tbname)
	return db, nil
}
