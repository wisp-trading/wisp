package manager_test

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	mockMonitoring "github.com/wisp-trading/sdk/mocks/github.com/wisp-trading/sdk/pkg/types/monitoring"
	"github.com/wisp-trading/sdk/pkg/types/logging"
	"github.com/wisp-trading/wisp/internal/services/live/manager"
	mockLive "github.com/wisp-trading/wisp/mocks/github.com/wisp-trading/wisp/pkg/live"
	"github.com/wisp-trading/wisp/pkg/live"
)

func runningInstance(id, name string, pid int) *live.Instance {
	return &live.Instance{
		ID:           id,
		StrategyName: name,
		PID:          pid,
		Status:       live.StatusRunning,
	}
}

var _ = Describe("InstanceManager lifecycle (P0)", func() {
	var (
		stateStore *mockLive.StateStore
		spawner    *mockLive.ProcessSpawner
	)

	BeforeEach(func() {
		stateStore = mockLive.NewStateStore(GinkgoT())
		spawner = mockLive.NewProcessSpawner(GinkgoT())
		stateStore.EXPECT().Save(mock.Anything).Return(nil).Maybe()
	})

	Describe("LoadRunning", func() {
		It("marks dead PIDs as crashed using Signal(0) liveness", func() {
			deadPID := 1<<22 - 7
			stateStore.EXPECT().Load().Return([]*live.Instance{
				runningInstance("dead-1", "gone", deadPID),
			}, nil).Once()

			im := manager.NewInstanceManager(stateStore, spawner, logging.NewNoOpLogger(), nil)
			Expect(im.LoadRunning(context.Background())).To(Succeed())

			crashed, err := im.List(live.StatusCrashed)
			Expect(err).NotTo(HaveOccurred())
			Expect(crashed).To(HaveLen(1))
			Expect(crashed[0].Error).To(ContainSubstring("Process not found"))
		})

		It("reattaches live processes with nil Cmd", func() {
			if runtime.GOOS == "windows" {
				Skip("Unix process model")
			}
			cmd := exec.Command("sleep", "30")
			Expect(cmd.Start()).To(Succeed())
			DeferCleanup(func() {
				_ = cmd.Process.Kill()
				_, _ = cmd.Process.Wait()
			})

			stateStore.EXPECT().Load().Return([]*live.Instance{
				runningInstance("live-1", "alive", cmd.Process.Pid),
			}, nil).Once()

			im := manager.NewInstanceManager(stateStore, spawner, logging.NewNoOpLogger(), nil)
			Expect(im.LoadRunning(context.Background())).To(Succeed())

			running, err := im.List(live.StatusRunning)
			Expect(err).NotTo(HaveOccurred())
			Expect(running).To(HaveLen(1))
			Expect(running[0].Cmd).To(BeNil())
			Expect(running[0].Cancel).NotTo(BeNil())
			running[0].Cancel()
		})
	})

	Describe("Kill", func() {
		It("kills reattached instances with nil Cmd via PID", func() {
			if runtime.GOOS == "windows" {
				Skip("Unix process model")
			}
			cmd := exec.Command("sleep", "60")
			Expect(cmd.Start()).To(Succeed())
			pid := cmd.Process.Pid

			// Reap in background — Signal(0) still succeeds on zombies without Wait.
			exited := make(chan error, 1)
			go func() { exited <- cmd.Wait() }()

			stateStore.EXPECT().Load().Return([]*live.Instance{
				runningInstance("reattach-1", "s", pid),
			}, nil).Once()

			im := manager.NewInstanceManager(stateStore, spawner, logging.NewNoOpLogger(), nil)
			Expect(im.LoadRunning(context.Background())).To(Succeed())
			Expect(im.Kill("reattach-1")).To(Succeed())

			Eventually(exited, 2*time.Second).Should(Receive())

			stopped, err := im.List(live.StatusStopped)
			Expect(err).NotTo(HaveOccurred())
			Expect(stopped).To(HaveLen(1))
		})
	})

	Describe("Stop", func() {
		It("prefers HTTP /shutdown before force-killing", func() {
			if runtime.GOOS == "windows" {
				Skip("Unix process model")
			}
			cmd := exec.Command("sleep", "60")
			Expect(cmd.Start()).To(Succeed())
			pid := cmd.Process.Pid
			DeferCleanup(func() {
				_ = cmd.Process.Kill()
				_, _ = cmd.Process.Wait()
			})

			querier := mockMonitoring.NewViewQuerier(GinkgoT())
			// Graceful path: HTTP /shutdown kills the process so wait returns promptly.
			querier.EXPECT().Shutdown("http-strat").Run(func(string) {
				_ = cmd.Process.Kill()
			}).Return(nil).Once()

			stateStore.EXPECT().Load().Return([]*live.Instance{
				runningInstance("http-1", "http-strat", pid),
			}, nil).Once()

			im := manager.NewInstanceManager(stateStore, spawner, logging.NewNoOpLogger(), querier)
			Expect(im.LoadRunning(context.Background())).To(Succeed())
			Expect(im.Stop("http-1")).To(Succeed())

			stopped, err := im.List(live.StatusStopped)
			Expect(err).NotTo(HaveOccurred())
			Expect(stopped).To(HaveLen(1))
		})
	})
})
