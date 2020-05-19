/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package command_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/projectriff/command-function-buildpack/command"
	"github.com/sclevine/spec"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "function-application")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
	})

	it("contributes invoker", func() {
		Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "test-artifact"), []byte{}, 0755)).To(Succeed())

		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{
			Name:     "riff-command",
			Metadata: map[string]interface{}{"artifact": "test-artifact"},
		})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "invoker",
					"version": "1.1.1",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"

		result, err := command.Build{}.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(2))
		Expect(result.Layers[0].Name()).To(Equal("invoker"))
		Expect(result.Layers[1].Name()).To(Equal("function"))

		Expect(result.Processes).To(ContainElements(
			libcnb.Process{Type: "command-function", Command: "command-function-invoker"},
			libcnb.Process{Type: "function", Command: "command-function-invoker"},
			libcnb.Process{Type: "web", Command: "command-function-invoker"},
		))
	})

}
