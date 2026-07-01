package agbatch

import "go.uber.org/fx"

// FxAgBatchModule provides the agbatch framework components for fx DI.
//
// Usage:
//
//	app := fx.New(
//	    agbatch.FxAgBatchModule,
//	    // Provide your jobs:
//	    fx.Provide(NewMyJob),
//	)
//
// Where NewMyJob returns a *Job or uses the builder.
var FxAgBatchModule = fx.Module(
	"fx_agbatch_module",
	fx.Provide(
		// Default in-memory repository
		func() JobRepository { return NewInMemoryRepository() },
		// Job launcher
		NewJobLauncher,
	),
)
