package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/nkootstra/cli/internal/auth"
	"github.com/nkootstra/cli/internal/config"
	"github.com/nkootstra/cli/internal/doctor"
	"github.com/nkootstra/cli/internal/version"
	"github.com/spf13/cobra"
)

var now = time.Now

func NewRootCommand() *cobra.Command {
	var showVersion bool
	cmd := &cobra.Command{
		Use:          "devcli",
		Short:        "Developer CLI with OIDC authentication",
		Long:         "devcli authenticates developers with an OIDC provider and reports local auth health.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "devcli %s\n", version.String())
				return err
			}
			return cmd.Help()
		},
	}

	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print the devcli version")
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
			return runLogin(cmd, flags)
		},
	}
	addLoginFlags(cmd, &flags)
	return cmd
}

func newLogoutCommand() *cobra.Command {
	var flags authFlags
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove the saved login session",
		Long:  "Remove the saved local devcli login session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd, flags)
		},
	}
	addAuthStateFlag(cmd, &flags)
	return cmd
}

func newWhoamiCommand() *cobra.Command {
	var flags authFlags
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the current logged-in identity",
		Long:  "Show the current logged-in identity stored by devcli login.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami(cmd, flags)
		},
	}
	addAuthStateFlag(cmd, &flags)
	return cmd
}

func newDoctorCommand() *cobra.Command {
	var flags authFlags
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check devcli configuration and auth health",
		Long:  "Check devcli configuration and auth health, including auth state, Docker mock-provider settings, and OIDC discovery health.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd, flags)
		},
	}
	addDoctorFlags(cmd, &flags)
	return cmd
}

func runLogin(cmd *cobra.Command, flags authFlags) error {
	cfg := resolveConfig(flags)
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
}

func runLogout(cmd *cobra.Command, flags authFlags) error {
	store := authStore(flags)
	state, err := store.Read(cmd.Context())
	if err != nil {
		return err
	}
	if err := store.Delete(cmd.Context()); err != nil {
		return err
	}
	if !state.Authenticated {
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "No active session remains.")
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
	return err
}

func runWhoami(cmd *cobra.Command, flags authFlags) error {
	store := authStore(flags)
	state, err := store.Read(cmd.Context())
	if err != nil {
		return err
	}
	session, err := auth.CurrentSession(state, now())
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	printField(out, "Subject", session.Subject)
	printField(out, "Email", session.Email)
	printField(out, "Name", session.Name)
	printField(out, "Issuer", session.Issuer)
	printField(out, "Client ID", session.Audience)
	return nil
}

func runDoctor(cmd *cobra.Command, flags authFlags) error {
	cfg := resolveConfig(flags)
	results := doctor.Run(cmd.Context(), doctor.Input{
		Config: cfg,
		Store:  auth.NewFileStore(cfg.AuthStatePath),
	})
	for _, result := range results {
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s: %s\n", result.Status, result.Name, result.Detail)
		if result.Remediation != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Fix: %s\n", result.Remediation)
		}
	}
	if doctor.HasFailure(results) {
		return fmt.Errorf("doctor found failing checks")
	}
	return nil
}

func printField(w io.Writer, label, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(w, "%s: %s\n", label, value)
}

func authStore(flags authFlags) auth.Store {
	cfg := resolveConfig(flags)
	return auth.NewFileStore(cfg.AuthStatePath)
}

func resolveConfig(flags authFlags) config.Config {
	return config.Resolve(config.Input{Overrides: flags.overrides()})
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

func addLoginFlags(cmd *cobra.Command, flags *authFlags) {
	addOIDCFlags(cmd, flags)
	addAuthStateFlag(cmd, flags)
}

func addDoctorFlags(cmd *cobra.Command, flags *authFlags) {
	addOIDCFlags(cmd, flags)
	addAuthStateFlag(cmd, flags)
	cmd.Flags().StringVar(&flags.mockIssuerURL, "mock-oidc-issuer-url", "", "Mock OIDC issuer URL for doctor/integration checks; env "+config.EnvMockOIDCIssuerURL)
}

func addOIDCFlags(cmd *cobra.Command, flags *authFlags) {
	cmd.Flags().StringVar(&flags.issuerURL, "oidc-issuer-url", "", "OIDC issuer URL; env "+config.EnvOIDCIssuerURL)
	cmd.Flags().StringVar(&flags.clientID, "oidc-client-id", "", "OIDC client ID; env "+config.EnvOIDCClientID)
	cmd.Flags().StringVar(&flags.scopes, "oidc-scopes", "", "OIDC scopes separated by spaces or commas; env "+config.EnvOIDCScopes)
	cmd.Flags().StringVar(&flags.redirectURL, "oidc-redirect-url", "", "OIDC redirect URL; env "+config.EnvOIDCRedirectURL)
	cmd.Flags().StringVar(&flags.callbackListenAddr, "oidc-callback-listen-addr", "", "Local callback listen address; env "+config.EnvOIDCCallbackListenAddr)
}

func addAuthStateFlag(cmd *cobra.Command, flags *authFlags) {
	cmd.Flags().StringVar(&flags.authStatePath, "auth-state-path", "", "Path to devcli auth state; env "+config.EnvAuthStatePath)
}
