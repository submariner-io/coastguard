package e2e

import (
	"testing"

	_ "github.com/submariner-io/coastguard/test/e2e/scenarios"
	"github.com/submariner-io/submariner/test/e2e/framework"
)

func init() {
	framework.ParseFlags()
}

func TestE2E(t *testing.T) {

	RunE2ETests(t)
}
