package latest

import (
	"testing"
)

func TestGithubTag_implement(t *testing.T) {
	var _ Source = &GithubTag{}
}
