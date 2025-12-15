package main

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/way-platform/mapon-go"
	"github.com/way-platform/mapon-go/cmd/mapon/internal/auth"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	if err := fang.Execute(
		context.Background(),
		newRootCommand(),
		fang.WithColorSchemeFunc(func(c lipgloss.LightDarkFunc) fang.ColorScheme {
			base := c(lipgloss.Black, lipgloss.White)
			baseInverted := c(lipgloss.White, lipgloss.Black)
			return fang.ColorScheme{
				Base:         base,
				Title:        base,
				Description:  base,
				Comment:      base,
				Flag:         base,
				FlagDefault:  base,
				Command:      base,
				QuotedString: base,
				Argument:     base,
				Help:         base,
				Dash:         base,
				ErrorHeader:  [2]color.Color{baseInverted, base},
				ErrorDetails: base,
			}
		}),
	); err != nil {
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mapon",
		Short: "Mapon API CLI",
	}
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")

	cmd.AddGroup(&cobra.Group{ID: "units", Title: "Units"})
	cmd.AddCommand(newListUnitsCommand())

	cmd.AddGroup(&cobra.Group{ID: "drivers", Title: "Drivers"})
	cmd.AddCommand(newListDriversCommand())

	cmd.AddGroup(&cobra.Group{ID: "routes", Title: "Routes"})
	cmd.AddCommand(newListRoutesCommand())

	cmd.AddGroup(&cobra.Group{ID: "objects", Title: "Objects"})
	cmd.AddCommand(newListObjectsCommand())

	cmd.AddGroup(&cobra.Group{ID: "alerts", Title: "Alerts"})
	cmd.AddCommand(newListAlertsCommand())

	cmd.AddGroup(auth.NewGroup())
	cmd.AddCommand(auth.NewCommand())

	cmd.AddGroup(&cobra.Group{
		ID:    "utils",
		Title: "Utils",
	})
	cmd.SetHelpCommandGroupID("utils")
	cmd.SetCompletionCommandGroupID("utils")
	return cmd
}

func newListUnitsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "units",
		Short:   "List units",
		GroupID: "units",
	}
	ids := cmd.Flags().StringSlice("id", nil, "Filter by unit ID")
	include := cmd.Flags().StringSlice("include", nil, "Include additional data (fuel, drivers, location, routes)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}

		var unitIDs []int64
		for _, idStr := range *ids {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}

		response, err := client.ListUnits(cmd.Context(), &mapon.ListUnitsRequest{
			UnitIDs: unitIDs,
			Include: *include,
		})
		if err != nil {
			return err
		}
		for _, unit := range response.Units {
			fmt.Println(protojson.Format(unit))
		}
		return nil
	}
	return cmd
}

func newListDriversCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "drivers",
		Short:   "List drivers",
		GroupID: "drivers",
	}
	id := cmd.Flags().Int64("id", 0, "Filter by driver ID")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		response, err := client.ListDrivers(cmd.Context(), &mapon.ListDriversRequest{
			ID: *id,
		})
		if err != nil {
			return err
		}
		for _, driver := range response.Drivers {
			fmt.Println(protojson.Format(driver))
		}
		return nil
	}
	return cmd
}

func newListRoutesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routes",
		Short:   "List routes",
		GroupID: "routes",
	}
	from := cmd.Flags().String("from", "", "Start time (RFC3339)")
	till := cmd.Flags().String("till", "", "End time (RFC3339)")
	ids := cmd.Flags().StringSlice("unit-id", nil, "Filter by unit ID")
	include := cmd.Flags().StringSlice("include", nil, "Include additional data (polyline)")

	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("till")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}

		fromTime, err := time.Parse(time.RFC3339, *from)
		if err != nil {
			return fmt.Errorf("invalid from time: %w", err)
		}
		tillTime, err := time.Parse(time.RFC3339, *till)
		if err != nil {
			return fmt.Errorf("invalid till time: %w", err)
		}
		var unitIDs []int64
		for _, idStr := range *ids {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		response, err := client.ListRoutes(cmd.Context(), &mapon.ListRoutesRequest{
			From:    fromTime,
			Till:    tillTime,
			UnitIDs: unitIDs,
			Include: *include,
		})
		if err != nil {
			return err
		}
		for _, route := range response.Routes {
			fmt.Println(protojson.Format(route))
		}
		return nil
	}
	return cmd
}

func newListObjectsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "objects",
		Short:   "List objects",
		GroupID: "objects",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		response, err := client.ListObjects(cmd.Context(), &mapon.ListObjectsRequest{})
		if err != nil {
			return err
		}
		for _, object := range response.Objects {
			fmt.Println(protojson.Format(object))
		}
		return nil
	}
	return cmd
}

func newListAlertsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alerts",
		Short:   "List alerts",
		GroupID: "alerts",
	}
	from := cmd.Flags().String("from", "", "Start time (RFC3339)")
	till := cmd.Flags().String("till", "", "End time (RFC3339)")
	ids := cmd.Flags().StringSlice("unit-id", nil, "Filter by unit ID")
	driver := cmd.Flags().Int64("driver-id", 0, "Filter by driver ID")

	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("till")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		fromTime, err := time.Parse(time.RFC3339, *from)
		if err != nil {
			return fmt.Errorf("invalid from time: %w", err)
		}
		tillTime, err := time.Parse(time.RFC3339, *till)
		if err != nil {
			return fmt.Errorf("invalid till time: %w", err)
		}
		var unitIDs []int64
		for _, idStr := range *ids {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		response, err := client.ListAlerts(cmd.Context(), &mapon.ListAlertsRequest{
			From:    fromTime,
			Till:    tillTime,
			UnitIDs: unitIDs,
			Driver:  *driver,
		})
		if err != nil {
			return err
		}
		for _, alert := range response.Alerts {
			fmt.Println(protojson.Format(alert))
		}
		return nil
	}
	return cmd
}

func newClient(cmd *cobra.Command) (*mapon.Client, error) {
	debug, err := cmd.Root().PersistentFlags().GetBool("debug")
	if err != nil {
		return nil, err
	}
	return auth.NewClient(
		cmd.Context(),
		mapon.WithDebug(debug),
	)
}
