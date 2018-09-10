package spectests

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"

	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/functions"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/influxql"
)

func init() {
	RegisterFixture(
		AggregateTest(func(aggregate flux.Operation) (stmt string, spec *flux.Spec) {
			return fmt.Sprintf(`SELECT %s(value) FROM db0..cpu GROUP BY host`, aggregate.Spec.Kind()),
				&flux.Spec{
					Operations: []*flux.Operation{
						{
							ID: "from0",
							Spec: &functions.FromOpSpec{
								BucketID: bucketID,
							},
						},
						{
							ID: "range0",
							Spec: &functions.RangeOpSpec{
								Start:    flux.Time{Absolute: time.Unix(0, influxql.MinTime)},
								Stop:     flux.Time{Absolute: time.Unix(0, influxql.MaxTime)},
								TimeCol:  execute.DefaultTimeColLabel,
								StartCol: execute.DefaultStartColLabel,
								StopCol:  execute.DefaultStopColLabel,
							},
						},
						{
							ID: "filter0",
							Spec: &functions.FilterOpSpec{
								Fn: &semantic.FunctionExpression{
									Params: []*semantic.FunctionParam{
										{Key: &semantic.Identifier{Name: "r"}},
									},
									Body: &semantic.LogicalExpression{
										Operator: ast.AndOperator,
										Left: &semantic.BinaryExpression{
											Operator: ast.EqualOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_measurement",
											},
											Right: &semantic.StringLiteral{
												Value: "cpu",
											},
										},
										Right: &semantic.BinaryExpression{
											Operator: ast.EqualOperator,
											Left: &semantic.MemberExpression{
												Object: &semantic.IdentifierExpression{
													Name: "r",
												},
												Property: "_field",
											},
											Right: &semantic.StringLiteral{
												Value: "value",
											},
										},
									},
								},
							},
						},
						{
							ID: "group0",
							Spec: &functions.GroupOpSpec{
								By: []string{"_measurement", "_start", "host"},
							},
						},
						&aggregate,
						{
							ID: "map0",
							Spec: &functions.MapOpSpec{
								Fn: &semantic.FunctionExpression{
									Params: []*semantic.FunctionParam{{
										Key: &semantic.Identifier{Name: "r"},
									}},
									Body: &semantic.ObjectExpression{
										Properties: []*semantic.Property{
											{
												Key: &semantic.Identifier{Name: "_time"},
												Value: &semantic.MemberExpression{
													Object: &semantic.IdentifierExpression{
														Name: "r",
													},
													Property: "_time",
												},
											},
											{
												Key: &semantic.Identifier{Name: string(aggregate.Spec.Kind())},
												Value: &semantic.MemberExpression{
													Object: &semantic.IdentifierExpression{
														Name: "r",
													},
													Property: "_value",
												},
											},
										},
									},
								},
								MergeKey: true,
							},
						},
						{
							ID: "yield0",
							Spec: &functions.YieldOpSpec{
								Name: "0",
							},
						},
					},
					Edges: []flux.Edge{
						{Parent: "from0", Child: "range0"},
						{Parent: "range0", Child: "filter0"},
						{Parent: "filter0", Child: "group0"},
						{Parent: "group0", Child: aggregate.ID},
						{Parent: aggregate.ID, Child: "map0"},
						{Parent: "map0", Child: "yield0"},
					},
					Now: Now(),
				}
		}),
	)
}