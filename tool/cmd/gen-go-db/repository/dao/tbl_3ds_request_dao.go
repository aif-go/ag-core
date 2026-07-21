package dao

import (
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"context"
	"reflect"
	"errors"
	"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"

	agdao "github.com/aif-go/ag-core/contribute/agdb/agdao"
	"strings"

	"gorm.io/gorm"
)

// Tbl3dsRequestDao tbl_3ds_request DAO
// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT
type Tbl3dsRequestDao struct {
	*gormdb.Repository
	info    agdao.TableInfo
	baseDao agdao.BaseDao
}

// ITbl3dsRequestDao Tbl3dsRequest DAO接口
type ITbl3dsRequestDao interface {
	InsertOne(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error)
	InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error)
	UpdateByPrimaryKey(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error)
	UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error)
	FindByPrimaryKey(ctx context.Context, primaryKey model.Tbl3dsRequestPrimarkey) (*model.Tbl3dsRequest, error)
	FindByStruct(ctx context.Context, entity *model.Tbl3dsRequest) ([]*model.Tbl3dsRequest, error)
	FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error)
	FindByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder, page *gormdb.Page) ([]*model.Tbl3dsRequest, *gormdb.PageResult, error)
	FindFirstOneByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder) (*model.Tbl3dsRequest, error)
}

// NewTbl3dsRequestDao get dao instance
func NewTbl3dsRequestDao(repository *gormdb.Repository, baseDao agdao.BaseDao) ITbl3dsRequestDao {
	InitTbl3dsRequestNamingSql()
	return &Tbl3dsRequestDao{
		Repository: repository,
		baseDao:    baseDao,
		info: agdao.TableInfo{
			TableName: "tbl_3ds_request",
		},
	}
}

// insertOne 插入一条数据库数据
func (dao *Tbl3dsRequestDao) InsertOne(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	result := db.Create(entity)
	return result.RowsAffected, result.Error
}

// InsertOneIgnorenNullCols 插入数据时，自动剔除零值的列
func (dao *Tbl3dsRequestDao) InsertOneIgnoreZeroValCols(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error) {
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
func (dao *Tbl3dsRequestDao) UpdateByPrimaryKey(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}

	// 4. 更新条件（主键）
	where := make(map[string]any)
	// 检查主键是否为空，如果为空继续检查唯一键
	if ((entity.Id == 0) || (entity.ClusterId == "")) {
		return 0, errors.New("when update,primary key or unique key is required")
	} else {
		where["ID"] = entity.Id
		where["CLUSTER_ID"] = entity.ClusterId
	}

	if len(where) == 0 {
		return 0, errors.New("when update,primary key or unique key is required")
	}
	// 5. 使用支持更新的列
	result := db.Model(&model.Tbl3dsRequest{}).Where(where).Save(entity)
	return result.RowsAffected, result.Error
}

// UpdateByPrimaryKeyIngoreZeroValCols 根据主键或者唯一键更新，自动剔除参数中的零值列
func (dao *Tbl3dsRequestDao) UpdateByPrimaryKeyIngoreZeroValCols(ctx context.Context, entity *model.Tbl3dsRequest) (int64, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return 0, err
	}	
	// 4. 更新条件（主键）
	where := make(map[string]any)
	// 检查主键是否为空，如果为空继续检查唯一键
	if ((entity.Id == 0) || (entity.ClusterId == "")) {
		return 0, errors.New("when update,primary key or unique key is required")
	} else {
		where["ID"] = entity.Id
		where["CLUSTER_ID"] = entity.ClusterId
	}

	if len(where) == 0 {
		return 0, errors.New("when update,primary key or unique key is required")
	}
	// 使用支持更新的列
	result := db.Model(&model.Tbl3dsRequest{}).Where(where).Updates(entity)
	return result.RowsAffected, result.Error
}

// FindByPrimaryKey 根据主键查询
func (dao *Tbl3dsRequestDao) FindByPrimaryKey(ctx context.Context, primaryKey model.Tbl3dsRequestPrimarkey) (*model.Tbl3dsRequest, error) {
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}
	
	var entity model.Tbl3dsRequest
	result := db.Where("ID = ? AND CLUSTER_ID = ?", primaryKey.Id, primaryKey.ClusterId).First(&entity)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &entity, result.Error
}

