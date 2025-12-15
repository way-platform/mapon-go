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
	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
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

	cmd.AddGroup(&cobra.Group{ID: "unit-data", Title: "Unit Data"})
	cmd.AddCommand(newListIgnitionsCommand())
	cmd.AddCommand(newListTemperaturesCommand())
	cmd.AddCommand(newListDigitalInputsCommand())
	cmd.AddCommand(newListDigitalInputsExtendedCommand())
	cmd.AddCommand(newListIbuttonsCommand())
	cmd.AddCommand(newListHumidityCommand())
	cmd.AddCommand(newListCanPeriodDataCommand())
	cmd.AddCommand(newGetCanPointDataCommand())
	cmd.AddCommand(newGetHistoryPointDataCommand())
	cmd.AddCommand(newGetUnitFieldsCommand())
	cmd.AddCommand(newGetUnitDebugInfoCommand())
	cmd.AddCommand(newGetDrivingTimeExtendedCommand())

	cmd.AddGroup(&cobra.Group{ID: "unit-groups", Title: "Unit Groups"})
	cmd.AddCommand(newListUnitGroupsCommand())
	cmd.AddCommand(newListUnitsInGroupCommand())

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

// --- Units ---

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

// --- Unit Data ---

func newListIgnitionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ignitions <unit-id ...>",
		Short:   "List ignition events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.ListIgnitions(cmd.Context(), &mapon.ListIgnitionsRequest{
			UnitIDs: unitIDs,
			From:    *from,
			To:      *to,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListTemperaturesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "temperatures <unit-id ...>",
		Short:   "List temperature data",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.ListTemperatures(cmd.Context(), &mapon.ListTemperaturesRequest{
			UnitIDs: unitIDs,
			From:    *from,
			To:      *to,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListDigitalInputsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "digital-inputs <unit-id ...>",
		Short:   "List digital input events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.ListDigitalInputs(cmd.Context(), &mapon.ListDigitalInputsRequest{
			UnitIDs: unitIDs,
			From:    *from,
			To:      *to,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListDigitalInputsExtendedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "digital-inputs-extended <unit-id ...>",
		Short:   "List extended digital input events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.ListDigitalInputsExtended(cmd.Context(), &mapon.ListDigitalInputsExtendedRequest{
			UnitIDs: unitIDs,
			From:    *from,
			To:      *to,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListIbuttonsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ibuttons <unit-id ...>",
		Short:   "List iButton events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.ListIbuttons(cmd.Context(), &mapon.ListIbuttonsRequest{
			UnitIDs: unitIDs,
			From:    *from,
			To:      *to,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListHumidityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "humidity <unit-id ...>",
		Short:   "List humidity data",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.ListHumidity(cmd.Context(), &mapon.ListHumidityRequest{
			UnitIDs: unitIDs,
			From:    *from,
			To:      *to,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListCanPeriodDataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "can-periods <unit-id>",
		Short:   "List CAN data for a period",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	include := cmd.Flags().StringSlice("include", nil, "Fields to include")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		res, err := client.ListCanPeriodData(cmd.Context(), &mapon.ListCanPeriodDataRequest{
			UnitID:  unitID,
			From:    *from,
			To:      *to,
			Include: *include,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetCanPointDataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "can-point <unit-id>",
		Short:   "Get CAN data at a specific time",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	datetime := cmd.Flags().Time("datetime", time.Now(), []string{time.DateOnly, time.RFC3339}, "Datetime")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		res, err := client.GetCanDataPoint(cmd.Context(), &mapon.GetCanPointDataRequest{
			UnitID:   unitID,
			Datetime: *datetime,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetHistoryPointDataCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "history-point <unit-id>",
		Short:   "Get historical data at a specific time",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	datetime := cmd.Flags().Time("datetime", time.Now(), []string{time.DateOnly, time.RFC3339}, "Datetime")
	include := cmd.Flags().StringSlice("include", nil, "Fields to include")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		res, err := client.GetHistoryPointData(cmd.Context(), &mapon.GetHistoryPointDataRequest{
			UnitID:   unitID,
			Datetime: *datetime,
			Include:  *include,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetUnitFieldsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fields <unit-id>",
		Short:   "Get unit custom fields",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		res, err := client.GetUnitFields(cmd.Context(), &mapon.GetUnitFieldsRequest{
			UnitID: unitID,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetUnitDebugInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug-info <unit-id ...>",
		Short:   "Get unit debug info",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		var unitIDs []int64
		for _, idStr := range args {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid unit ID %s: %w", idStr, err)
			}
			unitIDs = append(unitIDs, id)
		}
		res, err := client.GetUnitDebugInfo(cmd.Context(), &mapon.GetUnitDebugInfoRequest{
			UnitIDs: unitIDs,
		})
		if err != nil {
			return err
		}
		for _, u := range res.Units {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetDrivingTimeExtendedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "driving-time <unit-id>",
		Short:   "Get driving time extended",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		res, err := client.GetDrivingTimeExtended(cmd.Context(), &mapon.GetDrivingTimeExtendedRequest{
			UnitID: unitID,
		})
		if err != nil {
			return err
		}
		for _, d := range res.Drivers {
			fmt.Println(protojson.Format(d))
		}
		return nil
	}
	return cmd
}

// --- Unit Groups ---

func newListUnitGroupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unit-groups list",
		Short:   "List unit groups",
		GroupID: "unit-groups",
	}
	unitID := cmd.Flags().Int64("unit-id", 0, "Filter by Unit ID")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		res, err := client.ListUnitGroups(cmd.Context(), &mapon.ListUnitGroupsRequest{
			UnitID: *unitID,
		})
		if err != nil {
			return err
		}
		for _, g := range res.Groups {
			fmt.Println(protojson.Format(g))
		}
		return nil
	}
	return cmd
}

func newListUnitsInGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unit-groups units",
		Short:   "List units in a group",
		GroupID: "unit-groups",
	}
	groupID := cmd.Flags().Int64("group-id", 0, "Group ID")
	_ = cmd.MarkFlagRequired("group-id")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		res, err := client.ListUnitsInGroup(cmd.Context(), &mapon.ListUnitsInGroupRequest{
			GroupID: *groupID,
		})
		if err != nil {
			return err
		}
		list := &maponv1.UnitIDsList{}
		list.SetIds(res.UnitIDs)
		fmt.Println(protojson.Format(list))
		return nil
	}
	return cmd
}

// --- Other ---

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
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	ids := cmd.Flags().StringSlice("unit-id", nil, "Filter by unit ID")
	include := cmd.Flags().StringSlice("include", nil, "Include additional data (polyline)")

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
		response, err := client.ListRoutes(cmd.Context(), &mapon.ListRoutesRequest{
			From:    *from,
			To:      *to,
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
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	ids := cmd.Flags().StringSlice("unit-id", nil, "Filter by unit ID")
	driver := cmd.Flags().Int64("driver-id", 0, "Filter by driver ID")

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
		response, err := client.ListAlerts(cmd.Context(), &mapon.ListAlertsRequest{
			From:    *from,
			To:      *to,
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

// Helpers

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
