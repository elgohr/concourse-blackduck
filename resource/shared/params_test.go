package shared_test

import (
	"github.com/elgohr/concourse-blackduck/shared"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsValidWhenDirectoryIsFilled(t *testing.T) {
	p := shared.Params{
		Directory:"directory",
	}
	require.True(t, p.Valid())
}

func TestIsInvalidWhenDirectoryIsNotFilled(t *testing.T) {
	p := shared.Params{}
	require.False(t, p.Valid())
}
