package e2etests_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	graphql "github.com/hasura/go-graphql-client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tests")
}

var _ = Describe("graphql", func() {
	var simple *graphql.Client
	BeforeEach(func() {
		out := ReadOutputs()
		simple = graphql.NewClient(out.ClProtoASAppMain.SimpleGraphHttpUrl, http.DefaultClient)
		simple = simple.WithRequestModifier(func(r *http.Request) {
			r.Header.Set("x-api-key", out.ClProtoASAppMain.SimpleGraphSecretKey)
		})
	})

	DescribeTable("simple api queries", func(ctx context.Context, query any, exp string) {
		Expect(simple.Query(ctx, query, nil)).To(Succeed())
		Expect(json.Marshal(query)).To(MatchJSON(exp))
	},
		Entry("string", &struct{ Version string }{}, `{"Version":"v0.1.2"}`),
	)
})

func ReadOutputs() (out struct {
	ClProtoASAppMain struct {
		SimpleGraphSecretKey string `json:"SimpleGraphSecretKey913CC2BA"`
		SimpleGraphHttpUrl   string `json:"SimpleGraphHttpURL9040784C"`
		NestedGraphHttpUrl   string `json:"NestedGraphHttpURL12EC9D4F"`
		NestedGraphSecretKey string `json:"NestedGraphSecretKey550C5244"`
	}
}) {
	data, err := os.ReadFile(filepath.Join("..", "infra", "cdk.outputs.json"))
	Expect(err).ToNot(HaveOccurred())
	Expect(json.Unmarshal(data, &out)).To(Succeed())
	return
}
