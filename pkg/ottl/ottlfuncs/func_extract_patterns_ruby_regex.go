// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/pdata/pcommon"

	oniguruma "github.com/go-enry/go-oniguruma"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

type ExtractPatternsRubyRegexArguments[K any] struct {
	Target  ottl.StringGetter[K]
	Pattern string
}

func NewExtractPatternsRubyRegexFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("ExtractPatternsRubyRegex", &ExtractPatternsRubyRegexArguments[K]{}, createExtractPatternsRubyRegexFunction[K])
}

func createExtractPatternsRubyRegexFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*ExtractPatternsRubyRegexArguments[K])

	if !ok {
		return nil, fmt.Errorf("ExtractPatternsRubyRegexFactory args must be of type *ExtractPatternsRubyRegexArguments[K]")
	}

	return extractPatternsRubyRegex(args.Target, args.Pattern)
}

func extractPatternsRubyRegex[K any](target ottl.StringGetter[K], pattern string) (ottl.ExprFunc[K], error) {
	r, err := oniguruma.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("the pattern supplied to ExtractPatternsRubyRegex is not a valid pattern: %w", err)
	}

	namedCaptureGroups := 0
	for _, groupName := range r.getNamedGroupInfo() {
		if groupName != "" {
			namedCaptureGroups++
		}
	}

	if namedCaptureGroups == 0 {
		return nil, fmt.Errorf("at least 1 named capture group must be supplied in the given regex")
	}

	return func(ctx context.Context, tCtx K) (any, error) {
		val, err := target.Get(ctx, tCtx)
		if err != nil {
			return nil, err
		}

		matches := r.FindStringSubmatch(val)
		if matches == nil {
			return pcommon.NewMap(), nil
		}

		result := pcommon.NewMap()
		for i, subexp := range r.SubexpNames() {
			if i == 0 {
				// Skip whole match
				continue
			}
			if subexp != "" {
				result.PutStr(subexp, matches[i])
			}
		}
		return result, err
	}, nil
}
