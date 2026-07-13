package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/lxd"
)

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Manage LXC images",
}

var imagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached images",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := lxd.New()

		output, err := client.ExecNoTTY("local", "image", "list", "--format", "csv", "-c", "lfd")
		if err != nil {
			// Try alternative command
			output, err = client.ExecNoTTY("local", "image", "list")
			if err != nil {
				return fmt.Errorf("failed to list images: %w", err)
			}
		}

		fmt.Println("Cached Images:")
		fmt.Println("────────────────────────────────────────────")
		fmt.Print(output)

		return nil
	},
}

var imagesPullCmd = &cobra.Command{
	Use:   "pull <image>",
	Short: "Pull/cache an image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		image := args[0]
		client := lxd.New()

		fmt.Printf("Pulling image: %s...\n", image)

		// Create a temporary container to cache the image
		tempName := "tavpbox-image-cache"
		if err := client.Create(tempName, image, "64MB", 1); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}

		// Delete the temporary container (image stays cached)
		client.Delete(tempName)

		fmt.Printf("✓ Image '%s' cached\n", image)
		return nil
	},
}

var imagesRemoveCmd = &cobra.Command{
	Use:   "remove <fingerprint>",
	Short: "Remove a cached image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fingerprint := args[0]
		client := lxd.New()

		if _, err := client.ExecNoTTY("local", "image", "delete", fingerprint); err != nil {
			return fmt.Errorf("failed to remove image: %w", err)
		}

		fmt.Printf("✓ Image '%s' removed\n", fingerprint)
		return nil
	},
}

func init() {
	imagesCmd.AddCommand(imagesListCmd)
	imagesCmd.AddCommand(imagesPullCmd)
	imagesCmd.AddCommand(imagesRemoveCmd)
	rootCmd.AddCommand(imagesCmd)
}
