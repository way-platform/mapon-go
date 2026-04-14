package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	mapon "github.com/way-platform/mapon-go"
	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
	"golang.org/x/term"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// resolveCredentials returns credentials from the store.
func resolveCredentials(cfg *config) (*Credentials, error) {
	if cfg.credentialStore == nil {
		return nil, fmt.Errorf("no credential source configured")
	}
	return cfg.credentialStore.Load()
}

// NewCommand builds the full CLI command tree for the Mapon SDK.
func NewCommand(opts ...Option) *cobra.Command {
	cfg := config{}
	for _, opt := range opts {
		opt(&cfg)
	}
	cmd := &cobra.Command{
		Use:   "mapon",
		Short: "Mapon API CLI",
	}

	cmd.AddGroup(&cobra.Group{ID: "units", Title: "Units"})
	cmd.AddCommand(newListUnitsCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "unit-data", Title: "Unit Data"})
	cmd.AddCommand(newListIgnitionsCommand(&cfg))
	cmd.AddCommand(newListTemperaturesCommand(&cfg))
	cmd.AddCommand(newListDigitalInputsCommand(&cfg))
	cmd.AddCommand(newListDigitalInputsExtendedCommand(&cfg))
	cmd.AddCommand(newListIbuttonsCommand(&cfg))
	cmd.AddCommand(newListHumidityCommand(&cfg))
	cmd.AddCommand(newListCanPeriodDataCommand(&cfg))
	cmd.AddCommand(newGetCanPointDataCommand(&cfg))
	cmd.AddCommand(newGetHistoryPointDataCommand(&cfg))
	cmd.AddCommand(newGetUnitFieldsCommand(&cfg))
	cmd.AddCommand(newGetUnitDebugInfoCommand(&cfg))
	cmd.AddCommand(newGetDrivingTimeExtendedCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "unit-groups", Title: "Unit Groups"})
	cmd.AddCommand(newUnitGroupsCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "drivers", Title: "Drivers"})
	cmd.AddCommand(newListDriversCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "routes", Title: "Routes"})
	cmd.AddCommand(newListRoutesCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "objects", Title: "Objects"})
	cmd.AddCommand(newListObjectsCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "alerts", Title: "Alerts"})
	cmd.AddCommand(newListAlertsCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "data-forward", Title: "Data Forwarding"})
	cmd.AddCommand(newDataForwardCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "auth", Title: "Authentication"})
	cmd.AddCommand(newAuthCommand(&cfg))

	cmd.AddGroup(&cobra.Group{ID: "utils", Title: "Utils"})
	cmd.SetHelpCommandGroupID("utils")
	cmd.SetCompletionCommandGroupID("utils")
	return cmd
}

// --- Auth ---

func newAuthCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Short:   "Authenticate to the Mapon API",
		GroupID: "auth",
	}
	cmd.AddCommand(newLoginCommand(cfg))
	cmd.AddCommand(newLogoutCommand(cfg))
	return cmd
}

func newLoginCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the Mapon API",
	}
	apiKey := cmd.Flags().String("api-key", "", "API key for authentication")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		// Try loading stored credentials first.
		creds := &Credentials{}
		if cfg.credentialStore != nil {
			if loaded, err := cfg.credentialStore.Load(); err == nil {
				creds = loaded
			}
		}
		// Override with flag.
		if *apiKey != "" {
			creds.APIKey = *apiKey
		}
		// Prompt for missing API key.
		if creds.APIKey == "" {
			val, err := promptSecret(cmd, "Enter API key: ")
			if err != nil {
				return err
			}
			creds.APIKey = val
		}
		// Persist credentials.
		if cfg.credentialStore != nil {
			if err := cfg.credentialStore.Save(creds); err != nil {
				return fmt.Errorf("write credentials: %w", err)
			}
		}
		cmd.Println("Logged in to the Mapon API.")
		return nil
	}
	return cmd
}

