package auth

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewCommand returns a new [cobra.Command] for CLI authentication.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "Authenticate to the Mapon API",
		GroupID: "auth",
	}
	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newLogoutCommand())
	return cmd
}

func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the Mapon API",
	}
	apiKey := cmd.Flags().String("api-key", "", "API key to use for authentication")
	
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if *apiKey == "" {
			cmd.Print("Enter API key: ")
			input, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			*apiKey = string(input)
			cmd.Println()
		}
		
		authFile := File{
			APIKey: *apiKey,
		}
		
		if err := writeFile(&authFile); err != nil {
			return err
		}
		cmd.Println("Logged in.")
		return nil
	}
	return cmd
}

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from the Mapon API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := removeFile(); err != nil {
				return err
			}
			cmd.Println("Logged out.")
			return nil
		},
	}
}