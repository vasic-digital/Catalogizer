package challenges

import (
	"testing"

	"catalogizer/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAll(t *testing.T) {
	svc := services.NewChallengeService(t.TempDir())
	err := RegisterAll(svc)
	require.NoError(t, err)

	// Currently no challenges registered, but the call
	// should succeed without error.
	challenges := svc.ListChallenges()
	assert.NotNil(t, challenges)
}
