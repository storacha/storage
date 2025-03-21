package service

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s, err := NewPDPService(ctx, common.Address{}, nil, nil)
	defer s.Stop(context.Background())
	require.NoError(t, err)

	<-ctx.Done()

}
