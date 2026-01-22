package agdao

import "context"

type TbInfoOpt func(ctx context.Context, info *TableInfo)

func WithTbNameStrategy(strategy func(ctx context.Context, info *TableInfo) string) TbInfoOpt {
	return func(ctx context.Context, info *TableInfo) {
		tbname := strategy(ctx, info)
		info.TableName = tbname
	}
}
