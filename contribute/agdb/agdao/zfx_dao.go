package agdao

import "go.uber.org/fx"

type FxInBaseDao struct {
	fx.In

	TbInfoOpts []TbInfoOpt `group:"fx_ag_gorm_tbinfo_opt",optional:"true"`
}

// NewFxAgTbInfoOpt 创建fx表信息选项提供者
func NewFxAgTbInfoOpt(t any) any {
	return fx.Annotate(
		t,
		fx.ResultTags(`group:"fx_ag_gorm_tbinfo_opt"`),
	)
}

// FxNewAgServiceBuilder fx构建AgServiceBuilder
func FxNewAgGormBaseDao(in FxInBaseDao) (BaseDao, error) {
	bdao := &baseDao{}
	bdao.RegTbInfoOpt(in.TbInfoOpts...)
	return bdao, nil
}
