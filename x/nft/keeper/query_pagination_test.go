package keeper

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func TestNormalizeQueryPageRequest(t *testing.T) {
	testCases := []struct {
		name     string
		request  *query.PageRequest
		expected *query.PageRequest
	}{
		{
			name:     "nil request",
			expected: &query.PageRequest{Limit: 100},
		},
		{
			name:     "zero values",
			request:  &query.PageRequest{},
			expected: &query.PageRequest{Limit: 100},
		},
		{
			name:     "explicit limit",
			request:  &query.PageRequest{Limit: 25},
			expected: &query.PageRequest{Limit: 25},
		},
		{
			name:     "maximum limit",
			request:  &query.PageRequest{Limit: 100},
			expected: &query.PageRequest{Limit: 100},
		},
		{
			name: "cursor and reverse",
			request: &query.PageRequest{
				Key:     []byte("opaque-cursor"),
				Reverse: true,
			},
			expected: &query.PageRequest{
				Key:     []byte("opaque-cursor"),
				Limit:   100,
				Reverse: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var original *query.PageRequest
			if testCase.request != nil {
				copy := *testCase.request
				copy.Key = append([]byte(nil), testCase.request.Key...)
				original = &copy
			}

			normalized, err := normalizeQueryPageRequest(testCase.request)

			require.NoError(t, err)
			require.Equal(t, testCase.expected, normalized)
			if testCase.request != nil {
				require.NotSame(t, testCase.request, normalized)
				require.Equal(t, original, testCase.request)
			}
			if len(normalized.Key) > 0 {
				normalized.Key[0]++
				require.Equal(t, original.Key, testCase.request.Key)
			}
		})
	}
}

func TestNormalizeQueryPageRequestRejectsUnsupportedOptions(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		request *query.PageRequest
	}{
		{name: "limit above maximum", request: &query.PageRequest{Limit: 101}},
		{name: "non-zero offset", request: &query.PageRequest{Offset: 1}},
		{name: "total count", request: &query.PageRequest{CountTotal: true}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			normalized, err := normalizeQueryPageRequest(testCase.request)

			require.Nil(t, normalized)
			require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		})
	}
}
