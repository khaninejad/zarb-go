package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zarbchain/zarb-go/errors"
)

func TestQueryProposalType(t *testing.T) {
	m := &QueryProposalMessage{}
	assert.Equal(t, m.Type(), MessageTypeQueryProposal)
}

func TestQueryProposalMessage(t *testing.T) {
	t.Run("Invalid height", func(t *testing.T) {
		m := NewQueryProposalMessage(-1, 0)

		assert.Equal(t, errors.Code(m.SanityCheck()), errors.ErrInvalidHeight)
	})

	t.Run("Invalid round", func(t *testing.T) {
		m := NewQueryProposalMessage(0, -1)

		assert.Equal(t, errors.Code(m.SanityCheck()), errors.ErrInvalidRound)
	})

	t.Run("OK", func(t *testing.T) {
		m := NewQueryProposalMessage(100, 0)

		assert.NoError(t, m.SanityCheck())
		assert.Contains(t, m.Fingerprint(), "100")
	})
}
