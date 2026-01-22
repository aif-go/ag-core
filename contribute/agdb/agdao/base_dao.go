package agdao

import "context"

// BaseDao 基础Dao的基础增强能力
type BaseDao interface {
	ApplyTbInfoOpts(ctx context.Context, info *TableInfo)
}

type baseDao struct {
	tbInfoOpts []TbInfoOpt
}

func (dao *baseDao) ApplyTbInfoOpts(ctx context.Context, info *TableInfo) {
	for _, opt := range dao.tbInfoOpts {
		opt(ctx, info)
	}
}

func (dao *baseDao) RegTbInfoOpt(opts ...TbInfoOpt) {
	dao.tbInfoOpts = append(dao.tbInfoOpts, opts...)
}
