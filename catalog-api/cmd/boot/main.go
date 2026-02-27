package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"digital.vasic.containers/pkg/boot"
	"digital.vasic.containers/pkg/compose"
	"digital.vasic.containers/pkg/distribution"
	"digital.vasic.containers/pkg/endpoint"
	"digital.vasic.containers/pkg/envconfig"
	"digital.vasic.containers/pkg/health"
	"digital.vasic.containers/pkg/logging"
	"digital.vasic.containers/pkg/remote"
	"digital.vasic.containers/pkg/runtime"
	"digital.vasic.containers/pkg/scheduler"
)

func main() {
	envFile := flag.String("env", "../Containers/.env", "Path to .env file")
	composeFile := flag.String("compose", "docker-compose.dev.yml", "Path to docker-compose file")
	localOnly := flag.Bool("local", false, "Run only locally (no remote distribution)")
	dryRun := flag.Bool("dry-run", false, "Show distribution plan without deploying")
	timeout := flag.Duration("timeout", 5*time.Minute, "Boot timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	logger := logging.NopLogger{}

	fmt.Println("=== Catalogizer Distributed Boot ===")
	fmt.Printf("Environment file: %s\n", *envFile)
	fmt.Printf("Compose file: %s\n", *composeFile)
	fmt.Printf("Local only: %v\n", *localOnly)

	rt, err := runtime.AutoDetect(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to detect container runtime: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Detected runtime: %s\n", rt.Name())

	var envCfg *envconfig.DistributionConfig
	if _, err := os.Stat(*envFile); err == nil {
		envCfg, err = envconfig.LoadFromFile(*envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not load .env file: %v\n", err)
			envCfg = envconfig.LoadFromEnv()
		}
	} else {
		envCfg = envconfig.LoadFromEnv()
	}

	composePath := *composeFile
	orchestrator := compose.NewOrchestrator("docker", []string{"compose"}, ".", logger)

	var hostManager remote.HostManager
	var dist *distribution.DefaultDistributor

	if !*localOnly && envCfg.Enabled {
		fmt.Println("Remote distribution enabled")
		fmt.Printf("Scheduler strategy: %s\n", envCfg.Scheduler)

		exec, err := remote.NewSSHExecutor(logger)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create SSH executor: %v\n", err)
			os.Exit(1)
		}

		hostManager = remote.NewHostManager(exec, logger)

		remoteHosts := envCfg.ToRemoteHosts()
		for _, host := range remoteHosts {
			fmt.Printf("Registering host: %s (%s@%s:%d)\n",
				host.Name, host.User, host.Address, host.SSHPort())
			if err := hostManager.AddHost(host); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to add host %s: %v\n", host.Name, err)
			}
		}

		var strategy scheduler.PlacementStrategy
		switch envCfg.Scheduler {
		case "round_robin":
			strategy = scheduler.StrategyRoundRobin
		case "affinity":
			strategy = scheduler.StrategyAffinity
		case "spread":
			strategy = scheduler.StrategySpread
		case "bin_pack":
			strategy = scheduler.StrategyBinPack
		default:
			strategy = scheduler.StrategyResourceAware
		}

		sched := scheduler.NewScheduler(hostManager, logger,
			scheduler.WithStrategy(strategy),
		)

		dist = distribution.NewDistributor(
			distribution.WithScheduler(sched),
			distribution.WithHostManager(hostManager),
			distribution.WithLogger(logger),
		)

		fmt.Println("Probing remote hosts...")
		resources := hostManager.ProbeAll(ctx)
		for name, res := range resources {
			if res != nil {
				fmt.Printf("Host %s: CPU=%.1f%%, Memory=%.1f%%, Disk=%.1f%%\n",
					name, res.CPUPercent, res.MemoryPercent, res.DiskPercent)
			}
		}
	} else {
		fmt.Println("Running in local-only mode")
	}

	endpoints := map[string]endpoint.ServiceEndpoint{
		"catalog-api": endpoint.NewEndpoint().
			WithHost("localhost").
			WithPort("8080").
			WithHealthType("http").
			WithHealthPath("/health").
			WithRequired(true).
			WithComposeFile(composePath).
			WithServiceName("api").
			Build(),
		"catalog-web": endpoint.NewEndpoint().
			WithHost("localhost").
			WithPort("3000").
			WithHealthType("http").
			WithHealthPath("/").
			WithRequired(true).
			WithComposeFile(composePath).
			WithServiceName("web").
			Build(),
	}

	if *dryRun && dist != nil {
		fmt.Println("\n=== Dry Run: Distribution Plan ===")
		reqs := []scheduler.ContainerRequirements{
			{Name: "catalog-api", Image: "catalogizer-api:latest", CPUCores: 1, MemoryMB: 1024, DiskMB: 1024},
			{Name: "catalog-web", Image: "catalogizer-web:latest", CPUCores: 0.5, MemoryMB: 512, DiskMB: 100},
		}

		var strategy scheduler.PlacementStrategy = scheduler.StrategyResourceAware
		sched := scheduler.NewScheduler(hostManager, logger,
			scheduler.WithStrategy(strategy),
		)

		plan, err := sched.ScheduleBatch(ctx, reqs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to schedule: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nDistribution Plan:")
		fmt.Println("==================")
		for _, decision := range plan.Decisions {
			if decision.Score > 0 {
				fmt.Printf("  %s -> %s (score: %.2f)\n",
					decision.Requirement.Name, decision.HostName, decision.Score)
			} else {
				fmt.Printf("  %s -> FAILED: %s\n",
					decision.Requirement.Name, decision.Reason)
			}
		}

		fmt.Println("\nHost Snapshots:")
		for name, snap := range plan.HostSnapshots {
			fmt.Printf("  %s: CPU=%.1f%%, Memory=%.1f%%, Disk=%.1f%%\n",
				name, snap.CPUPercent, snap.MemoryPercent, snap.DiskPercent)
		}
		os.Exit(0)
	}

	healthChecker := health.NewDefaultChecker()

	opts := []boot.BootManagerOption{
		boot.WithRuntime(rt),
		boot.WithOrchestrator(orchestrator),
		boot.WithHealthChecker(healthChecker),
		boot.WithLogger(logger),
	}

	if dist != nil {
		opts = append(opts, boot.WithDistributor(dist))
	}
	if hostManager != nil {
		opts = append(opts, boot.WithHostManager(hostManager))
	}

	mgr := boot.NewBootManager(endpoints, opts...)

	fmt.Println("Starting boot sequence...")
	summary, err := mgr.BootAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Boot failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Boot Summary ===")
	fmt.Printf("Started:  %d\n", summary.Started)
	fmt.Printf("Failed:   %d\n", summary.Failed)
	fmt.Printf("Skipped:  %d\n", summary.Skipped)
	fmt.Printf("Duration: %s\n", summary.TotalDuration)

	if summary.Failed > 0 {
		fmt.Println("\nFailed services:")
		for name, result := range summary.Results {
			if result.Status == "failed" {
				fmt.Printf("  - %s: %s\n", name, result.Error)
			}
		}
		os.Exit(1)
	}

	fmt.Println("\nAll services booted successfully!")

	statusURL := "http://localhost:8080/health"
	fmt.Printf("\nAPI Health: %s\n", statusURL)
	fmt.Println("Web UI: http://localhost:3000")
	fmt.Println("\nPress Ctrl+C to stop...")

	<-ctx.Done()
	fmt.Println("\nShutting down...")
}
