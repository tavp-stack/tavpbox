package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/podman"
)

var imageBuildCmd = &cobra.Command{
	Use:   "build [name]",
	Short: "Build custom image from current container",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return fmt.Errorf(".tavpbox.yml not found")
		}

		client := podman.New()
		cname := client.ContainerName(cfg.Name)

		// Check if container exists
		containers, _ := client.List()
		found := false
		for _, c := range containers {
			if c.Name == cname {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("container %s not found. Run: tavpbox create", cfg.Name)
		}

		imageName := cfg.Name + "-custom"
		if len(args) > 0 {
			imageName = args[0]
		}

		// Build image from container
		fmt.Printf("Building image '%s' from container '%s'...\n", imageName, cname)
		_, err = client.Exec(cname, "bash", "-c", "apt-get clean && rm -rf /var/lib/apt/lists/*")
		if err != nil {
			fmt.Printf("  ⚠ Cleanup warning: %v\n", err)
		}

		_, err = client.Commit(cname, imageName)
		if err != nil {
			return fmt.Errorf("build image: %w", err)
		}

		fmt.Printf("✓ Image '%s' built successfully!\n", imageName)
		fmt.Printf("  Use in .tavpbox.yml: image: %s\n", imageName)
		return nil
	},
}

var imagePushCmd = &cobra.Command{
	Use:   "push <image>",
	Short: "Push image to registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]
		fmt.Printf("Pushing image '%s'...\n", image)

		client := podman.New()
		_, err := client.Push(image)
		if err != nil {
			return fmt.Errorf("push image: %w", err)
		}

		fmt.Printf("✓ Image '%s' pushed successfully!\n", image)
		return nil
	},
}

var imagePullCmd = &cobra.Command{
	Use:   "pull <image>",
	Short: "Pull image from registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]
		fmt.Printf("Pulling image '%s'...\n", image)

		client := podman.New()
		err := client.Pull(image)
		if err != nil {
			return fmt.Errorf("pull image: %w", err)
		}

		fmt.Printf("✓ Image '%s' pulled successfully!\n", image)
		return nil
	},
}

var imageListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List local images",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := podman.New()
		output, err := client.ListImages()
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage custom images",
	Long: `Manage custom pre-built images for faster container creation.

TAVPBox provides official base images:
  - ghcr.io/tavp-stack/tavpbox-php    (PHP 8.2 + Nginx + Node.js + Phalcon)
  - ghcr.io/tavp-stack/tavpbox-node   (Node.js 20 + Nginx)
  - ghcr.io/tavp-stack/tavpbox-go     (Go 1.22 + Nginx)
  - ghcr.io/tavp-stack/tavpbox-python (Python 3.12 + Nginx)

Usage:
  tavpbox image build --name my-php    # Build custom image
  tavpbox image push ghcr.io/user/img  # Push to registry
  tavpbox image pull ghcr.io/user/img  # Pull from registry
  tavpbox image list                    # List local images

Custom image in .tavpbox.yml:
  image: ghcr.io/myuser/my-php:latest`,
}

func init() {
	imageCmd.AddCommand(imageBuildCmd, imagePushCmd, imagePullCmd, imageListCmd)
	rootCmd.AddCommand(imageCmd)
}