func newLogoutCommand(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from the Mapon API",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cfg.credentialStore != nil {
				if err := cfg.credentialStore.Clear(); err != nil {
					return fmt.Errorf("clear credentials: %w", err)
				}
			}
			cmd.Println("Logged out.")
			return nil
		},
	}
}

// --- Client ---

func newClient(cmd *cobra.Command, cfg *config) (*mapon.Client, error) {
	creds, err := resolveCredentials(cfg)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("no credentials found, please login using `mapon auth login`")
		}
		return nil, err
	}
	var opts []mapon.ClientOption
	if cfg.httpClient != nil {
		opts = append(opts, mapon.WithHTTPClient(cfg.httpClient))
	}
	opts = append(opts, mapon.WithAPIKey(creds.APIKey))
	return mapon.NewClient(cmd.Context(), opts...)
}

// --- Units ---

func newListUnitsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "units",
		Short:   "List units",
		GroupID: "units",
	}
	ids := cmd.Flags().StringSlice("id", nil, "Filter by unit ID")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
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
		req := &maponv1.ListUnitsRequest{}
		req.SetUnitIds(unitIDs)
		response, err := client.ListUnits(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, unit := range response.GetUnits() {
			fmt.Println(protojson.Format(unit))
		}
		return nil
	}
	return cmd
}

// --- Unit Data ---

func newListIgnitionsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ignitions <unit-id ...>",
		Short:   "List ignition events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.ListIgnitionsRequest{}
		req.SetUnitIds(unitIDs)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		res, err := client.ListIgnitions(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListTemperaturesCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "temperatures <unit-id ...>",
		Short:   "List temperature data",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.ListTemperaturesRequest{}
		req.SetUnitIds(unitIDs)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		res, err := client.ListTemperatures(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListDigitalInputsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "digital-inputs <unit-id ...>",
		Short:   "List digital input events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.ListDigitalInputsRequest{}
		req.SetUnitIds(unitIDs)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		res, err := client.ListDigitalInputs(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListDigitalInputsExtendedCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "digital-inputs-extended <unit-id ...>",
		Short:   "List extended digital input events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.ListDigitalInputsExtendedRequest{}
		req.SetUnitIds(unitIDs)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		res, err := client.ListDigitalInputsExtended(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListIbuttonsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ibuttons <unit-id ...>",
		Short:   "List iButton events",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.ListIbuttonsRequest{}
		req.SetUnitIds(unitIDs)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		res, err := client.ListIbuttons(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListHumidityCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "humidity <unit-id ...>",
		Short:   "List humidity data",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.ListHumidityRequest{}
		req.SetUnitIds(unitIDs)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		res, err := client.ListHumidity(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newListCanPeriodDataCommand(cfg *config) *cobra.Command {
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
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		req := &maponv1.ListCanPeriodDataRequest{}
		req.SetUnitId(unitID)
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		req.SetInclude(*include)
		res, err := client.ListCanPeriodData(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetCanPointDataCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "can-point <unit-id>",
		Short:   "Get CAN data at a specific time",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	datetime := cmd.Flags().Time("datetime", time.Now(), []string{time.DateOnly, time.RFC3339}, "Datetime")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		req := &maponv1.GetCanDataPointRequest{}
		req.SetUnitId(unitID)
		req.SetDatetime(timestamppb.New(*datetime))
		res, err := client.GetCanDataPoint(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetHistoryPointDataCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "history-point <unit-id>",
		Short:   "Get historical data at a specific time",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	datetime := cmd.Flags().Time("datetime", time.Now(), []string{time.DateOnly, time.RFC3339}, "Datetime")
	include := cmd.Flags().StringSlice("include", nil, "Fields to include")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		req := &maponv1.GetHistoryPointDataRequest{}
		req.SetUnitId(unitID)
		req.SetDatetime(timestamppb.New(*datetime))
		req.SetInclude(*include)
		res, err := client.GetHistoryPointData(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetUnitFieldsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fields <unit-id>",
		Short:   "Get unit custom fields",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		req := &maponv1.GetUnitFieldsRequest{}
		req.SetUnitId(unitID)
		res, err := client.GetUnitFields(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetUnitDebugInfoCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug-info <unit-id ...>",
		Short:   "Get unit debug info",
		GroupID: "unit-data",
		Args:    cobra.MinimumNArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitIDs, err := parseUnitIDs(args)
		if err != nil {
			return err
		}
		req := &maponv1.GetUnitDebugInfoRequest{}
		req.SetUnitIds(unitIDs)
		res, err := client.GetUnitDebugInfo(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, u := range res.GetUnits() {
			fmt.Println(protojson.Format(u))
		}
		return nil
	}
	return cmd
}

func newGetDrivingTimeExtendedCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "driving-time <unit-id>",
		Short:   "Get driving time extended",
		GroupID: "unit-data",
		Args:    cobra.ExactArgs(1),
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		unitID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unit ID %s: %w", args[0], err)
		}
		req := &maponv1.GetDrivingTimeExtendedRequest{}
		req.SetUnitId(unitID)
		res, err := client.GetDrivingTimeExtended(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, d := range res.GetDrivers() {
			fmt.Println(protojson.Format(d))
		}
		return nil
	}
	return cmd
}

// --- Unit Groups ---

func newUnitGroupsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unit-groups",
		Short:   "Unit group commands",
		GroupID: "unit-groups",
	}
	cmd.AddCommand(newListUnitGroupsCommand(cfg))
	cmd.AddCommand(newListUnitsInGroupCommand(cfg))
	return cmd
}

func newListUnitGroupsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List unit groups",
	}
	unitID := cmd.Flags().Int64("unit-id", 0, "Filter by Unit ID")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		req := &maponv1.ListUnitGroupsRequest{}
		req.SetUnitId(*unitID)
		res, err := client.ListUnitGroups(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, g := range res.GetGroups() {
			fmt.Println(protojson.Format(g))
		}
		return nil
	}
	return cmd
}

func newListUnitsInGroupCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "units",
		Short: "List units in a group",
	}
	groupID := cmd.Flags().Int64("group-id", 0, "Group ID")
	_ = cmd.MarkFlagRequired("group-id")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		req := &maponv1.ListUnitsInGroupRequest{}
		req.SetGroupId(*groupID)
		res, err := client.ListUnitsInGroup(cmd.Context(), req)
		if err != nil {
			return err
		}
		list := &maponv1.UnitIDsList{}
		list.SetIds(res.GetUnitIds())
		fmt.Println(protojson.Format(list))
		return nil
	}
	return cmd
}

// --- Drivers ---

func newListDriversCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "drivers",
		Short:   "List drivers",
		GroupID: "drivers",
	}
	id := cmd.Flags().Int64("id", 0, "Filter by driver ID")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		req := &maponv1.ListDriversRequest{}
		req.SetId(*id)
		response, err := client.ListDrivers(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, driver := range response.GetDrivers() {
			fmt.Println(protojson.Format(driver))
		}
		return nil
	}
	return cmd
}

// --- Routes ---

func newListRoutesCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "routes",
		Short:   "List routes",
		GroupID: "routes",
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	ids := cmd.Flags().StringSlice("unit-id", nil, "Filter by unit ID")
	include := cmd.Flags().StringSlice("include", nil, "Include additional data (polyline)")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
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
		req := &maponv1.ListRoutesRequest{}
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		req.SetUnitIds(unitIDs)
		req.SetInclude(*include)
		response, err := client.ListRoutes(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, route := range response.GetRoutes() {
			fmt.Println(protojson.Format(route))
		}
		return nil
	}
	return cmd
}

// --- Objects ---

func newListObjectsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "objects",
		Short:   "List objects",
		GroupID: "objects",
	}
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		response, err := client.ListObjects(cmd.Context(), &maponv1.ListObjectsRequest{})
		if err != nil {
			return err
		}
		for _, object := range response.GetObjects() {
			fmt.Println(protojson.Format(object))
		}
		return nil
	}
	return cmd
}

// --- Alerts ---

func newListAlertsCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alerts",
		Short:   "List alerts",
		GroupID: "alerts",
	}
	from := cmd.Flags().Time("from", time.Now().Add(-time.Hour*24), []string{time.DateOnly, time.RFC3339}, "From time")
	to := cmd.Flags().Time("to", time.Now(), []string{time.DateOnly, time.RFC3339}, "To time")
	ids := cmd.Flags().StringSlice("unit-id", nil, "Filter by unit ID")
	driver := cmd.Flags().Int64("driver-id", 0, "Filter by driver ID")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
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
		req := &maponv1.ListAlertsRequest{}
		req.SetFromTime(timestamppb.New(*from))
		req.SetToTime(timestamppb.New(*to))
		req.SetUnitIds(unitIDs)
		req.SetDriver(*driver)
		response, err := client.ListAlerts(cmd.Context(), req)
		if err != nil {
			return err
		}
		for _, alert := range response.GetAlerts() {
			fmt.Println(protojson.Format(alert))
		}
		return nil
	}
	return cmd
}

// --- Helpers ---

func parseUnitIDs(args []string) ([]int64, error) {
	var unitIDs []int64
	for _, idStr := range args {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unit ID %s: %w", idStr, err)
		}
		unitIDs = append(unitIDs, id)
	}
	return unitIDs, nil
}

func promptSecret(cmd *cobra.Command, prompt string) (string, error) {
	cmd.Print(prompt)
	input, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	cmd.Println()
	return string(input), nil
}

// --- Data Forward ---

func newDataForwardCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "data-forward",
		Short:   "Manage data forwarding endpoints",
		GroupID: "data-forward",
	}
	cmd.AddCommand(newListDataForwardsCommand(cfg))
	cmd.AddCommand(newSaveDataForwardCommand(cfg))
	cmd.AddCommand(newDeleteDataForwardCommand(cfg))
	return cmd
}

func newListDataForwardsCommand(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List data forwarding endpoints",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient(cmd, cfg)
			if err != nil {
				return err
			}
			resp, err := client.ListDataForwards(cmd.Context(), maponv1.ListDataForwardsRequest_builder{}.Build())
			if err != nil {
				return err
			}
			for _, ep := range resp.GetEndpoints() {
				fmt.Printf("id=%d url=%s packs=%v\n", ep.GetId(), ep.GetUrl(), ep.GetPacks())
			}
			return nil
		},
	}
}

func newSaveDataForwardCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Register a data forwarding endpoint",
	}
	id := cmd.Flags().Int64("id", 0, "Existing endpoint ID to update (omit to create new)")
	webhookURL := cmd.Flags().String("url", "", "Webhook URL to receive data")
	_ = cmd.MarkFlagRequired("url")
	packs := cmd.Flags().Int32Slice("pack", nil, "Pack ID to forward (repeatable, e.g. --pack 1 --pack 3)")
	_ = cmd.MarkFlagRequired("pack")
	unitIDs := cmd.Flags().Int64Slice("unit-id", nil, "Unit IDs to forward (omit for all units)")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		resp, err := client.SaveDataForward(cmd.Context(),
			maponv1.SaveDataForwardRequest_builder{
				Id:      new(*id),
				Url:     new(*webhookURL),
				Packs:   *packs,
				UnitIds: *unitIDs,
			}.Build())
		if err != nil {
			return err
		}
		fmt.Printf("registered endpoint id=%d\n", resp.GetEndpointId())
		return nil
	}
	return cmd
}

func newDeleteDataForwardCommand(cfg *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a data forwarding endpoint",
	}
	id := cmd.Flags().Int64("id", 0, "Endpoint ID to delete")
	_ = cmd.MarkFlagRequired("id")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		client, err := newClient(cmd, cfg)
		if err != nil {
			return err
		}
		if _, err := client.DeleteDataForward(cmd.Context(),
			maponv1.DeleteDataForwardRequest_builder{
				EndpointId: new(*id),
			}.Build()); err != nil {
			return err
		}
		fmt.Printf("deleted endpoint id=%d\n", *id)
		return nil
	}
	return cmd
}
