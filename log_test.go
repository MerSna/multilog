package multilog

import (
	"testing"
)

func TestName(t *testing.T) {
	l := G()
	l.Info(123)

}
