package compose

import "a.yandex-team.ru/strm/plgo/pkg/compose/builder"

type (
	GraphBuilder builder.GraphBuilder
	Graph        builder.Graph
)

var (
	NewGraphBuilder = builder.NewGraphBuilder
)
