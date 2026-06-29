package cli

import (
	"fmt"

	"github.com/nkootstra/cli/internal/auth"
	"github.com/nkootstra/cli/internal/config"
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
	var flags authFlags
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with an OIDC provider",
		Long:  "Authenticate with an OIDC provider using an authorization-code flow with PKCE.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Resolve(config.Input{Overrides: flags.overrides()})
			store := auth.NewFileStore(cfg.AuthStatePath)
			_, err := auth.Login(cmd.Context(), auth.LoginOptions{
				Config: cfg,
				Store:  store,
				OpenURL: func(url string) error {
					_, err := fmt.Fprintf(cmd.OutOrStdout(), "Open this URL to continue login:\n%s\n", url)
					return err
				},
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Login complete.")
			return err
		},
	}
	addAuthFlags(cmd, &flags)
	return cmd
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

type authFlags struct {
	issuerURL          string
	clientID           string
	scopes             string
	redirectURL        string
	callbackListenAddr string
	authStatePath      string
	mockIssuerURL      string
}

func (f authFlags) overrides() config.Overrides {
	return config.Overrides{
		OIDCIssuerURL:          f.issuerURL,
		OIDCClientID:           f.clientID,
		OIDCScopes:             config.ParseScopes(f.scopes),
		OIDCRedirectURL:        f.redirectURL,
		OIDCCallbackListenAddr: f.callbackListenAddr,
		AuthStatePath:          f.authStatePath,
		MockOIDCIssuerURL:      f.mockIssuerURL,
	}
}

func addAuthFlags(cmd *cobra.Command, flags *authFlags) {
	cmd.Flags().StringVar(&flags.issuerURL, "oidc-issuer-url", "", "OIDC issuer URL; env "+config.EnvOIDCIssuerURL)
	cmd.Flags().StringVar(&flags.clientID, "oidc-client-id", "", "OIDC client ID; env "+config.EnvOIDCClientID)
	cmd.Flags().StringVar(&flags.scopes, "oidc-scopes", "", "OIDC scopes separated by spaces or commas; env "+config.EnvOIDCScopes)
	cmd.Flags().StringVar(&flags.redirectURL, "oidc-redirect-url", "", "OIDC redirect URL; env "+config.EnvOIDCRedirectURL)
	cmd.Flags().StringVar(&flags.callbackListenAddr, "oidc-callback-listen-addr", "", "Local callback listen address; env "+config.EnvOIDCCallbackListenAddr)
	cmd.Flags().StringVar(&flags.authStatePath, "auth-state-path", "", "Path to devcli auth state; env "+config.EnvAuthStatePath)
	cmd.Flags().StringVar(&flags.mockIssuerURL, "mock-oidc-issuer-url", "", "Mock OIDC issuer URL for doctor/integration checks; env "+config.EnvMockOIDCIssuerURL)
}
