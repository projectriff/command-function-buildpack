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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/projectriff/command-function-buildpack/command"
)

func testFunction(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "function-application")
		Expect(err).NotTo(HaveOccurred())

		ctx.Layers.Path, err = ioutil.TempDir("", "function-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("returns error if path does not exist", func() {
		_, err := command.NewFunction(ctx.Application.Path, "test-command")

		Expect(err).To(MatchError(fmt.Sprintf("command %s does not exist", "test-command")))
	})

	it("returns error if path is not a file", func() {
		Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "test-command"), 0755)).To(Succeed())

		_, err := command.NewFunction(ctx.Application.Path, "test-command")

		Expect(err).To(MatchError(fmt.Sprintf("command %s is not an executable file", "test-command")))
	})

	it("returns error if path is not executable", func() {
		Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "test-command"), []byte{}, 0644)).To(Succeed())

		_, err := command.NewFunction(ctx.Application.Path, "test-command")

		Expect(err).To(MatchError(fmt.Sprintf("command %s is not an executable file", "test-command")))
	})

	it("contributes function", func() {
		Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "test-command"), []byte{}, 0755)).To(Succeed())

		f, err := command.NewFunction(ctx.Application.Path, "test-command")
		Expect(err).NotTo(HaveOccurred())

		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = f.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Launch).To(BeTrue())
		Expect(layer.LaunchEnvironment["FUNCTION_URI.override"]).To(Equal(filepath.Join(ctx.Application.Path, "test-command")))
	})

}
