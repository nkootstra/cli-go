package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "devcli",
		Short:         "Developer CLI with OIDC authentication",
		Long:          "devcli authenticates developers with an OIDC provider and reports local auth health.",
		SilenceUsage: true,
	}

	cmd.AddCommand(newLoginCommand())
	cmd.AddCommand(newLogoutCommand())
	cmd.AddCommand(newWhoamiCommand())
	cmd.AddCommand(newDoctorCommand())

	return cmd
}

func newLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with an OIDC provider",
		Long:  "Authenticate with an OIDC provider using an authorization-code flow with PKCE.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved login session",
		Long:  "Remove the saved local devcli login session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newWhoamiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the current logged-in identity",
		Long:  "Show the current logged-in identity stored by devcli login.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check devcli configuration and auth health",
		Long:  "Check devcli configuration, auth state, Docker mock-provider settings, and OIDC discovery health.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