// FindByStruct 根据实体查询
func (dao *Tbl3dsRequestDao) FindByStruct(ctx context.Context, entity *model.Tbl3dsRequest) ([]*model.Tbl3dsRequest, error) {
	var list []*model.Tbl3dsRequest
	db, err := dao.newDB(ctx)
	if err != nil {
		return nil, err
	}

	// 检查主键是否为空
	if entity.Id != 0 {
		db = db.Where("ID = ?", entity.Id)
		if entity.ClusterId != "" {
			db = db.Where("CLUSTER_ID = ?", entity.ClusterId)
		}
		result := db.Find(&list)
		return list, result.Error
	}

	// 检查索引列，确保使用了索引
	indexUsed := false
	// 检查索引 IDX_INSERT_TIMESTAMP
	if !entity.InsertTimestamp.IsZero() {
		db = db.Where("INSERT_TIMESTAMP = ?", entity.InsertTimestamp)
		indexUsed = true
	}
	// 检查索引 IDX_ORDER_ID
	if entity.OrderId != "" {
		db = db.Where("ORDER_ID = ?", entity.OrderId)
		if entity.MerchantId != "" {
			db = db.Where("MERCHANT_ID = ?", entity.MerchantId)
		}
		indexUsed = true
	}
	// 检查索引 IDX_RRN
	if entity.RetrievalReferenceNumber != "" {
		db = db.Where("RETRIEVAL_REFERENCE_NUMBER = ?", entity.RetrievalReferenceNumber)
		if entity.MerchantId != "" {
			db = db.Where("MERCHANT_ID = ?", entity.MerchantId)
		}
		if entity.TransactionType != "" {
			db = db.Where("TRANSACTION_TYPE = ?", entity.TransactionType)
		}
		indexUsed = true
	}

	// 检查索引列，确保使用了索引
	if !indexUsed {
		return nil, errors.New("query not use any index")
	}

	// 除了主键和索引以外的其他列如果有值，也作为查询条件
	colnames, colvals, err := entity.ListZeroValueCols(true, true, true, false)
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
func (dao *Tbl3dsRequestDao) FindByCustomerRule(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) (any, error) {

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
	case "Xxxxx":
		return dao.doXxxxx(ctx, namingInfo, args)
	default:
		return nil, errors.New("not found naming sql")
	}
}

// FindByCondition 根据条件构建器查询
func (dao *Tbl3dsRequestDao) FindByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder, page *gormdb.Page) ([]*model.Tbl3dsRequest, *gormdb.PageResult, error) {
	var list []*model.Tbl3dsRequest
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
func (dao *Tbl3dsRequestDao) FindFirstOneByCondition(ctx context.Context, condition *conditonwhere.WhereClauseBuilder, orderBuilder *gormdb.OrderBuilder) (*model.Tbl3dsRequest, error) {
	var entity model.Tbl3dsRequest
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

// doXxxxx 执行Xxxxx查询（非分页）
func (dao *Tbl3dsRequestDao) doXxxxx(ctx context.Context, namingInfo *gormdb.NameingSqlArgInfo, args any) ([]*model.Tbl3dsRequest, error) {

	queryArgs, ok := args.(*model.Tbl3dsRequestXxxxxArg)
	if !ok {
		return nil, errors.New("doXxxxx args type not match")
	}

	sqlName := dao.DbType + "_" + "Tbl3dsRequest" + "_" + namingInfo.SqlName
	execSql := Tbl3dsRequestNamingSqlMap[sqlName]
	if execSql == "" {
		return nil, errors.New("not found naming sql")
	}

	newwhere,err:=queryArgs.FieldMask.BuildWhereFromConfig("Xxxxx",model.Tbl3dsRequestConditionMap)
	if err != nil {
		return nil, err
	}
	// 校验新的where条件是否使用了索引列，避免全表扫描
	check:=conditonwhere.ValidateLeadingCol(newwhere,model.Tbl3dsRequestIndexLeadingCols)
	if !check {
		return nil, errors.New("query not use any index")
	}

	newTableName := dao.getApplyInfo(ctx).TableName
	if newTableName != "" {
		enity := &model.Tbl3dsRequest{}
		execSql = strings.ReplaceAll(execSql, "FROM "+enity.TableName()+" WHERE", "FROM "+newTableName+" WHERE")
	}

	argsMap := queryArgs.ConvertToMap()
	var list []*model.Tbl3dsRequest
	result := dao.DB(ctx).Raw(execSql, argsMap).Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}
	return list, nil
}



// getInfo 获取表信息
func (dao *Tbl3dsRequestDao) getInfo() agdao.TableInfo {
	return dao.info
}

// getApplyInfo 获取应用表信息
func (dao *Tbl3dsRequestDao) getApplyInfo(ctx context.Context) agdao.TableInfo {
	info := dao.getInfo()
	dao.baseDao.ApplyTbInfoOpts(ctx, &info)
	return info
}

// newDB 创建一个新的DB实例
func (dao *Tbl3dsRequestDao) newDB(ctx context.Context) (*gorm.DB, error) {
	db := dao.DB(ctx)
	info := dao.getApplyInfo(ctx)
	tbname := info.TableName
	if tbname == "" {
		return nil, errors.New("表名不能为空")
	}

	db = db.Table(tbname)
	return db, nil
}




