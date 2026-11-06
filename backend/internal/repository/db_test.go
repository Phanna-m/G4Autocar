package repository

import (
	"reflect"
	"testing"
)

func TestMigrationFiles(t *testing.T) {
	tests := []struct {
		name         string
		setupApplied bool
		want         []string
	}{
		{
			name:         "Supabase setup skips baseline and seed scripts",
			setupApplied: true,
			want: []string{
				"000004_admin_modules.sql",
				"000005_seed_more_vehicles.sql",
				"000006_car_only_catalog.sql",
				"000007_checkout_delivery.sql",
			},
		},
		{
			name: "regular migrations run when Supabase setup was not applied",
			want: []string{
				"000001_init_schema.sql",
				"000002_seed_data.sql",
				"000003_seed_toyota_catalog.sql",
				"000004_admin_modules.sql",
				"000005_seed_more_vehicles.sql",
				"000006_car_only_catalog.sql",
				"000007_checkout_delivery.sql",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := migrationFiles(tt.setupApplied); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("migrationFiles(%t) = %v, want %v", tt.setupApplied, got, tt.want)
			}
		})
	}
}

func TestIsVehicleCategory(t *testing.T) {
	tests := []struct {
		categoryType string
		want         bool
	}{
		{categoryType: "GASOLINE_VEHICLE", want: true},
		{categoryType: "ELECTRIC_VEHICLE", want: true},
		{categoryType: "BICYCLE", want: false},
		{categoryType: "SPARE_PART", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.categoryType, func(t *testing.T) {
			if got := isVehicleCategory(tt.categoryType); got != tt.want {
				t.Fatalf("isVehicleCategory(%q) = %t, want %t", tt.categoryType, got, tt.want)
			}
		})
	}
}
