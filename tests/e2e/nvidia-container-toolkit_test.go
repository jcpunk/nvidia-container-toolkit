/*
 * Copyright (c) 2025, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Integration tests for Docker runtime
var _ = Describe("docker", Ordered, ContinueOnFailure, func() {
	var runner Runner

	// Install the NVIDIA Container Toolkit
	BeforeAll(func(ctx context.Context) {
		runner = NewRunner(
			WithHost(sshHost),
			WithPort(sshPort),
			WithSshKey(sshKey),
			WithSshUser(sshUser),
		)

		if installCTK {
			installer, err := NewToolkitInstaller(
				WithRunner(runner),
				WithImage(imageName+":"+imageTag),
				WithTemplate(dockerInstallTemplate),
			)
			Expect(err).ToNot(HaveOccurred())

			err = installer.Install()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	// GPUs are accessible in a container: Running nvidia-smi -L inside the
	// container shows the same output inside the container as outside the
	// container. This means that the following commands must all produce
	// the same output
	When("running nvidia-smi -L", Ordered, func() {
		var hostOutput string
		var err error

		BeforeAll(func(ctx context.Context) {
			hostOutput, _, err = runner.Run("nvidia-smi -L")
			Expect(err).ToNot(HaveOccurred())

			_, _, err := runner.Run("docker pull ubuntu")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should support NVIDIA_VISIBLE_DEVICES", func(ctx context.Context) {
			containerOutput, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=all ubuntu nvidia-smi -L")
			Expect(err).ToNot(HaveOccurred())
			Expect(containerOutput).To(Equal(hostOutput))
		})

		It("should support automatic CDI spec generation", func(ctx context.Context) {
			containerOutput, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all ubuntu nvidia-smi -L")
			Expect(err).ToNot(HaveOccurred())
			Expect(containerOutput).To(Equal(hostOutput))
		})

		It("should support automatic CDI spec generation with the --gpus flag", func(ctx context.Context) {
			containerOutput, _, err := runner.Run("docker run --rm -i --gpus=all --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all ubuntu nvidia-smi -L")
			Expect(err).ToNot(HaveOccurred())
			Expect(containerOutput).To(Equal(hostOutput))
		})

		It("should support the --gpus flag using the nvidia-container-runtime", func(ctx context.Context) {
			containerOutput, _, err := runner.Run("docker run --rm -i --runtime=nvidia --gpus all ubuntu nvidia-smi -L")
			Expect(err).ToNot(HaveOccurred())
			Expect(containerOutput).To(Equal(hostOutput))
		})

		It("should support the --gpus flag using the nvidia-container-runtime-hook", func(ctx context.Context) {
			containerOutput, _, err := runner.Run("docker run --rm -i --gpus all ubuntu nvidia-smi -L")
			Expect(err).ToNot(HaveOccurred())
			Expect(containerOutput).To(Equal(hostOutput))
		})
	})

	// A vectorAdd sample runs in a container with access to all GPUs.
	// The following should all produce the same result.
	When("Running the cuda-vectorAdd sample", Ordered, func() {
		var referenceOutput string

		BeforeAll(func(ctx context.Context) {
			_, _, err := runner.Run("docker pull nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should support NVIDIA_VISIBLE_DEVICES", func(ctx context.Context) {
			var err error
			referenceOutput, _, err = runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=all nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())

			Expect(referenceOutput).To(ContainSubstring("Test PASSED"))
		})

		It("should support automatic CDI spec generation", func(ctx context.Context) {
			out2, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(Equal(out2))
		})

		It("should support the --gpus flag using the nvidia-container-runtime", func(ctx context.Context) {
			out3, _, err := runner.Run("docker run --rm -i --runtime=nvidia --gpus all nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(Equal(out3))
		})

		It("should support the --gpus flag using the nvidia-container-runtime-hook", func(ctx context.Context) {
			out4, _, err := runner.Run("docker run --rm -i --gpus all nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(Equal(out4))
		})
	})

	// A deviceQuery sample runs in a container with access to all GPUs
	// The following should all produce the same result.
	When("Running the cuda-deviceQuery sample", Ordered, func() {
		var referenceOutput string

		BeforeAll(func(ctx context.Context) {
			_, _, err := runner.Run("docker pull nvcr.io/nvidia/k8s/cuda-sample:devicequery-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should support NVIDIA_VISIBLE_DEVICES", func(ctx context.Context) {
			var err error
			referenceOutput, _, err = runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=all nvcr.io/nvidia/k8s/cuda-sample:devicequery-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(ContainSubstring("Result = PASS"))
		})

		It("should support automatic CDI spec generation", func(ctx context.Context) {
			out2, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all nvcr.io/nvidia/k8s/cuda-sample:devicequery-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(Equal(out2))
		})

		It("should support the --gpus flag using the nvidia-container-runtime", func(ctx context.Context) {
			out3, _, err := runner.Run("docker run --rm -i --runtime=nvidia --gpus all nvcr.io/nvidia/k8s/cuda-sample:devicequery-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(Equal(out3))
		})

		It("should support the --gpus flag using the nvidia-container-runtime-hook", func(ctx context.Context) {
			out4, _, err := runner.Run("docker run --rm -i --gpus all nvcr.io/nvidia/k8s/cuda-sample:devicequery-cuda12.5.0")
			Expect(err).ToNot(HaveOccurred())
			Expect(referenceOutput).To(Equal(out4))
		})
	})

	When("Testing CUDA Forward compatibility", Ordered, func() {
		BeforeAll(func(ctx context.Context) {
			_, _, err := runner.Run("docker pull nvcr.io/nvidia/cuda:12.9.0-base-ubi8")
			Expect(err).ToNot(HaveOccurred())

			compatOutput, _, err := runner.Run("docker run --rm -i -e NVIDIA_VISIBLE_DEVICES=void nvcr.io/nvidia/cuda:12.9.0-base-ubi8 bash -c \"ls /usr/local/cuda/compat/libcuda.*.*\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(compatOutput).ToNot(BeEmpty())

			compatDriverVersion := strings.TrimPrefix(filepath.Base(compatOutput), "libcuda.so.")
			compatMajor := strings.SplitN(compatDriverVersion, ".", 2)[0]

			driverOutput, _, err := runner.Run("nvidia-smi -q | grep \"Driver Version\"")
			Expect(err).ToNot(HaveOccurred())
			parts := strings.SplitN(driverOutput, ":", 2)
			Expect(parts).To(HaveLen(2))

			hostDriverVersion := strings.TrimSpace(parts[1])
			Expect(hostDriverVersion).ToNot(BeEmpty())
			driverMajor := strings.SplitN(hostDriverVersion, ".", 2)[0]

			if driverMajor >= compatMajor {
				GinkgoLogr.Info("CUDA Forward Compatibility tests require an older driver version", "hostDriverVersion", hostDriverVersion, "compatDriverVersion", compatDriverVersion)
				Skip("CUDA Forward Compatibility tests require an older driver version")
			}
		})

		It("should work with the nvidia runtime in legacy mode", func(ctx context.Context) {
			ldconfigOut, _, err := runner.Run("docker run --rm -i -e NVIDIA_DISABLE_REQUIRE=true --runtime=nvidia --gpus all nvcr.io/nvidia/cuda:12.9.0-base-ubi8 bash -c \"ldconfig -p | grep libcuda.so.1\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(ldconfigOut).To(ContainSubstring("/usr/local/cuda-12.9/compat/"))
		})

		It("should work with the nvidia runtime in CDI mode", func(ctx context.Context) {
			ldconfigOut, _, err := runner.Run("docker run --rm -i -e NVIDIA_DISABLE_REQUIRE=true  --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all nvcr.io/nvidia/cuda:12.9.0-base-ubi8 bash -c \"ldconfig -p | grep libcuda.so.1\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(ldconfigOut).To(ContainSubstring("/usr/local/cuda-12.9/compat/"))
		})

		It("should work with nvidia-container-runtime-hook", func(ctx context.Context) {
			ldconfigOut, _, err := runner.Run("docker run --rm -i -e NVIDIA_DISABLE_REQUIRE=true --runtime=runc --gpus all nvcr.io/nvidia/cuda:12.9.0-base-ubi8 bash -c \"ldconfig -p | grep libcuda.so.1\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(ldconfigOut).To(ContainSubstring("/usr/local/cuda-12.9/compat/"))
		})
	})

	When("Disabling device node creation", Ordered, func() {
		BeforeAll(func(ctx context.Context) {
			_, _, err := runner.Run("docker pull ubuntu")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should work with nvidia-container-runtime-hook", func(ctx context.Context) {
			output, _, err := runner.Run("docker run --rm -i --runtime=runc --gpus=all ubuntu bash -c \"grep ModifyDeviceFiles: /proc/driver/nvidia/params\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("ModifyDeviceFiles: 0\n"))
		})

		It("should work with automatic CDI spec generation", func(ctx context.Context) {
			output, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all ubuntu bash -c \"grep ModifyDeviceFiles: /proc/driver/nvidia/params\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("ModifyDeviceFiles: 0\n"))
		})
	})

	When("A container is run using CDI", Ordered, func() {
		BeforeAll(func(ctx context.Context) {
			_, _, err := runner.Run("docker pull ubuntu")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should include libcuda.so in the ldcache", func(ctx context.Context) {
			ldcacheOutput, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all ubuntu bash -c \"ldconfig -p | grep 'libcuda.so'\"")
			Expect(err).ToNot(HaveOccurred())
			Expect(ldcacheOutput).ToNot(BeEmpty())

			ldcacheLines := strings.Split(ldcacheOutput, "\n")
			var libs []string
			for _, line := range ldcacheLines {
				parts := strings.SplitN(line, " (", 2)
				libs = append(libs, strings.TrimSpace(parts[0]))
			}

			Expect(libs).To(ContainElements([]string{"libcuda.so", "libcuda.so.1"}))
		})
	})

	When("Running containers with shared mount propagation", Ordered, func() {
		var mountsBefore string
		var tmpDirPath string

		BeforeAll(func(ctx context.Context) {
			_, _, err := runner.Run("docker pull ubuntu")
			Expect(err).ToNot(HaveOccurred())

			tmpDirPath = GinkgoT().TempDir()
			_, _, err = runner.Run("mkdir -p " + tmpDirPath)
			Expect(err).ToNot(HaveOccurred())

			output, _, err := runner.Run("mount | sort")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeEmpty())
			mountsBefore = output
		})

		AfterEach(func() {
			mountsAfter, _, err := runner.Run("mount | sort")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountsAfter).To(Equal(mountsBefore))
		})

		It("should not leak mounts when using the nvidia-container-runtime-hook", Label("legacy"), func(ctx context.Context) {
			_, _, err := runner.Run("docker run --rm -i --runtime=runc -e NVIDIA_VISIBLE_DEVICES=all -e NVIDIA_DRIVER_CAPABILITIES=all --mount type=bind,source=" + tmpDirPath + ",target=/empty,bind-propagation=shared ubuntu true")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not leak mounts when using the nvidia-container-runtime", func(ctx context.Context) {
			_, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=all -e NVIDIA_DRIVER_CAPABILITIES=all --mount type=bind,source=" + tmpDirPath + ",target=/empty,bind-propagation=shared ubuntu true")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not leak mounts when using CDI mode", func(ctx context.Context) {
			_, _, err := runner.Run("docker run --rm -i --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=all -e NVIDIA_DRIVER_CAPABILITIES=runtime.nvidia.com/gpu=all --mount type=bind,source=" + tmpDirPath + ",target=/empty,bind-propagation=shared ubuntu true")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("Running a container where the firmware folder resolves outside the container root", Ordered, func() {
		var outputDir string
		BeforeAll(func(ctx context.Context) {
			output, _, err := runner.Run("mktemp -d -p $(pwd)")
			Expect(err).ToNot(HaveOccurred())
			outputDir = strings.TrimSpace(output)

			_, _, err = runner.Run("docker pull ubuntu")
			Expect(err).ToNot(HaveOccurred())

			_, _, err = runner.Run(`docker build -t firmware-test \
            --build-arg RM_VERSION="$(basename $(ls -d /lib/firmware/nvidia/*.*))" \
            --build-arg CURRENT_DIR="` + outputDir + `" \
            - <<EOF
FROM ubuntu
RUN mkdir -p /lib/firmware/nvidia/
ARG RM_VERSION
ARG CURRENT_DIR
RUN ln -s /../../../../../../../../\$CURRENT_DIR /lib/firmware/nvidia/\$RM_VERSION
EOF`)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func(ctx context.Context) {
			output, _, err := runner.Run("ls -A " + outputDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(BeEmpty())
		})

		AfterAll(func(ctx context.Context) {
			if outputDir != "" {
				runner.Run(fmt.Sprintf("rm -rf %s", outputDir))
			}
		})

		It("should not fail when using CDI", func(ctx context.Context) {
			output, _, err := runner.Run("docker run --rm --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=runtime.nvidia.com/gpu=all firmware-test")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(BeEmpty())
		})

		It("should not fail when using the nvidia-container-runtime", func(ctx context.Context) {
			_, _, err := runner.Run("docker run --rm --runtime=nvidia -e NVIDIA_VISIBLE_DEVICES=all -e NVIDIA_DRIVER_CAPABILITIES=all firmware-test")
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail when using the nvidia-container-runtime-hook", Label("legacy"), func(ctx context.Context) {
			_, stderr, err := runner.Run("docker run --rm --runtime=runc --gpus=all firmware-test")
			Expect(err).To(HaveOccurred())
			Expect(stderr).To(ContainSubstring("nvidia-container-cli.real: mount error: path error:"))
		})
	})
})
