package params_test

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestError_Error(t *testing.T) {

	e := &params.Error{
		Code:    200,
		Message: "OK",
	}

	require.Equal(t, "code: `200`, message: `OK`", e.Error())

}
